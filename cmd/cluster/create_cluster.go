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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			logrus.Fatalf("could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		credentials, _ := cmd.Flags().GetStringToString("credentials")

		username := credentials["username"]
		password := credentials["password"]
		regionInfoMapList := []map[string]string{}
		if cmd.Flags().Changed("region-info") {
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
					}
				}

				if _, ok := regionInfoMap["region"]; !ok {
					logrus.Fatalln("Region not specified in region info")
				}
				if _, ok := regionInfoMap["num-nodes"]; !ok {
					logrus.Fatalln("Number of nodes not specified in region info")
				}

				regionInfoMapList = append(regionInfoMapList, regionInfoMap)
			}
		}

		if cmd.Flags().Changed("node-config") {
			nodeConfig, _ := cmd.Flags().GetStringToInt("node-config")
			if _, ok := nodeConfig["num-cores"]; !ok {
				logrus.Fatalln("Number of cores not specified in node config")
			}
		}

		var cmkSpec *ybmclient.CMKSpec = nil
		if cmd.Flags().Changed("cmk-spec") {
			cmkString, _ := cmd.Flags().GetString("cmk-spec")
			cmkProvider := ""
			cmkAwsSecretKey := ""
			cmkAwsAccessKey := ""
			cmkAwsArnList := []string{}

			for _, cmkInfo := range strings.Split(cmkString, ",") {
				kvp := strings.Split(cmkInfo, "=")
				if len(kvp) != 2 {
					logrus.Fatalln("Incorrect format in cmk spec")
				}
				key := kvp[0]
				val := kvp[1]
				switch key {
				case "provider":
					if len(strings.TrimSpace(val)) != 0 {
						cmkProvider = val
					}
				case "aws-secret-key":
					if len(strings.TrimSpace(val)) != 0 {
						cmkAwsSecretKey = val
					}
				case "aws-access-key":
					if len(strings.TrimSpace(val)) != 0 {
						cmkAwsAccessKey = val
					}
				case "aws-arn":
					if len(strings.TrimSpace(val)) != 0 {
						cmkAwsArnList = append(cmkAwsArnList, val)
					}
				}
			}

			if cmkProvider == "" ||
				(cmkProvider == "AWS" && (cmkAwsSecretKey == "" || cmkAwsAccessKey == "" || len(cmkAwsArnList) == 0)) {
				logrus.Fatalln("Incorrect format in cmk spec")
			}

			cmkSpec = ybmclient.NewCMKSpec(cmkProvider)
			if cmkProvider == "AWS" {
				cmkSpec.AwsCmkSpec = ybmclient.NewAWSCMKSpec(cmkAwsSecretKey, cmkAwsAccessKey, cmkAwsArnList)
			}
		}

		clusterSpec, err := authApi.CreateClusterSpec(cmd, regionInfoMapList)
		if err != nil {
			logrus.Fatalf("Error while creating cluster spec: %v", err)
		}

		dbCredentials := ybmclient.NewCreateClusterRequestDbCredentials()
		dbCredentials.Ycql = ybmclient.NewDBCredentials(username, password)
		dbCredentials.Ysql = ybmclient.NewDBCredentials(username, password)

		createClusterRequest := ybmclient.NewCreateClusterRequest(*clusterSpec, *dbCredentials)

		if cmkSpec != nil {
			logrus.Debug("Setting up CMK spec for cluster creation")
			createClusterRequest.SecurityCmkSpec = cmkSpec
		}

		resp, r, err := authApi.CreateCluster().CreateClusterRequest(*createClusterRequest).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ClusterApi.CreateCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}

		clusterID := resp.GetData().Info.Id
		clusterData := []ybmclient.ClusterData{resp.GetData()}

		msg := fmt.Sprintf("The cluster %s is being created", formatter.Colorize(clusterName, formatter.GREEN_COLOR))

		if viper.GetBool("wait") {
			returnStatus, err := authApi.WaitForTaskCompletion(clusterID, ybmclient.ENTITYTYPEENUM_CLUSTER, "CREATE_CLUSTER", []string{"FAILED", "SUCCEEDED"}, msg, 1800)
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
				logrus.Fatalf("Error when calling `ClusterApi.ListClusters`: %s", ybmAuthClient.GetApiErrorDetails(err))
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
	createClusterCmd.Flags().String("cloud-type", "", "[OPTIONAL] The cloud provider where database needs to be deployed. AWS or GCP.")
	createClusterCmd.Flags().String("cluster-type", "", "[OPTIONAL] Cluster replication type. SYNCHRONOUS or GEO_PARTITIONED.")
	createClusterCmd.Flags().StringToInt("node-config", nil, "[OPTIONAL] Configuration of the cluster nodes. Please provide key value pairs num-cores=<num-cores>,disk-size-gb=<disk-size-gb> as the value. If specified  num-cores is mandatory, disk-size-gb is optional.")
	createClusterCmd.Flags().StringArray("region-info", []string{}, `[OPTIONAL] Region information for the cluster. Please provide key value pairs region=<region-name>,num-nodes=<number-of-nodes>,vpc=<vpc-name> as the value. If specified, region and num-nodes are mandatory, vpc is optional. Information about multiple regions can be specified.`)
	createClusterCmd.Flags().String("cluster-tier", "", "[OPTIONAL] The tier of the cluster. Sandbox or Dedicated.")
	createClusterCmd.Flags().String("fault-tolerance", "", "[OPTIONAL] The fault tolerance of the cluster. The possible values are NONE, ZONE and REGION.")
	createClusterCmd.Flags().String("database-version", "", "[OPTIONAL] The database version of the cluster. Stable or Preview.")
	createClusterCmd.Flags().String("cmk-spec", "", "[OPTIONAL] The customer managed key spec for the cluster. Please provide key value pairs provider=AWS,aws-secret-key=<secret-key>,aws-access-key=<access-key>,aws-arn=<arn1>,aws-arn=<arn2> . If specified, all parameters for that provider are mandatory.")

}
