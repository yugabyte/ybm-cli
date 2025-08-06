// Licensed to Yugabyte, Inc. under one or more contributor license
// agreements. See the NOTICE file distributed with this work for
// additional information regarding copyright ownership. Yugabyte
// licenses this file to you under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cluster

import (
	"fmt"
	"os"
	"strings"

	"encoding/base64"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	encryption "github.com/yugabyte/ybm-cli/cmd/cluster/encryption"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// createClusterCmd represents the cluster command
var createClusterCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cluster",
	Long:  "Create a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		credentials, _ := cmd.Flags().GetStringToString("credentials")

		username := credentials["username"]
		password := credentials["password"]
		regionInfoMapList := []map[string]string{}
		changedRegionInfo := cmd.Flags().Changed("region-info")

		if changedRegionInfo {
			regionInfoList, _ := cmd.Flags().GetStringArray("region-info")
			for _, regionInfoString := range regionInfoList {
				regionInfoMap := map[string]string{}
				for _, regionInfo := range strings.Split(regionInfoString, ",") {
					kvp := strings.Split(regionInfo, "=")
					if len(kvp) != 2 {
						logrus.Fatalln("Incorrect format in region info")
					}
					key := kvp[0]
					val := kvp[1]
					switch key {
					case "region":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["region"] = val
						}
					case "num-nodes":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["num-nodes"] = val
						}
					case "vpc":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["vpc"] = val
						}
					case "num-cores":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["num-cores"] = val
						}
					case "disk-size-gb":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["disk-size-gb"] = val
						}
					case "disk-iops":
						if len(strings.TrimSpace(val)) != 0 {
							regionInfoMap["disk-iops"] = val
						}
					}
				}

				if _, ok := regionInfoMap["region"]; !ok {
					logrus.Fatalln("Region not specified in region info")
				}
				if _, ok := regionInfoMap["num-nodes"]; !ok {
					logrus.Fatalln("Number of nodes not specified in region info")
				}
				if _, ok := regionInfoMap["num-cores"]; !ok {
					logrus.Fatalln("Number of cores not specified in region info")
				}
				if _, ok := regionInfoMap["disk-size-gb"]; !ok {
					logrus.Fatalln("Disk size not specified in region info")
				}

				regionInfoMapList = append(regionInfoMapList, regionInfoMap)
			}
		}

		cmkSpec, err := encryption.GetCmkSpecFromCommand(cmd)
		if err != nil {
			logrus.Fatalf("Error while getting CMK spec: %s", err)
		}

		clusterSpec, err := authApi.CreateClusterSpec(cmd, regionInfoMapList)
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		dbCredentials := ybmclient.NewCreateClusterRequestEncryptedDbCredentialsWithDefaults()
		dbCredentials.Ycql = *ybmclient.NewEncryptedDBCredentials(encodeBase64(username), encodeBase64(password))
		dbCredentials.Ysql = *ybmclient.NewEncryptedDBCredentials(encodeBase64(username), encodeBase64(password))

		createClusterRequest := ybmclient.NewCreateClusterRequest(*clusterSpec)
		createClusterRequest.SetEncryptedDbCredentials(*dbCredentials)

		// Enable connection pooling feature if requested
		enableConnectionPooling, _ := cmd.Flags().GetBool("enable-connection-pooling")
		if enableConnectionPooling {
			if !util.IsFeatureFlagEnabled(util.CONNECTION_POOLING) {
				logrus.Fatalf("Connection pooling feature is not enabled yet — it will be available soon.")
			}
			// Set connection pooling in features array for cluster creation
			features := []ybmclient.CreateClusterFeatureEnum{ybmclient.CREATECLUSTERFEATUREENUM_ENABLE_CONNECTION_POOLING}
			createClusterRequest.SetFeatures(features)
			logrus.Debugf("Features array set to: %v", features)
			logrus.Info("Connection pooling will be enabled during cluster creation")
		}

		if cmkSpec != nil {
			logrus.Debug("Setting up CMK spec for cluster creation")
			createClusterRequest.SecurityCmkSpec = *ybmclient.NewNullableCMKSpec(cmkSpec)
		}

		// Debug log the complete request payload
		logrus.Debugf("Create Cluster Request Payload: %+v", *createClusterRequest)
		if createClusterRequest.Features != nil {
			logrus.Debugf("Features in final request: %v", *createClusterRequest.Features)
		}

		resp, r, err := authApi.CreateCluster().CreateClusterRequest(*createClusterRequest).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		clusterID := resp.GetData().Info.Id
		clusterData := []ybmclient.ClusterData{resp.GetData()}

		msg := fmt.Sprintf("The cluster %s is being created", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, ybmclient.TASKTYPEENUM_CREATE_CLUSTER, []string{"FAILED", "SUCCEEDED"}, msg)
			if err != nil {
				logrus.Fatalf("error when getting task status: %s", err)
			}
			if returnStatus != "SUCCEEDED" {
				logrus.Fatalf("Operation failed with error: %s", returnStatus)
			}
			fmt.Printf("The cluster %s has been created\n", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

			respC, r, err := authApi.ListClusters().Name(clusterName).Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", r)
				logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
			}
			clusterData = respC.GetData()
		} else {
			fmt.Println(msg)

			// Inform user about connection pooling if requested
			if enableConnectionPooling {
				fmt.Printf("Note: Connection pooling has been enabled for this cluster.\n")
			}
		}

		clustersCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewClusterFormat(viper.GetString("output")),
		}

		formatter.ClusterWrite(clustersCtx, clusterData)
	},
}

func init() {
	ClusterCmd.AddCommand(createClusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createClusterCmd.Flags().SortFlags = false
	createClusterCmd.Flags().String("cluster-name", "", "[REQUIRED] Name of the cluster.")
	createClusterCmd.MarkFlagRequired("cluster-name")
	createClusterCmd.Flags().StringToString("credentials", nil, `[REQUIRED] Credentials to login to the cluster. Please provide key value pairs username=<user-name>,password=<password>.`)
	createClusterCmd.MarkFlagRequired("credentials")
	createClusterCmd.Flags().String("cloud-provider", "", "[OPTIONAL] The cloud provider where database needs to be deployed. AWS, AZURE or GCP. Default AWS.")
	createClusterCmd.Flags().String("cluster-tier", "", "[OPTIONAL] The tier of the cluster. Sandbox or Dedicated. Default Sandbox.")
	createClusterCmd.Flags().String("cluster-type", "", "[OPTIONAL] Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED. Default SYNCHRONOUS.")
	createClusterCmd.Flags().String("database-version", "", "[OPTIONAL] The database version of the cluster. Production, Innovation, Preview, or 'Early Access'. Default depends on cluster tier, Sandbox is Preview, Dedicated is Production.")
	if util.IsFeatureFlagEnabled(util.ENTERPRISE_SECURITY) {
		createClusterCmd.Flags().Bool("enterprise-security", false, "[OPTIONAL] The security level of cluster. Advanced security will have security checks for cluster. Default false.")
	}
	createClusterCmd.Flags().String("encryption-spec", "", `[OPTIONAL] The customer managed key spec for the cluster.
	Please provide key value pairs as follows:
	For AWS: 
	cloud-provider=AWS,aws-secret-key=<secret-key>,aws-access-key=<access-key>,aws-arn=<arn1>,aws-arn=<arn2> .
	aws-access-key can be ommitted if the environment variable YBM_AWS_SECRET_KEY is set. If the environment variable is not set, the user will be prompted to enter the value.
	For GCP:
	cloud-provider=GCP,gcp-resource-id=<resource-id>,gcp-service-account-path=<service-account-path>.
	For AZURE:
	cloud-provider=AZURE,azu-client-id=<client-id>,azu-client-secret=<client-secret>,azu-tenant-id=<tenant-id>,azu-key-name=<key-name>,azu-key-vault-uri=<key-vault-uri>.
	If specified, all parameters for that provider are mandatory.`)
	createClusterCmd.Flags().String("fault-tolerance", "", "[OPTIONAL] Fault tolerance of the cluster. The possible values are NONE, NODE, ZONE, or REGION. Default NONE.")
	createClusterCmd.Flags().Int32("num-faults-to-tolerate", 0, "[OPTIONAL] The number of domain faults to tolerate for the level specified. The possible values are 0 for NONE, 1 for ZONE and [1-3] for anything else. Defaults to 0 for NONE, 1 otherwise.")
	createClusterCmd.Flags().StringArray("region-info", []string{}, `Region information for the cluster, provided as key-value pairs. Arguments are region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name>,num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,disk-iops=<disk-iops> (AWS only). region, num-nodes, num-cores, disk-size-gb are required. Specify one --region-info flag for each region in the cluster.`)
	createClusterCmd.MarkFlagRequired("region-info")
	createClusterCmd.Flags().String("preferred-region", "", "[OPTIONAL] The preferred region in a multi region cluster. The preferred region handles all read and write requests from clients.")
	createClusterCmd.Flags().String("default-region", "", "[OPTIONAL] The primary region in a partition-by-region cluster. The primary region is where all the tables not created in a tablespace reside.")
	createClusterCmd.Flags().Bool("enable-connection-pooling", false, "[OPTIONAL] Enable connection pooling for the cluster after creation. Default false.")
}
