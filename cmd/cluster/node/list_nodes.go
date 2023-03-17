package node

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var listNodeCmd = &cobra.Command{
	Use:   "list",
	Short: "List nodes for a cluster",
	Long:  "List nodes for a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf("Could not initiate api client: %s", err.Error())
		}
		authApi.GetInfo("", "")

		clusterName, _ := cmd.Flags().GetString("cluster-name")
		clusterId, err := authApi.GetClusterIdByName(clusterName)
		if err != nil {
			logrus.Fatalf("%s", ybmAuthClient.GetApiErrorDetails(err))
		}

		resp, r, err := authApi.GetClusterNode(clusterId).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		if len(resp.GetData()) == 0 {
			logrus.Fatalf("No nodes found")
		}

		nodesCtx := formatter.Context{
			Output: os.Stdout,
			Format: formatter.NewNodeFormat(viper.GetString("output")),
		}
		formatter.NodeWrite(nodesCtx, resp.GetData())

	},
}

func init() {
	NodeCmd.AddCommand(listNodeCmd)
	listNodeCmd.Flags().String("cluster-name", "", "[REQUIRED] The name of the cluster to get details.")
	listNodeCmd.MarkFlagRequired("cluster-name")
}
