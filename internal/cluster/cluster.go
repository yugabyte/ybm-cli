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
		if err.Error() == "409 Conflict" {
			logrus.Debugf("Failed to get allow lists because cluster %s is not ready yet", f.Cluster.Info.Id)
		} else {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf("Error when calling `ClusterApi.ListClusterNetworkAllowLists`: %s", ybmAuthClient.GetApiErrorDetails(err))
		}
	}
	if _, ok := resp.GetDataOk(); ok {
		f.AllowList = resp.GetData()
	}
}
