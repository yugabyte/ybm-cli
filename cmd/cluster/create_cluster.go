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
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	encryption "github.com/yugabyte/ybm-cli/cmd/cluster/encryption"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

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
		changedNodeInfo := cmd.Flags().Changed("node-config")

		defaultNumCores := 0
		defaultDiskSizeGb := 0
		defaultDiskIops := 0
		if changedNodeInfo {
			nodeConfig, _ := cmd.Flags().GetStringToInt("node-config")
			numCores, ok := nodeConfig["num-cores"]

			if ok {
				defaultNumCores = numCores
			}
			if diskSizeGb, ok := nodeConfig["disk-size-gb"]; ok {
				defaultDiskSizeGb = diskSizeGb
			}
			if diskIops, ok := nodeConfig["disk-iops"]; ok {
				defaultDiskIops = diskIops
			}
		}

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
				if _, ok := regionInfoMap["num-cores"]; !ok && defaultNumCores > 0 {
					regionInfoMap["num-cores"] = strconv.Itoa(defaultNumCores)
				}
				if _, ok := regionInfoMap["disk-size-gb"]; !ok && defaultDiskSizeGb > 0 {
					regionInfoMap["disk-size-gb"] = strconv.Itoa(defaultDiskSizeGb)
				}
				if _, ok := regionInfoMap["disk-iops"]; !ok && defaultDiskIops > 0 {
					regionInfoMap["disk-iops"] = strconv.Itoa(defaultDiskIops)
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

		dbCredentials := ybmclient.NewCreateClusterRequestDbCredentialsWithDefaults()
		dbCredentials.Ycql = *ybmclient.NewDBCredentials(username, password)
		dbCredentials.Ysql = *ybmclient.NewDBCredentials(username, password)

		createClusterRequest := ybmclient.NewCreateClusterRequest(*clusterSpec, *dbCredentials)

		if cmkSpec != nil {
			logrus.Debug("Setting up CMK spec for cluster creation")
			createClusterRequest.SecurityCmkSpec = *ybmclient.NewNullableCMKSpec(cmkSpec)
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
	createClusterCmd.Flags().String("database-version", "", "[OPTIONAL] The database version of the cluster. Production, Innovation, Preview or Early Access. Default depends on cluster tier, Sandbox is Preview, Dedicated is Production.")
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
	createClusterCmd.Flags().StringToInt("node-config", nil, "[OPTIONAL] Number of vCPUs and disk size per node for the cluster, provided as key-value pairs. Arguments are num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,disk-iops=<disk-iops> (AWS only). num-cores is required.")
	createClusterCmd.Flags().MarkDeprecated("node-config", "use --region-info to specify num-cores, disk-size-gb, and disk-iops")
	createClusterCmd.Flags().StringArray("region-info", []string{}, `Region information for the cluster, provided as key-value pairs. Arguments are region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name>,num-cores=<num-cores>,disk-size-gb=<disk-size-gb>,disk-iops=<disk-iops> (AWS only). region, num-nodes, num-cores, disk-size-gb are required. Specify one --region-info flag for each region in the cluster.`)
	createClusterCmd.Flags().String("preferred-region", "", "[OPTIONAL] The preferred region in a multi region cluster. The preferred region handles all read and write requests from clients.")
	createClusterCmd.Flags().String("default-region", "", "[OPTIONAL] The primary region in a partition-by-region cluster. The primary region is where all the tables not created in a tablespace reside.")
}
