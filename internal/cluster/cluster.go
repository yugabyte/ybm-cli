package cluster

import (
	"github.com/sirupsen/logrus"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

type FullCluster struct {
	authApi ybmAuthClient.AuthApiClient
	Cluster ybmclient.ClusterData
}

func NewFullCluster(authApi ybmAuthClient.AuthApiClient, clusterId string) *FullCluster {
	resp, r, err := authApi.GetCluster(clusterId).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		logrus.Fatalf("Error when calling `ClusterApi.GetCluster`: %s", ybmAuthClient.GetApiErrorDetails(err))
	}

	return &FullCluster{
		authApi: authApi,
		Cluster: resp.GetData(),
	}
}

func (f *FullCluster) GetVPCs() {
	if f.Cluster.Spec.h

}
