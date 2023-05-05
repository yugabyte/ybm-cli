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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yugabyte/ybm-cli/cmd/util"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	"golang.org/x/exp/slices"
)

// AuthApiClient is a auth YBM Client

var cliVersion = "v0.1.0"

type AuthApiClient struct {
	ApiClient *ybmclient.APIClient
	AccountID string
	ProjectID string
	ctx       context.Context
}

func SetVersion(version string) {
	cliVersion = version
}

func GetVersion() string {
	return cliVersion
}

// NewAuthClient function is returning a new AuthApiClient Client
func NewAuthApiClient() (*AuthApiClient, error) {
	url, err := ParseURL(viper.GetString("host"))
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	apiKey := viper.GetString("apiKey")
	// If the api key is empty, then tell the user to run the auth command.
	if len(apiKey) == 0 {
		logrus.Fatalln("No valid API key detected. Please run `ybm auth` to authenticate with YugabyteDB Managed.")
	}

	return NewAuthApiClientCustomUrlKey(url, apiKey)
}

func NewAuthApiClientCustomUrlKey(url *url.URL, apiKey string) (*AuthApiClient, error) {
	configuration := ybmclient.NewConfiguration()
	//Configure the client

	configuration.Host = url.Host
	configuration.Scheme = url.Scheme
	apiClient := ybmclient.NewAPIClient(configuration)

	apiClient.GetConfig().AddDefaultHeader("Authorization", "Bearer "+apiKey)
	apiClient.GetConfig().UserAgent = "ybm-cli/" + cliVersion
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	return &AuthApiClient{
		apiClient,
		"",
		"",
		ctx,
	}, nil
}

func (a *AuthApiClient) ListAccounts() ybmclient.ApiListAccountsRequest {
	return a.ApiClient.AccountApi.ListAccounts(a.ctx)
}

func (a *AuthApiClient) ListProjects() ybmclient.ApiListProjectsRequest {
	return a.ApiClient.ProjectApi.ListProjects(a.ctx, a.AccountID)
}

func (a *AuthApiClient) Ping() ybmclient.ApiGetPingRequest {
	return a.ApiClient.HealthCheckApi.GetPing(a.ctx)
}

func (a *AuthApiClient) GetAccountID(accountID string) (string, error) {
	//If an account ID is provided then we use this one
	if len(accountID) > 0 {
		return accountID, nil
	}
	accountResp, resp, err := a.ListAccounts().Execute()
	if err != nil {
		errMsg := getErrorMessage(resp, err)
		if strings.Contains(err.Error(), "is not a valid") {
			logrus.Debugln("The deserialization of the response failed due to following error. "+
				"Skipping as this should not impact the functionality of the provider.",
				map[string]interface{}{"errMsg": err.Error()})
		} else {
			return "", fmt.Errorf(errMsg)
		}
	}
	accountData := accountResp.GetData()
	if len(accountData) == 0 {
		return "", fmt.Errorf("the user is not associated with any accounts")
	}
	if len(accountData) > 1 {
		return "", fmt.Errorf("the user is associated with multiple accounts, please provide an account ID")
	}
	return accountData[0].Info.Id, nil
}

func (a *AuthApiClient) GetProjectID(projectID string) (string, error) {
	// If a projectID is specified then we use this one.
	if len(projectID) > 0 {
		return projectID, nil
	}

	projectResp, resp, err := a.ListProjects().Execute()
	if err != nil {
		errMsg := getErrorMessage(resp, err)
		if strings.Contains(err.Error(), "is not a valid") {
			logrus.Debugln("The deserialization of the response failed due to following error. "+
				"Skipping as this should not impact the functionality of the provider.",
				map[string]interface{}{"errMsg": err.Error()})
		} else {
			return "", fmt.Errorf(errMsg)
		}
	}
	projectData := projectResp.GetData()
	if len(projectData) == 0 {
		return "", fmt.Errorf("the account is not associated with any projects")
	}
	if len(projectData) > 1 {
		return "", fmt.Errorf("the account is associated with multiple projects, please provide a project id")
	}

	return projectData[0].Id, nil
}

func (a *AuthApiClient) CreateClusterSpec(cmd *cobra.Command, regionInfoList []map[string]string) (*ybmclient.ClusterSpec, error) {

	var diskSizeGb int32
	var memoryMb int32
	var trackId string
	var trackName string
	var regionInfoProvided bool
	var err error

	clusterRegionInfo := []ybmclient.ClusterRegionInfo{}
	totalNodes := 0
	for _, regionInfo := range regionInfoList {
		numNodes, err := strconv.ParseInt(regionInfo["num-nodes"], 10, 32)
		if err != nil {
			return nil, err
		}
		regionNodes := int32(numNodes)
		region := regionInfo["region"]
		totalNodes += int(regionNodes)
		cloudInfo := *ybmclient.NewCloudInfoWithDefaults()
		cloudInfo.SetRegion(region)
		if cmd.Flags().Changed("cloud-provider") {
			cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
			cloudInfo.SetCode(ybmclient.CloudEnum(cloudProvider))
		}
		info := *ybmclient.NewClusterRegionInfo(
			*ybmclient.NewPlacementInfo(cloudInfo, int32(regionNodes)),
		)
		if vpcName, ok := regionInfo["vpc"]; ok {
			vpcID, err := a.GetVpcIdByName(vpcName)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			info.PlacementInfo.SetVpcId(vpcID)
		}
		if cmd.Flags().Changed("cluster-type") {
			clusterType, _ := cmd.Flags().GetString("cluster-type")
			if clusterType == "GEO_PARTITIONED" {
				info.PlacementInfo.SetMultiZone(true)
			}
		}
		info.SetIsDefault(false)
		clusterRegionInfo = append(clusterRegionInfo, info)
	}

	// This is to populate region in top level cloud info
	region := ""
	regionCount := len(clusterRegionInfo)
	if regionCount > 0 {
		regionInfoProvided = true
		region = clusterRegionInfo[0].PlacementInfo.CloudInfo.Region
		if regionCount == 1 {
			clusterRegionInfo[0].SetIsDefault(true)
		}
	}

	// For the default tier which is FREE, isProduction has to be false
	isProduction := false

	clusterName, _ := cmd.Flags().GetString("cluster-name")
	if cmd.Flags().Changed("new-name") {
		clusterName, _ = cmd.Flags().GetString("new-name")
	}
	cloudInfo := *ybmclient.NewCloudInfoWithDefaults()
	if cmd.Flags().Changed("cloud-provider") {
		cloudProvider, _ := cmd.Flags().GetString("cloud-provider")
		cloudInfo.SetCode(ybmclient.CloudEnum(cloudProvider))
	}
	if regionInfoProvided {
		cloudInfo.SetRegion(region)
	}

	clusterInfo := *ybmclient.NewClusterInfoWithDefaults()
	if cmd.Flags().Changed("cluster-tier") {
		clusterTierCli, _ := cmd.Flags().GetString("cluster-tier")
		clusterTier, err := util.GetClusterTier(clusterTierCli)
		if err != nil {
			return nil, err
		}
		if clusterTier == "PAID" {
			isProduction = true
		}
		clusterInfo.SetClusterTier(ybmclient.ClusterTier(clusterTier))
	}

	if totalNodes != 0 {
		clusterInfo.SetNumNodes(int32(totalNodes))
	}
	if cmd.Flags().Changed("fault-tolerance") {
		faultTolerance, _ := cmd.Flags().GetString("fault-tolerance")
		clusterInfo.SetFaultTolerance(ybmclient.ClusterFaultTolerance(faultTolerance))
	}
	clusterInfo.SetIsProduction(isProduction)
	clusterInfo.SetNodeInfo(*ybmclient.NewClusterNodeInfoWithDefaults())

	if cmd.Flags().Changed("node-config") {
		nodeConfig, _ := cmd.Flags().GetStringToInt("node-config")
		numCores := nodeConfig["num-cores"]

		clusterInfo.NodeInfo.SetNumCores(int32(numCores))

		if diskSize, ok := nodeConfig["disk-size-gb"]; ok {
			diskSizeGb = int32(diskSize)
		}

	}

	cloud := string(cloudInfo.GetCode())
	region = cloudInfo.GetRegion()
	tier := string(clusterInfo.GetClusterTier())
	numCores := clusterInfo.NodeInfo.GetNumCores()

	memoryMb, err = a.GetFromInstanceType("memory", cloud, tier, region, int32(numCores))
	if err != nil {
		return nil, err
	}
	clusterInfo.NodeInfo.SetMemoryMb(memoryMb)

	// Computing the default disk size if it is not provided
	if diskSizeGb == 0 {
		diskSizeGb, err = a.GetFromInstanceType("disk", cloud, tier, region, int32(numCores))
		if err != nil {
			return nil, err
		}
	}
	clusterInfo.NodeInfo.SetDiskSizeGb(diskSizeGb)

	if cmd.Flags().Changed("cluster-type") {
		clusterType, _ := cmd.Flags().GetString("cluster-type")
		clusterInfo.SetClusterType(ybmclient.ClusterType(clusterType))
	}

	// Compute track ID for database version
	softwareInfo := *ybmclient.NewSoftwareInfoWithDefaults()
	if cmd.Flags().Changed("database-version") {
		trackName, _ = cmd.Flags().GetString("database-version")
		trackId, err = a.GetTrackIdByName(trackName)
		if err != nil {
			return nil, err
		}
		softwareInfo.SetTrackId(trackId)
	}

	clusterSpec := ybmclient.NewClusterSpec(
		clusterName,
		clusterInfo,
		softwareInfo)
	clusterSpec.SetCloudInfo(cloudInfo)
	if regionInfoProvided {
		clusterSpec.SetClusterRegionInfo(clusterRegionInfo)
	}

	return clusterSpec, nil
}

func (a *AuthApiClient) GetInfo(providedAccountID string, providedProjectID string) {
	var err error
	a.AccountID, err = a.GetAccountID(providedAccountID)
	if err != nil {
		logrus.Errorf("could not initiate api client: %s", err.Error())
		os.Exit(1)
	}
	a.ProjectID, err = a.GetProjectID(providedProjectID)
	if err != nil {
		logrus.Errorf("could not initiate api client: %s", err.Error())
		os.Exit(1)
	}
}

func (a *AuthApiClient) GetClusterByName(clusterName string) (ybmclient.ClusterData, error) {
	clusterResp, resp, err := a.ListClusters().Name(clusterName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return ybmclient.ClusterData{}, err
	}
	clusterData := clusterResp.GetData()

	if len(clusterData) != 0 {
		return clusterData[0], nil
	}

	return ybmclient.ClusterData{}, fmt.Errorf("could not get cluster data for cluster name: %s", clusterName)
}

func (a *AuthApiClient) GetEndpointsForClusterByName(clusterName string) ([]ybmclient.Endpoint, string, error) {
	clusterData, err := a.GetClusterByName(clusterName)
	if err != nil {
		return nil, "", err
	}
	clusterId := clusterData.Info.GetId()
	clusterEndpoints := clusterData.Info.GetClusterEndpoints()
	jsonEndpoints, _ := json.Marshal(clusterEndpoints)
	logrus.Debugf("Found endpoints: %v\n", string(jsonEndpoints))

	return clusterEndpoints, clusterId, nil
}

func (a *AuthApiClient) GetEndpointByIdForClusterByName(clusterName string, endpointId string) (ybmclient.Endpoint, string, error) {
	endpoints, clusterId, err := a.GetEndpointsForClusterByName(clusterName)
	if err != nil {
		// return the error
		return ybmclient.Endpoint{}, "", err
	}

	endpoints = util.Filter(endpoints, func(endpoint ybmclient.Endpoint) bool {
		return endpoint.GetId() == endpointId || endpoint.GetPseId() == endpointId
	})

	if len(endpoints) == 0 {
		logrus.Fatalf("Endpoint not found\n")
	}
	if len(endpoints) > 1 {
		logrus.Fatalf("Multiple endpoints found\n")
	}

	return endpoints[0], clusterId, nil
}

func (a *AuthApiClient) GetClusterIdByName(clusterName string) (string, error) {
	clusterData, err := a.GetClusterByName(clusterName)
	if err == nil {
		return clusterData.Info.GetId(), nil
	}

	return "", fmt.Errorf("could not get cluster data for cluster name: %s", clusterName)
}

func (a *AuthApiClient) CreateCluster() ybmclient.ApiCreateClusterRequest {
	return a.ApiClient.ClusterApi.CreateCluster(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) GetCluster(clusterId string) ybmclient.ApiGetClusterRequest {
	return a.ApiClient.ClusterApi.GetCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) GetClusterVpcById(clusterId string) (string, error) {
	clusterResp, resp, err := a.GetCluster(clusterId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return "", err
	}
	vpcId := clusterResp.Data.Spec.NetworkInfo.GetSingleTenantVpcId()
	return vpcId, nil
}

func (a *AuthApiClient) GetClusterCloudById(clusterId string) (ybmclient.CloudEnum, error) {
	clusterResp, resp, err := a.GetCluster(clusterId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return "", err
	}
	clusterCloud := clusterResp.Data.Spec.CloudInfo.GetCode()
	return clusterCloud, nil
}

func (a *AuthApiClient) GetConnectionCertificate() (string, error) {
	certResp, resp, err := a.ApiClient.ClusterApi.GetConnectionCertificate(a.ctx).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return "", err
	}
	certData := certResp.GetData()
	return certData, nil
}

func (a *AuthApiClient) EditCluster(clusterId string) ybmclient.ApiEditClusterRequest {
	return a.ApiClient.ClusterApi.EditCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) ListClusters() ybmclient.ApiListClustersRequest {
	return a.ApiClient.ClusterApi.ListClusters(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) DeleteCluster(clusterId string) ybmclient.ApiDeleteClusterRequest {
	return a.ApiClient.ClusterApi.DeleteCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) PauseCluster(clusterId string) ybmclient.ApiPauseClusterRequest {
	return a.ApiClient.ClusterApi.PauseCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) ResumeCluster(clusterId string) ybmclient.ApiResumeClusterRequest {
	return a.ApiClient.ClusterApi.ResumeCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) GetPrivateServiceEndpoint(clusterId string, endpointId string) ybmclient.ApiGetPrivateServiceEndpointRequest {
	return a.ApiClient.ClusterApi.GetPrivateServiceEndpoint(a.ctx, a.AccountID, a.ProjectID, clusterId, endpointId)
}

func (a *AuthApiClient) EditPrivateServiceEndpoint(clusterId string, endpointId string) ybmclient.ApiEditPrivateServiceEndpointRequest {
	return a.ApiClient.ClusterApi.EditPrivateServiceEndpoint(a.ctx, a.AccountID, a.ProjectID, clusterId, endpointId)
}

func (a *AuthApiClient) DeletePrivateServiceEndpoint(clusterId string, endpointId string) ybmclient.ApiDeletePrivateServiceEndpointRequest {
	return a.ApiClient.ClusterApi.DeletePrivateServiceEndpoint(a.ctx, a.AccountID, a.ProjectID, clusterId, endpointId)
}

func (a *AuthApiClient) CreatePrivateServiceEndpoint(clusterId string) ybmclient.ApiCreatePrivateServiceEndpointRequest {
	return a.ApiClient.ClusterApi.CreatePrivateServiceEndpoint(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) CreatePrivateServiceEndpointRegionSpec(regionArnMap map[string][]string) []ybmclient.PrivateServiceEndpointRegionSpec {
	pseSpecs := []ybmclient.PrivateServiceEndpointRegionSpec{}

	for regionId, arnList := range regionArnMap {
		local := *ybmclient.NewPrivateServiceEndpointRegionSpec(regionId, arnList)
		pseSpecs = append(pseSpecs, local)
	}
	return pseSpecs
}

func (a *AuthApiClient) CreatePrivateServiceEndpointSpec(regionArnMap map[string][]string) []ybmclient.PrivateServiceEndpointSpec {
	pseSpecs := []ybmclient.PrivateServiceEndpointSpec{}

	for regionId, arnList := range regionArnMap {
		regionSpec := *ybmclient.NewPrivateServiceEndpointRegionSpec(regionId, arnList)
		local := *ybmclient.NewPrivateServiceEndpointSpec([]ybmclient.PrivateServiceEndpointRegionSpec{regionSpec})
		pseSpecs = append(pseSpecs, local)
	}
	return pseSpecs
}

func (a *AuthApiClient) CreateReadReplica(clusterId string) ybmclient.ApiCreateReadReplicaRequest {
	return a.ApiClient.ReadReplicaApi.CreateReadReplica(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) EditReadReplicas(clusterId string) ybmclient.ApiEditReadReplicasRequest {
	return a.ApiClient.ReadReplicaApi.EditReadReplicas(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) ListReadReplicas(clusterId string) ybmclient.ApiListReadReplicasRequest {
	return a.ApiClient.ReadReplicaApi.ListReadReplicas(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) DeleteReadReplica(clusterId string) ybmclient.ApiDeleteReadReplicaRequest {
	return a.ApiClient.ReadReplicaApi.DeleteReadReplica(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) CreateVpc() ybmclient.ApiCreateVpcRequest {
	return a.ApiClient.NetworkApi.CreateVpc(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListSingleTenantVpcs() ybmclient.ApiListSingleTenantVpcsRequest {
	return a.ApiClient.NetworkApi.ListSingleTenantVpcs(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListSingleTenantVpcsByName(name string) ybmclient.ApiListSingleTenantVpcsRequest {
	if name == "" {
		return a.ListSingleTenantVpcs()
	}
	return a.ListSingleTenantVpcs().Name(name)
}

func (a *AuthApiClient) DeleteVpc(vpcId string) ybmclient.ApiDeleteVpcRequest {
	return a.ApiClient.NetworkApi.DeleteVpc(a.ctx, a.AccountID, a.ProjectID, vpcId)
}

func (a *AuthApiClient) GetVpcIdByName(vpcName string) (string, error) {
	vpcResp, resp, err := a.ListSingleTenantVpcs().Name(vpcName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return "", err
	}
	vpcData := vpcResp.GetData()

	if len(vpcData) != 0 {
		return vpcData[0].Info.GetId(), nil
	}

	return "", fmt.Errorf("could not get vpc data for vpc name: %s", vpcName)
}

func (a *AuthApiClient) GetSingleTenantVpc(vpcId string) ybmclient.ApiGetSingleTenantVpcRequest {
	return a.ApiClient.NetworkApi.GetSingleTenantVpc(a.ctx, a.AccountID, a.ProjectID, vpcId)
}

func (a *AuthApiClient) GetVpcNameById(vpcId string) (string, error) {
	vpcNameResp, resp, err := a.GetSingleTenantVpc(vpcId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(b)
		return "", err
	}
	return vpcNameResp.GetData().Spec.Name, nil
}

func (a *AuthApiClient) GetVpcPeering(vpcPeeringID string) ybmclient.ApiGetVpcPeeringRequest {
	return a.ApiClient.NetworkApi.GetVpcPeering(a.ctx, a.AccountID, a.ProjectID, vpcPeeringID)
}

func (a *AuthApiClient) CreateVpcPeering() ybmclient.ApiCreateVpcPeeringRequest {
	return a.ApiClient.NetworkApi.CreateVpcPeering(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListVpcPeerings() ybmclient.ApiListVpcPeeringsRequest {
	return a.ApiClient.NetworkApi.ListVpcPeerings(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) DeleteVpcPeering(vpcPeeringId string) ybmclient.ApiDeleteVpcPeeringRequest {
	return a.ApiClient.NetworkApi.DeleteVpcPeering(a.ctx, a.AccountID, a.ProjectID, vpcPeeringId)
}

func (a *AuthApiClient) CreateNetworkAllowList() ybmclient.ApiCreateNetworkAllowListRequest {
	return a.ApiClient.NetworkApi.CreateNetworkAllowList(a.ctx, a.AccountID, a.ProjectID)
}
func (a *AuthApiClient) DeleteNetworkAllowList(allowListId string) ybmclient.ApiDeleteNetworkAllowListRequest {
	return a.ApiClient.NetworkApi.DeleteNetworkAllowList(a.ctx, a.AccountID, a.ProjectID, allowListId)
}
func (a *AuthApiClient) ListNetworkAllowLists() ybmclient.ApiListNetworkAllowListsRequest {
	return a.ApiClient.NetworkApi.ListNetworkAllowLists(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) GetBackup(backupID string) ybmclient.ApiGetBackupRequest {
	return a.ApiClient.BackupApi.GetBackup(a.ctx, a.AccountID, a.ProjectID, backupID)
}

func (a *AuthApiClient) GetNetworkAllowListIdByName(networkAllowListName string) (string, error) {
	nalResp, resp, err := a.ListNetworkAllowLists().Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return "", err
	}
	nalData, err := util.FindNetworkAllowList(nalResp.Data, networkAllowListName)

	if err != nil {
		return "", err
	}

	return nalData.Info.GetId(), nil

}

func (a *AuthApiClient) EditClusterNetworkAllowLists(clusterId string, allowListIds []string) ybmclient.ApiEditClusterNetworkAllowListsRequest {
	return a.ApiClient.ClusterApi.EditClusterNetworkAllowLists(a.ctx, a.AccountID, a.ProjectID, clusterId).RequestBody(allowListIds)
}

func (a *AuthApiClient) ListClusterNetworkAllowLists(clusterId string) ybmclient.ApiListClusterNetworkAllowListsRequest {
	return a.ApiClient.ClusterApi.ListClusterNetworkAllowLists(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) ListClusterCMKs(clusterId string) ybmclient.ApiGetClusterCMKRequest {
	return a.ApiClient.ClusterApi.GetClusterCMK(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) EditClusterCMKs(clusterId string) ybmclient.ApiEditClusterCMKRequest {
	return a.ApiClient.ClusterApi.EditClusterCMK(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) ListBackups() ybmclient.ApiListBackupsRequest {
	return a.ApiClient.BackupApi.ListBackups(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) RestoreBackup() ybmclient.ApiRestoreBackupRequest {
	return a.ApiClient.BackupApi.RestoreBackup(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) CreateBackup() ybmclient.ApiCreateBackupRequest {
	return a.ApiClient.BackupApi.CreateBackup(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) DeleteBackup(backupId string) ybmclient.ApiDeleteBackupRequest {
	return a.ApiClient.BackupApi.DeleteBackup(a.ctx, a.AccountID, a.ProjectID, backupId)
}

func (a *AuthApiClient) GetTrack(trackId string) ybmclient.ApiGetTrackRequest {
	return a.ApiClient.SoftwareReleaseApi.GetTrack(a.ctx, a.AccountID, trackId)
}

func (a *AuthApiClient) ListTracks() ybmclient.ApiListTracksRequest {
	return a.ApiClient.SoftwareReleaseApi.ListTracks(a.ctx, a.AccountID)
}

func (a *AuthApiClient) GetTrackNameById(trackId string) (string, error) {
	trackNameResp, resp, err := a.GetTrack(trackId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(b)
		return "", err
	}

	return trackNameResp.GetData().Spec.GetName(), nil

}

func (a *AuthApiClient) GetTrackIdByName(trackName string) (string, error) {
	tracksNameResp, resp, err := a.ListTracks().Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(b)
		return "", err
	}

	for _, track := range tracksNameResp.GetData() {
		if track.Spec.GetName() == trackName {
			return track.Info.GetId(), nil
		}
	}
	return "", fmt.Errorf("the database version doesn't exist")

}
func (a *AuthApiClient) CreateCdcStream(clusterId string) ybmclient.ApiCreateCdcStreamRequest {
	return a.ApiClient.CdcApi.CreateCdcStream(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) DeleteCdcStream(cdcStreamId string, clusterId string) ybmclient.ApiDeleteCdcStreamRequest {
	return a.ApiClient.CdcApi.DeleteCdcStream(a.ctx, a.AccountID, a.ProjectID, clusterId, cdcStreamId)
}

func (a *AuthApiClient) EditCdcStream(cdcStreamId string, clusterId string) ybmclient.ApiEditCdcStreamRequest {
	return a.ApiClient.CdcApi.EditCdcStream(a.ctx, a.AccountID, a.ProjectID, clusterId, cdcStreamId)
}

func (a *AuthApiClient) ListCdcStreamsForAccount() ybmclient.ApiListCdcStreamsForAccountRequest {
	return a.ApiClient.CdcApi.ListCdcStreamsForAccount(a.ctx, a.AccountID)
}

func (a *AuthApiClient) GetCdcStream(cdcStreamId string, clusterId string) ybmclient.ApiGetCdcStreamRequest {
	return a.ApiClient.CdcApi.GetCdcStream(a.ctx, a.AccountID, a.ProjectID, clusterId, cdcStreamId)
}

func (a *AuthApiClient) GetCdcStreamIDByStreamName(cdcStreamName string) (string, error) {
	streamResp, resp, err := a.ListCdcStreamsForAccount().Name(cdcStreamName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(b)
		return "", err
	}
	streamData := streamResp.GetData()

	if len(streamData) != 0 {
		return streamData[0].Info.GetId(), nil
	}

	return "", fmt.Errorf("couldn't find any cdcStream with the given name")
}

func (a *AuthApiClient) GetSupportedNodeConfigurations(cloud string, tier string, region string) ybmclient.ApiGetSupportedNodeConfigurationsRequest {
	return a.ApiClient.ClusterApi.GetSupportedNodeConfigurations(a.ctx).AccountId(a.AccountID).Cloud(cloud).Tier(tier).Regions([]string{region})
}

func (a *AuthApiClient) GetFromInstanceType(resource string, cloud string, tier string, region string, numCores int32) (int32, error) {
	instanceResp, resp, err := a.GetSupportedNodeConfigurations(cloud, tier, region).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(b)
		return 0, err
	}
	instanceData := instanceResp.GetData()
	nodeConfigList, ok := instanceData[region]
	if !ok || len(nodeConfigList) == 0 {
		return 0, fmt.Errorf("no instances configured for the given region")
	}

	return getFromNodeConfig(resource, numCores, nodeConfigList)

}

func (a *AuthApiClient) CreateCdcSink() ybmclient.ApiCreateCdcSinkRequest {
	return a.ApiClient.CdcApi.CreateCdcSink(a.ctx, a.AccountID)
}

func (a *AuthApiClient) DeleteCdcSink(cdcSinkId string) ybmclient.ApiDeleteCdcSinkRequest {
	return a.ApiClient.CdcApi.DeleteCdcSink(a.ctx, a.AccountID, cdcSinkId)
}

func (a *AuthApiClient) EditCdcSink(cdcSinkId string) ybmclient.ApiEditCdcSinkRequest {
	return a.ApiClient.CdcApi.EditCdcSink(a.ctx, a.AccountID, cdcSinkId)
}

func (a *AuthApiClient) GetCdcSink(cdcSinkId string) ybmclient.ApiGetCdcSinkRequest {
	return a.ApiClient.CdcApi.GetCdcSink(a.ctx, a.AccountID, cdcSinkId)
}

func (a *AuthApiClient) ListCdcSinks() ybmclient.ApiListCdcSinksRequest {
	return a.ApiClient.CdcApi.ListCdcSinks(a.ctx, a.AccountID)
}

func (a *AuthApiClient) GetCdcSinkIDBySinkName(cdcSinkName string) (string, error) {
	sinkResp, resp, err := a.ListCdcSinks().Name(cdcSinkName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(b)
		return "", err
	}

	sinkData := sinkResp.GetData()

	if len(sinkData) != 0 {
		return sinkData[0].Info.GetId(), nil
	}

	return "", fmt.Errorf("couldn't find any cdcSink with the given name")
}

func (a *AuthApiClient) GetClusterNode(clusterId string) ybmclient.ApiGetClusterNodesRequest {
	return a.ApiClient.ClusterApi.GetClusterNodes(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) PerformNodeOperation(clusterId string) ybmclient.ApiPerformNodeOperationRequest {
	return a.ApiClient.ClusterApi.PerformNodeOperation(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) GetSupportedCloudRegions() ybmclient.ApiGetSupportedCloudRegionsRequest {
	return a.ApiClient.ClusterApi.GetSupportedCloudRegions(a.ctx)
}
func (a *AuthApiClient) ListTasks() ybmclient.ApiListTasksRequest {
	return a.ApiClient.TaskApi.ListTasks(a.ctx, a.AccountID)
}

func (a *AuthApiClient) WaitForTaskCompletion(entityId string, entityType ybmclient.EntityTypeEnum, taskType ybmclient.TaskTypeEnum, completionStatus []string, message string) (string, error) {

	if strings.ToLower(os.Getenv("YBM_CI")) == "true" {
		return a.WaitForTaskCompletionCI(entityId, entityType, taskType, completionStatus, message)
	}
	return a.WaitForTaskCompletionFull(entityId, entityType, taskType, completionStatus, message)

}

func (a *AuthApiClient) WaitForTaskCompletionCI(entityId string, entityType ybmclient.EntityTypeEnum, taskType ybmclient.TaskTypeEnum, completionStatus []string, message string) (string, error) {
	var taskList ybmclient.TaskListResponse
	var resp *http.Response
	var err error
	currentStatus := "UNKNOWN"
	previousStatus := "UNKNOWN"
	output := fmt.Sprintf(" %s: %s", message, currentStatus)
	timeout := time.After(viper.GetDuration("timeout"))
	checkEveryInSec := time.Tick(2 * time.Second)
	fmt.Println(output)
	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("wait timeout, operation could still be on-going")
		case <-a.ctx.Done():
			return "", fmt.Errorf("receive interrupt signal, operation could still be on-going")
		case <-checkEveryInSec:
			apiRequest := a.ListTasks().TaskType(taskType).ProjectId(a.ProjectID).EntityId(entityId).Limit(1)
			//Sometime the api do not need any entity type, for example VPC, VPC_PEERING
			if len(entityType) > 0 {
				apiRequest.EntityType(entityType)
			}
			taskList, resp, err = apiRequest.Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", resp)
				return "", fmt.Errorf("error when calling `TaskApi.ListTasks`: %s", GetApiErrorDetails(err))
			}

			if v, ok := taskList.GetDataOk(); ok && v != nil {
				c := taskList.GetData()
				if len(c) > 0 {
					if status, ok := c[0].GetInfoOk(); ok {
						previousStatus = currentStatus
						currentStatus = status.GetState()
					}
					output = fmt.Sprintf(" %s: %s", message, currentStatus)
					if taskProgressInfo, _ := c[0].Info.GetTaskProgressInfoOk(); ok && taskProgressInfo != nil {

						for index, action := range taskProgressInfo.GetActions() {
							output = output + "\n" + ". Task " + strconv.Itoa(index+1) + ": " + action.GetName() + " " + strconv.Itoa(int(action.GetPercentComplete())) + "% completed"
						}
					}
				}
			}
			if slices.Contains(completionStatus, currentStatus) {
				return currentStatus, nil
			}
			if previousStatus != currentStatus {
				fmt.Println(output)
			}
		}
	}

}

func (a *AuthApiClient) WaitForTaskCompletionFull(entityId string, entityType ybmclient.EntityTypeEnum, taskType ybmclient.TaskTypeEnum, completionStatus []string, message string) (string, error) {
	var taskList ybmclient.TaskListResponse
	var resp *http.Response
	var err error

	currentStatus := "UNKNOWN"
	output := fmt.Sprintf(" %s: %s", message, currentStatus)
	s := spinner.New(spinner.CharSets[36], 300*time.Millisecond)
	s.Color("green", "bold")
	// start animating the spinner
	s.Start()
	s.Suffix = " " + output
	s.FinalMSG = ""
	defer s.Stop()
	timeout := time.After(viper.GetDuration("timeout"))
	checkEveryInSec := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			s.Stop()
			return "", fmt.Errorf("wait timeout, operation could still be on-going")
		case <-a.ctx.Done():
			s.Stop()
			return "", fmt.Errorf("receive interrupt signal, operation could still be on-going")
		case <-checkEveryInSec:
			apiRequest := a.ListTasks().TaskType(taskType).ProjectId(a.ProjectID).EntityId(entityId).Limit(1)
			//Sometime the api do not need any entity type, for example VPC, VPC_PEERING
			if len(entityType) > 0 {
				apiRequest.EntityType(entityType)
			}
			taskList, resp, err = apiRequest.Execute()
			if err != nil {
				logrus.Debugf("Full HTTP response: %v", resp)
				return "", fmt.Errorf("error when calling `TaskApi.ListTasks`: %s", GetApiErrorDetails(err))
			}

			if v, ok := taskList.GetDataOk(); ok && v != nil {
				c := taskList.GetData()
				if len(c) > 0 {
					if status, ok := c[0].GetInfoOk(); ok {
						currentStatus = status.GetState()
					}
					output = fmt.Sprintf(" %s: %s", message, currentStatus)
					if taskProgressInfo, _ := c[0].Info.GetTaskProgressInfoOk(); ok && taskProgressInfo != nil {

						for index, action := range taskProgressInfo.GetActions() {
							output = output + "\n" + ". Task " + strconv.Itoa(index+1) + ": " + action.GetName() + " " + strconv.Itoa(int(action.GetPercentComplete())) + "% completed"
						}
					}
				}
			}
			s.Suffix = output
			if slices.Contains(completionStatus, currentStatus) {
				return currentStatus, nil
			}
		}
	}

}

func getFromNodeConfig(resource string, numCores int32, nodeConfigList []ybmclient.NodeConfigurationResponseItem) (int32, error) {
	resourceValue := int32(0)
	for _, nodeConfig := range nodeConfigList {
		if nodeConfig.GetNumCores() == numCores {
			switch resource {
			case "disk":
				resourceValue = nodeConfig.GetIncludedDiskSizeGb()
			case "memory":
				resourceValue = nodeConfig.GetMemoryMb()
			}
			return resourceValue, nil
		}
	}
	return 0, fmt.Errorf("node with the given number of CPU cores doesn't exist in the given region")
}

// Utils functions

// GetApiErrorDetails will return the api Error message if present
// If not present will return the original err.Error()
func GetApiErrorDetails(err error) string {
	switch castedError := err.(type) {
	case ybmclient.GenericOpenAPIError:
		if v := getAPIError(castedError.Body()); v != nil {
			if d, ok := v.GetErrorOk(); ok {
				return fmt.Sprintf("%s%s", d.GetDetail(), "\n")
			}
		}
	}
	return err.Error()

}

func getAPIError(b []byte) *ybmclient.ApiError {
	apiError := ybmclient.NewApiErrorWithDefaults()
	err := json.Unmarshal(b, &apiError)
	if err != nil {
		return nil
	}
	return apiError
}

func ParseURL(host string) (*url.URL, error) {
	if strings.HasPrefix(strings.ToLower(host), "http://") {
		logrus.Warnf("you are using insecure api endpoint %s", host)
	} else if !strings.HasPrefix(strings.ToLower(host), "https://") {
		host = "https://" + host
	}

	endpoint, err := url.ParseRequestURI(host)
	if err != nil {
		return nil, fmt.Errorf("could not parse ybm server url (%s): %w", host, err)
	}
	return endpoint, err
}

func getErrorMessage(response *http.Response, err error) string {
	errMsg := err.Error()
	if response != nil {
		request, dumpErr := httputil.DumpRequest(response.Request, true)
		if dumpErr != nil {
			additional := "Error while dumping request: " + dumpErr.Error()
			errMsg = errMsg + "\n\n\nDump error:" + additional
		} else {
			reqString := string(request)
			// Replace the Authorization Bearer header with obfuscated value
			re := regexp.MustCompile(`eyJ(.*)`)
			reqString = re.ReplaceAllString(reqString, `***`)
			errMsg = errMsg + "\n\nAPI Request:\n" + reqString
		}

		response, dumpErr := httputil.DumpResponse(response, true)
		if dumpErr != nil {
			additional := "Error while dumping response: " + dumpErr.Error()
			errMsg = errMsg + "\n\n\nDump error:" + additional
		} else {
			errMsg = errMsg + "\n\nAPI Response:\n" + string(response)
		}
	}
	return errMsg
}
