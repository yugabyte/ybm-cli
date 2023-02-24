package cluster

import (
	"github.com/sirupsen/logrus"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// This struct is an attempt to consilidate Cluster information
// VPC, NetworkAllowList etc..
type FullCluster struct {
	Cluster ybmclient.ClusterData
	//VPC details "vpcid" => details
	Vpc map[string]ybmclient.SingleTenantVpcDataResponse
	//AllowList Attach to the cluster
	AllowList []ybmclient.NetworkAllowListData
}

func NewFullCluster(authApi ybmAuthClient.AuthApiClient, clusterData ybmclient.ClusterData) *FullCluster {
	fc := &FullCluster{
		Cluster: clusterData,
		Vpc:     map[string]ybmclient.SingleTenantVpcDataResponse{},
	}
	// Add VPC information
	fc.SetVPCs(authApi)
	fc.SetAllowLists(authApi)
	return fc
}

func (f *FullCluster) SetVPCs(authApi ybmAuthClient.AuthApiClient) {
	var VpcIds []string
	if _, ok := f.Cluster.GetSpecOk(); ok {
		for _, v := range f.Cluster.Spec.ClusterRegionInfo {
			if v.PlacementInfo.VpcId.IsSet() {
				if len(v.PlacementInfo.GetVpcId()) > 0 {
					VpcIds = append(VpcIds, v.PlacementInfo.GetVpcId())

				}
			}
		}
	}
	if len(VpcIds) > 0 {
		resp, r, err := authApi.ListSingleTenantVpcs().Ids(VpcIds).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `NetworkApi.ListSingleTenantVpcs`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
		if _, ok := resp.GetDataOk(); ok {
			for _, v := range resp.GetData() {
				f.Vpc[v.Info.Id] = v
			}
		}
	}
}

func (f *FullCluster) SetAllowLists(authApi ybmAuthClient.AuthApiClient) {
	resp, r, err := authApi.ListClusterNetworkAllowLists(f.Cluster.Info.Id).Execute()
	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		logrus.Fatalf("Error when calling `ClusterApi.ListClusterNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
	}
	if _, ok := resp.GetDataOk(); ok {
		f.AllowList = resp.GetData()
	}
}
