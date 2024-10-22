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
	"sort"
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

// AuthApiClient is a auth YugabyteDB Aeon Client

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
		logrus.Fatalln("No valid API key detected. Please run `ybm auth` to authenticate with YugabyteDB Aeon.")
	}

	return NewAuthApiClientCustomUrlKey(url, apiKey)
}

func NewAuthApiClientCustomUrlKey(url *url.URL, apiKey string) (*AuthApiClient, error) {
	configuration := ybmclient.NewConfiguration()
	//Configure the client

	configuration.Host = url.Host
	//configuration.Debug = true
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
	accountResp, _, err := a.ListAccounts().Execute()
	if err != nil {
		return "", err
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

	accountResp, _, err := a.ListAccounts().Execute()
	if err != nil {
		return "", err
	}
	projectData := accountResp.GetData()[0].Info.GetProjects()
	if len(projectData) == 0 {
		return "", fmt.Errorf("the account is not associated with any projects")
	}
	if len(projectData) > 1 {
		return "", fmt.Errorf("the account is associated with multiple projects, please provide a project id")
	}
	return projectData[0].Info.Id, nil
}

func (a *AuthApiClient) buildClusterSpec(cmd *cobra.Command, regionInfoList []map[string]string, regionNodeConfigsMap map[string][]ybmclient.NodeConfigurationResponseItem) (*ybmclient.ClusterSpec, error) {

	var diskSizeGb int32
	var diskIops int32
	var memoryMb int32
	var trackId string
	var trackName string
	var regionInfoProvided bool
	var err error

	clusterRegionInfo := []ybmclient.ClusterRegionInfo{}
	totalNodes := 0
	regionNodeInfoMap := map[string]*ybmclient.OptionalClusterNodeInfo{}
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

		regionNodeInfo := ybmclient.NewOptionalClusterNodeInfo(0, 0, 0)
		if numCores, ok := regionInfo["num-cores"]; ok {
			i, err := strconv.Atoi(numCores)
			if err != nil {
				logrus.Fatalf("Unable to parse num-cores integer in %s", region)
			}
			regionNodeInfo.SetNumCores(int32(i))
		}
		if diskSizeGb, ok := regionInfo["disk-size-gb"]; ok {
			i, err := strconv.Atoi(diskSizeGb)
			if err != nil {
				logrus.Fatalf("Unable to parse disk-size-gb integer in %s", region)
			}
			regionNodeInfo.SetDiskSizeGb(int32(i))
		}
		if diskIops, ok := regionInfo["disk-iops"]; ok {
			i, err := strconv.Atoi(diskIops)
			if err != nil {
				logrus.Fatalf("Unable to parse disk-iops integer in %s", region)
			}
			regionNodeInfo.SetDiskIops(int32(i))
		}
		regionNodeInfoMap[region] = regionNodeInfo
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
	if cmd.Flags().Changed("preferred-region") {
		preferredRegion, _ := cmd.Flags().GetString("preferred-region")
		if clusterInfo.GetFaultTolerance() != ybmclient.ClusterFaultTolerance("REGION") {
			return nil, fmt.Errorf("preferred region is allowed only for regional level fault tolerance")
		}

		if len(clusterRegionInfo) <= 1 {
			return nil, fmt.Errorf("preferred region is allowed only if there is more than one region")
		}

		err = util.SetPreferredRegion(clusterRegionInfo, preferredRegion)
		if err != nil {
			return nil, err
		}

	}

	if cmd.Flags().Changed("num-faults-to-tolerate") {
		numFaultsToTolerate, _ := cmd.Flags().GetInt32("num-faults-to-tolerate")
		if valid, err := util.ValidateNumFaultsToTolerate(numFaultsToTolerate, clusterInfo.GetFaultTolerance()); !valid {
			return nil, err
		}
		clusterInfo.SetNumFaultsToTolerate(numFaultsToTolerate)
	}
	if util.IsFeatureFlagEnabled(util.ENTERPRISE_SECURITY) {
		if cmd.Flags().Changed("enterprise-security") {
			enterpriseSecurity, _ := cmd.Flags().GetBool("enterprise-security")
			clusterInfo.SetEnterpriseSecurity(enterpriseSecurity)
		}
	}
	clusterInfo.SetIsProduction(isProduction)

	if cmd.Flags().Changed("cluster-type") {
		clusterType, _ := cmd.Flags().GetString("cluster-type")
		clusterInfo.SetClusterType(ybmclient.ClusterType(clusterType))
	}

	cloud := string(cloudInfo.GetCode())
	tier := string(clusterInfo.GetClusterTier())

	if regionInfoProvided {
		geoPartitioned := clusterInfo.GetClusterType() == "GEO_PARTITIONED"
		clusterNodeInfoWithDefaults := *ybmclient.NewClusterNodeInfoWithDefaults()
		// Create slice of desired regions.
		regions := make([]string, 0, len(regionNodeInfoMap))
		for k, _ := range regionNodeInfoMap {
			regions = append(regions, k)
		}

		if regionNodeConfigsMap == nil {
			// Grab available node configurations by region.
			regionNodeConfigsMap = a.GetSupportedNodeConfigurationsV2(cloud, tier, regions, geoPartitioned)
		}

		// Create slice of region keys of node configurations response.
		nodeConfigurationsRegions := make([]string, 0, len(regionNodeConfigsMap))
		for k, _ := range regionNodeConfigsMap {
			nodeConfigurationsRegions = append(nodeConfigurationsRegions, k)
		}
		// For each desired region, grab appropriate node configuration and set node info.
		for _, r := range regions {
			nodeConfigs := []ybmclient.NodeConfigurationResponseItem{}
			if slices.Contains(nodeConfigurationsRegions, r) {
				nodeConfigs = regionNodeConfigsMap[r]
			} else {
				// Requested region not found in node configurations map.
				// In this case, the map key is a string of all (comma-separated) regions,
				// and the value is a list of node configurations that are available in all regions.
				// So, we just use look through the first map value to find a node configuration to use.
				nodeConfigs = regionNodeConfigsMap[nodeConfigurationsRegions[0]]
			}
			requestedNodeInfo := regionNodeInfoMap[r]
			requestedNumCores := requestedNodeInfo.GetNumCores()
			userProvidedNumCores := requestedNumCores != 0

			var nodeConfig *ybmclient.NodeConfigurationResponseItem = nil
			if !userProvidedNumCores {
				requestedNumCores = clusterNodeInfoWithDefaults.GetNumCores()
			}
			for i, nc := range nodeConfigs {
				if nc.GetNumCores() == requestedNumCores {
					nodeConfig = &nodeConfigs[i]
					break
				}
			}
			if nodeConfig == nil {
				logrus.Fatalf("No instance type found with %d cores in region %s\n", requestedNumCores, r)
			}
			regionNodeInfoMap[r].SetNumCores(nodeConfig.GetNumCores())
			regionNodeInfoMap[r].SetMemoryMb(nodeConfig.GetMemoryMb())
			if requestedNodeInfo.GetDiskSizeGb() == 0 {
				// User did not specify a disk size. Default to included disk size.
				regionNodeInfoMap[r].SetDiskSizeGb(nodeConfig.GetIncludedDiskSizeGb())
			}
		}
		// Set per-region node info and cluster node info.
		var currRegionNodeInfo *ybmclient.OptionalClusterNodeInfo = nil
		for i, regionInfo := range clusterRegionInfo {
			r := regionInfo.GetPlacementInfo().CloudInfo.Region
			clusterRegionInfo[i].SetNodeInfo(*regionNodeInfoMap[r])
			logrus.Debugf("region=%s, node-info=%v\n", r, clusterRegionInfo[i].GetNodeInfo())
			if currRegionNodeInfo != nil && !geoPartitioned && *currRegionNodeInfo != clusterRegionInfo[i].GetNodeInfo() {
				// Asymmetric node configurations are only allowed for geo-partitioned clusters.
				logrus.Fatalln("Synchronous cluster regions must have identical node configurations")
			}
			currRegionNodeInfo = (&clusterRegionInfo[i]).NodeInfo.Get()
		}
		clusterInfo.SetNodeInfo(ToClusterNodeInfo(regionNodeInfoMap[regions[0]]))
	} else {
		clusterInfo.SetNodeInfo(*ybmclient.NewClusterNodeInfoWithDefaults())
		if cmd.Flags().Changed("node-config") {
			nodeConfig, _ := cmd.Flags().GetStringToInt("node-config")
			numCores := nodeConfig["num-cores"]
			clusterInfo.NodeInfo.Get().SetNumCores(int32(numCores))
			if diskSize, ok := nodeConfig["disk-size-gb"]; ok {
				diskSizeGb = int32(diskSize)
			}
			if diskIopsInt, ok := nodeConfig["disk-iops"]; ok {
				diskIops = int32(diskIopsInt)
			}
		}
		region = cloudInfo.GetRegion()
		numCores := clusterInfo.NodeInfo.Get().GetNumCores()
		memoryMb, err = a.GetFromInstanceType("memory", cloud, tier, region, numCores)
		if err != nil {
			return nil, err
		}
		clusterInfo.NodeInfo.Get().SetMemoryMb(memoryMb)

		// Computing the default disk size if it is not provided
		if diskSizeGb == 0 {
			diskSizeGb, err = a.GetFromInstanceType("disk", cloud, tier, region, numCores)
			if err != nil {
				return nil, err
			}
		}
		clusterInfo.NodeInfo.Get().SetDiskSizeGb(diskSizeGb)

		if diskIops > 0 {
			clusterInfo.NodeInfo.Get().SetDiskIops(diskIops)
		}
	}

	if cmd.Flags().Changed("default-region") {
		defaultRegion, _ := cmd.Flags().GetString("default-region")
		if clusterInfo.GetClusterType() != ybmclient.ClusterType("GEO_PARTITIONED") {
			return nil, fmt.Errorf("default region is allowed only for geo partitioned clusters")
		}

		if len(clusterRegionInfo) <= 1 {
			return nil, fmt.Errorf("default region is allowed only if there is more than one region")
		}

		err = util.SetDefaultRegion(clusterRegionInfo, defaultRegion)
		if err != nil {
			return nil, err
		}

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

func (a *AuthApiClient) CreateClusterSpec(cmd *cobra.Command, regionInfoList []map[string]string) (*ybmclient.ClusterSpec, error) {
	return a.buildClusterSpec(cmd, regionInfoList, nil)
}

func (a *AuthApiClient) EditClusterSpec(cmd *cobra.Command, regionInfoList []map[string]string, clusterID string) (*ybmclient.ClusterSpec, error) {
	regions := make([]string, 0, len(regionInfoList))
	for _, regionInfo := range regionInfoList {
		region := regionInfo["region"]
		regions = append(regions, region)
	}
	regionNodeConfigsMap := a.GetSupportedNodeConfigurationsForEdit(clusterID, regions)
	return a.buildClusterSpec(cmd, regionInfoList, regionNodeConfigsMap)
}

func ToClusterNodeInfo(opt *ybmclient.OptionalClusterNodeInfo) ybmclient.ClusterNodeInfo {
	clusterNodeInfo := *ybmclient.NewClusterNodeInfoWithDefaults()
	clusterNodeInfo.SetNumCores(opt.GetNumCores())
	clusterNodeInfo.SetMemoryMb(opt.GetMemoryMb())
	clusterNodeInfo.SetDiskSizeGb(opt.GetDiskSizeGb())
	if iops, _ := opt.GetDiskIopsOk(); iops != nil {
		clusterNodeInfo.SetDiskIops(*iops)
	}
	return clusterNodeInfo
}

func (a *AuthApiClient) GetInfo(providedAccountID string, providedProjectID string) {
	var err error
	a.AccountID, err = a.GetAccountID(providedAccountID)
	if err != nil {
		logrus.Errorf(GetApiErrorDetails(err))
		os.Exit(1)
	}
	a.ProjectID, err = a.GetProjectID(providedProjectID)
	if err != nil {
		logrus.Errorf(GetApiErrorDetails(err))
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

func (a *AuthApiClient) GetDrByName(clusterId string, drName string) (ybmclient.XClusterDrData, error) {
	drResp, resp, err := a.ListXClusterDr(clusterId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return ybmclient.XClusterDrData{}, err
	}
	drData := drResp.GetData()

	for _, drDatum := range drData {
		if drDatum.Spec.GetName() == drName {
			return drDatum, nil
		}
	}

	return ybmclient.XClusterDrData{}, fmt.Errorf("could not get dr data for dr name: %s", drName)
}

func (a *AuthApiClient) ExtractProviderFromClusterName(clusterId string) ([]string, error) {
	clusterResp, _, err := a.GetCluster(clusterId).Execute()
	clusterData := clusterResp.GetData()
	providers := []string{}
	if err != nil {
		return nil, err
	}

	if ok := clusterData.Spec.HasClusterRegionInfo(); ok {
		if len(clusterData.GetSpec().ClusterRegionInfo) > 0 {
			sort.Slice(clusterData.GetSpec().ClusterRegionInfo, func(i, j int) bool {
				return string(clusterData.GetSpec().ClusterRegionInfo[i].PlacementInfo.CloudInfo.Code) < string(clusterData.GetSpec().ClusterRegionInfo[j].PlacementInfo.CloudInfo.Code)
			})
			for _, p := range clusterData.GetSpec().ClusterRegionInfo {
				//Check uniqueness of Cloud (in case multi cloud with strange distribution, AWS, GCP,AWS)
				if !slices.Contains(providers, string(p.PlacementInfo.CloudInfo.Code)) {
					providers = append(providers, string(p.PlacementInfo.CloudInfo.Code))
				}
			}
		}
	}
	return providers, nil
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

func (a *AuthApiClient) GetDrIdByName(clusterId string, drName string) (string, error) {
	drData, err := a.GetDrByName(clusterId, drName)
	if err == nil {
		return drData.Info.GetId(), nil
	}

	return "", fmt.Errorf("could not get dr data for dr name: %s", drName)
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

func (a *AuthApiClient) AssignDbAuditLogsExporterConfig(clusterId string) ybmclient.ApiAssociateDbAuditExporterConfigRequest {
	return a.ApiClient.ClusterApi.AssociateDbAuditExporterConfig(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) UpdateDbAuditExporterConfig(clusterId string, integrationId string) ybmclient.ApiUpdateDbAuditExporterConfigRequest {
	return a.ApiClient.ClusterApi.UpdateDbAuditExporterConfig(a.ctx, a.AccountID, a.ProjectID, clusterId, integrationId)
}

func (a *AuthApiClient) CreatePrivateServiceEndpointRegionSpec(regionArnMap map[string][]string) []ybmclient.PrivateServiceEndpointRegionSpec {
	pseSpecs := []ybmclient.PrivateServiceEndpointRegionSpec{}

	for regionId, arnList := range regionArnMap {
		local := *ybmclient.NewPrivateServiceEndpointRegionSpec(arnList)
		local.ClusterRegionInfoId = *ybmclient.NewNullableString(&regionId)
		pseSpecs = append(pseSpecs, local)
	}
	return pseSpecs
}

func (a *AuthApiClient) CreatePrivateServiceEndpointSpec(regionArnMap map[string][]string) []ybmclient.PrivateServiceEndpointSpec {
	pseSpecs := []ybmclient.PrivateServiceEndpointSpec{}

	for regionId, arnList := range regionArnMap {
		regionSpec := *ybmclient.NewPrivateServiceEndpointRegionSpec(arnList)
		regionSpec.ClusterRegionInfoId = *ybmclient.NewNullableString(&regionId)
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

func (a *AuthApiClient) GetBillingUsage(startTimestamp string, endTimestamp string, clusterIds []string) ybmclient.ApiGetBillingUsageRequest {
	return a.ApiClient.BillingApi.GetBillingUsage(a.ctx, a.AccountID).StartTimestamp(startTimestamp).EndTimestamp(endTimestamp).Granularity(ybmclient.GRANULARITYENUM_DAILY).ClusterIds(clusterIds)
}

func (a *AuthApiClient) ListClustersByDateRange(startTimestamp string, endTimestamp string) ybmclient.ApiListClustersByDateRangeRequest {
	return a.ApiClient.BillingApi.ListClustersByDateRange(a.ctx, a.AccountID).StartTimestamp(startTimestamp).EndTimestamp(endTimestamp).Tier(ybmclient.CLUSTERTIER_PAID)
}

func (a *AuthApiClient) ListClusterCMKs(clusterId string) ybmclient.ApiGetClusterCMKRequest {
	return a.ApiClient.ClusterApi.GetClusterCMK(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) EditClusterCMKs(clusterId string) ybmclient.ApiEditClusterCMKRequest {
	return a.ApiClient.ClusterApi.EditClusterCMK(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) UpdateClusterCmkState(clusterId string, cmkId string) ybmclient.ApiUpdateClusterCmkStateRequest {
	return a.ApiClient.ClusterApi.UpdateClusterCmkState(a.ctx, a.AccountID, a.ProjectID, clusterId, cmkId)
}

func (a *AuthApiClient) ListBackups() ybmclient.ApiListBackupsRequest {
	return a.ApiClient.BackupApi.ListBackups(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListBackupPolicies(clusterId string, fetchOnlyActive bool) ybmclient.ApiListBackupSchedulesRequest {
	if fetchOnlyActive {
		return a.ApiClient.BackupApi.ListBackupSchedules(a.ctx, a.AccountID, a.ProjectID).EntityId(clusterId).State("ACTIVE")
	}
	return a.ApiClient.BackupApi.ListBackupSchedules(a.ctx, a.AccountID, a.ProjectID).EntityId(clusterId)
}

func (a *AuthApiClient) UpdateBackupPolicy(schedulId string) ybmclient.ApiModifyBackupScheduleRequest {
	return a.ApiClient.BackupApi.ModifyBackupSchedule(a.ctx, a.AccountID, a.ProjectID, schedulId)
}

func (a *AuthApiClient) ListBackupPoliciesV2(clusterId string, fetchOnlyActive bool) ybmclient.ApiListBackupSchedulesV2Request {
	if fetchOnlyActive {
		return a.ApiClient.BackupApi.ListBackupSchedulesV2(a.ctx, a.AccountID, a.ProjectID, clusterId).State("ACTIVE")
	}
	return a.ApiClient.BackupApi.ListBackupSchedulesV2(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) UpdateBackupPolicyV2(clusterId, scheduleId string) ybmclient.ApiModifyBackupScheduleV2Request {
	return a.ApiClient.BackupApi.ModifyBackupScheduleV2(a.ctx, a.AccountID, a.ProjectID, clusterId, scheduleId)
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
		// Temporary backwards compatibility between Stable and Production tracks.
		if track.Spec.GetName() == trackName || (track.Spec.GetName() == "Production" && trackName == "Stable") {
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

func (a *AuthApiClient) GetSupportedNodeConfigurationsV2(cloud string, tier string, regions []string, geoPartitioned bool) map[string][]ybmclient.NodeConfigurationResponseItem {
	// For single-region clusters, set isMultiRegion = false.
	// For azure clusters, set isMultiRegion = false.
	// For geo clusters if FF is enabled, set isMultiRegion = false
	// For all other clusters, set isMultiRegion = true
	isMultiRegion := true
	if len(regions) == 1 || cloud == "AZURE" || (geoPartitioned) {
		isMultiRegion = false
	}
	instanceResp, resp, err := a.ApiClient.ClusterApi.GetSupportedNodeConfigurations(a.ctx).AccountId(a.AccountID).Cloud(cloud).Tier(tier).Regions(regions).IsMultiRegion(isMultiRegion).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(b)
		logrus.Fatalln(err)
	}
	return instanceResp.GetData()
}

func (a *AuthApiClient) GetSupportedNodeConfigurationsForEdit(clusterId string, regions []string) map[string][]ybmclient.NodeConfigurationResponseItem {
	instanceResp, resp, err := a.ApiClient.ClusterApi.GetSupportedNodeConfigurationsForClusterEdit(a.ctx, a.AccountID, a.ProjectID, clusterId).Regions(regions).PerRegion(true).ShowDisabled(false).ClusterType("PRIMARY").Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(b)
		logrus.Fatalln(err)
	}
	return instanceResp.GetData()
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

func (a *AuthApiClient) ListAllRbacRoles() ybmclient.ApiListRbacRolesRequest {
	return a.ApiClient.RoleApi.ListRbacRoles(a.ctx, a.AccountID).RoleTypes("ALL").Limit(100)
}

func (a *AuthApiClient) ListSystemRbacRoles() ybmclient.ApiListRbacRolesRequest {
	return a.ApiClient.RoleApi.ListRbacRoles(a.ctx, a.AccountID).RoleTypes("SYSTEM").Limit(100)
}

func (a *AuthApiClient) ListAllRbacRolesWithPermissions() ybmclient.ApiListRbacRolesRequest {
	return a.ApiClient.RoleApi.ListRbacRoles(a.ctx, a.AccountID).RoleTypes("ALL").Limit(100).IncludePermissions(true)
}

func (a *AuthApiClient) ListSystemRbacRolesWithPermissions() ybmclient.ApiListRbacRolesRequest {
	return a.ApiClient.RoleApi.ListRbacRoles(a.ctx, a.AccountID).RoleTypes("SYSTEM").Limit(100).IncludePermissions(true)
}

func (a *AuthApiClient) CreateRoleSpec(cmd *cobra.Command, name string, permissionsMap map[string][]string) (*ybmclient.RoleSpec, error) {

	rolePermissions := []ybmclient.ResourcePermissionInfo{}
	for resource, ops := range permissionsMap {
		operationGroups := []ybmclient.ResourceOperationGroup{}
		for _, op := range ops {
			operationGroups = append(operationGroups, *ybmclient.NewResourceOperationGroup(ybmclient.ResourceOperationGroupEnum(op)))

		}
		rolePermissions = append(rolePermissions, *ybmclient.NewResourcePermissionInfo(ybmclient.ResourceTypeEnum(resource), operationGroups))
	}

	roleSpec := ybmclient.NewRoleSpec(
		name,
		rolePermissions)

	if cmd.Flags().Changed("description") {
		description, _ := cmd.Flags().GetString("description")
		roleSpec.SetDescription(description)
	}

	return roleSpec, nil
}

func (a *AuthApiClient) CreateRole() ybmclient.ApiCreateRoleRequest {
	return a.ApiClient.RoleApi.CreateRole(a.ctx, a.AccountID)
}

func (a *AuthApiClient) UpdateRole(roleId string) ybmclient.ApiUpdateRoleRequest {
	return a.ApiClient.RoleApi.UpdateRole(a.ctx, a.AccountID, roleId)
}

func (a *AuthApiClient) DeleteRole(roleId string) ybmclient.ApiDeleteRoleRequest {
	return a.ApiClient.RoleApi.DeleteRole(a.ctx, a.AccountID, roleId)
}

func (a *AuthApiClient) GetRoleIdByName(roleName string) (string, error) {
	roleData, err := a.GetRoleByName(roleName)
	if err == nil {
		return roleData.Info.GetId(), nil
	}

	return "", fmt.Errorf("could not get role data for role name: %s", roleName)
}

func (a *AuthApiClient) GetRoleByName(roleName string) (ybmclient.RoleData, error) {
	roleResp, resp, err := a.ListAllRbacRoles().DisplayName(roleName).Execute()
	if err != nil {
		if strings.TrimSpace(GetApiErrorDetails(err)) == strings.TrimSpace(util.GetCustomRoleFeatureFlagDisabledError()) {
			systemRoleResponse, systemRoleResp, systemRoleErr := a.ListSystemRbacRoles().DisplayName(roleName).Execute()

			if systemRoleErr != nil {
				b, _ := httputil.DumpResponse(systemRoleResp, true)
				logrus.Debug(string(b))
				return ybmclient.RoleData{}, systemRoleErr
			} else {
				roleResp = systemRoleResponse
			}
		} else {
			c, _ := httputil.DumpResponse(resp, true)
			logrus.Debug(string(c))
			return ybmclient.RoleData{}, err
		}
	}

	roleData := roleResp.GetData()

	if len(roleData) != 0 {
		return roleData[0], nil
	}

	return ybmclient.RoleData{}, fmt.Errorf("could not get role data for role name: %s", roleName)
}

func (a *AuthApiClient) ListResourcePermissions() ybmclient.ApiListResourcePermissionsRequest {
	return a.ApiClient.AuthApi.ListResourcePermissions(a.ctx)
}

func (a *AuthApiClient) GetSensitivePermissions() (map[string][]string, error) {
	resourcePermissionsResp, resp, err := a.ListResourcePermissions().Execute()
	if err != nil {
		c, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(c)
		return nil, err
	}

	permissions := resourcePermissionsResp.GetData()
	sensitivePermissionsMap := map[string][]string{}
	for _, permission := range permissions {
		resourceType := string(permission.Info.GetResourceType())
		for i := 0; i < len(permission.Info.OperationGroups); i++ {
			operationGroup := string(permission.Info.OperationGroups[i].GetOperationGroup())
			tags := permission.Info.OperationGroups[i].GetTags()
			if len(tags) == 0 || tags[0] != "SENSITIVE" {
				continue
			}
			sensitivePermissionsMap[resourceType] = append(sensitivePermissionsMap[resourceType], operationGroup)
		}
	}
	return sensitivePermissionsMap, nil
}

func (a *AuthApiClient) ContainsSensitivePermissions(permissionsMap map[string][]string) (bool, error) {
	sensitivePermissions, err := a.GetSensitivePermissions()
	if err != nil {
		return false, err
	}
	for sensitiveResourceType, sensitiveOps := range sensitivePermissions {
		if ops, ok := permissionsMap[sensitiveResourceType]; ok {
			for _, op := range ops {
				for _, sensitiveOp := range sensitiveOps {
					if op == sensitiveOp {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

func (a *AuthApiClient) RoleContainsSensitivePermissions(roleId string) (bool, error) {
	sensitivePermissions, err := a.GetSensitivePermissions()
	if err != nil {
		return false, err
	}

	roleResp, _, err := a.GetRole(roleId).Execute()
	if err != nil {
		return false, err
	}

	rolePermissions := roleResp.Data.Info.GetEffectivePermissions()
	for _, rolePermission := range rolePermissions {
		roleResourceType := string(rolePermission.GetResourceType())
		for i := 0; i < len(rolePermission.OperationGroups); i++ {
			roleOp := string(rolePermission.OperationGroups[i].GetOperationGroup())
			if sensitiveOps, ok := sensitivePermissions[roleResourceType]; ok {
				for _, sensitiveOp := range sensitiveOps {
					if sensitiveOp == roleOp {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

func (a *AuthApiClient) GetRole(roleId string) ybmclient.ApiGetRoleRequest {
	return a.ApiClient.RoleApi.GetRole(a.ctx, a.AccountID, roleId)
}

func (a *AuthApiClient) ListApiKeys() ybmclient.ApiListApiKeysRequest {
	return a.ApiClient.AuthApi.ListApiKeys(a.ctx, a.AccountID)
}

func (a *AuthApiClient) RevokeApiKey(keyId string) ybmclient.ApiRevokeApiKeyRequest {
	return a.ApiClient.AuthApi.RevokeApiKey(a.ctx, a.AccountID, keyId)
}

func (a *AuthApiClient) CreateApiKeySpec(name string, expiryHours int) (*ybmclient.ApiKeySpec, error) {
	apiKeySpec := ybmclient.NewApiKeySpec(name, int32(expiryHours))
	return apiKeySpec, nil
}

func (a *AuthApiClient) CreateApiKey() ybmclient.ApiCreateApiKeyRequest {
	return a.ApiClient.AuthApi.CreateApiKey(a.ctx, a.AccountID)
}

func (a *AuthApiClient) GetKeyIdByName(name string) (string, error) {
	apiKeyData, err := a.GetApiKeyByName(name)
	if err == nil {
		return apiKeyData.Info.GetId(), nil
	}

	return "", fmt.Errorf("could not get API key data for name: %s", name)
}

func (a *AuthApiClient) GetApiKeyByName(name string) (ybmclient.ApiKeyData, error) {
	keyResp, resp, err := a.ListApiKeys().ApiKeyName(name).Execute()
	if err != nil {
		c, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(c)
		return ybmclient.ApiKeyData{}, err
	}

	keyData := keyResp.GetData()

	if len(keyData) != 0 {
		return keyData[0], nil
	}

	return ybmclient.ApiKeyData{}, fmt.Errorf("could not get API Key data for name: %s", name)
}

func (a *AuthApiClient) ListAccountUsers() ybmclient.ApiListAccountUsersRequest {
	return a.ApiClient.AccountApi.ListAccountUsers(a.ctx, a.AccountID)
}

func (a *AuthApiClient) CreateBatchInviteUserSpec(email string, roleId string) (*ybmclient.BatchInviteUserSpec, error) {
	users := []ybmclient.InviteUserSpec{}
	user := *ybmclient.NewInviteUserSpecWithDefaults()
	user.SetEmail(email)

	user.SetRoleId(roleId)

	users = append(users, user)

	usersSpec := *ybmclient.NewBatchInviteUserSpecWithDefaults()
	usersSpec.SetUserList(users)

	return &usersSpec, nil
}

func (a *AuthApiClient) BatchInviteAccountUsers() ybmclient.ApiBatchInviteAccountUsersRequest {
	return a.ApiClient.AccountApi.BatchInviteAccountUsers(a.ctx, a.AccountID)
}

func (a *AuthApiClient) ModifyUserRole(userId string) ybmclient.ApiModifyUserRoleRequest {
	return a.ApiClient.AccountApi.ModifyUserRole(a.ctx, a.AccountID, userId)
}

func (a *AuthApiClient) RemoveAccountUser(userId string) ybmclient.ApiRemoveAccountUserRequest {
	return a.ApiClient.AccountApi.RemoveAccountUser(a.ctx, a.AccountID, userId)
}

func (a *AuthApiClient) GetUserIdByEmail(email string) (string, error) {
	userData, err := a.GetUserByEmail(email)
	if err == nil {
		return userData.Info.GetId(), nil
	}

	return "", fmt.Errorf("could not get user data for email: %s", email)
}

func (a *AuthApiClient) GetUserByEmail(email string) (ybmclient.UserData, error) {
	userResp, resp, err := a.ListAccountUsers().Email(email).Execute()
	if err != nil {
		c, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(c)
		return ybmclient.UserData{}, err
	}

	userData := userResp.GetData()

	if len(userData) != 0 {
		return userData[0], nil
	}

	return ybmclient.UserData{}, fmt.Errorf("could not get user data for email: %s", email)
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
	checkEveryInSec := time.Tick(10 * time.Second)
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
				return "", fmt.Errorf(GetApiErrorDetails(err))
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
				} else {
					currentStatus = "SUCCEEDED"
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
	checkEveryInSec := time.Tick(10 * time.Second)

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
				return "", fmt.Errorf(GetApiErrorDetails(err))
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
				} else {
					currentStatus = "SUCCEEDED"
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
				if slices.Contains([]string{"JWT has expired", "Invalid JWT"}, d.GetDetail()) && d.GetStatus() == http.StatusUnauthorized {
					return fmt.Sprintf("%s. Please run \"ybm auth\" again and provide a new API key", d.GetDetail())
				}
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
		logrus.Warnf("you are using insecure api endpoint %s\n", host)
	} else if !strings.HasPrefix(strings.ToLower(host), "https://") {
		host = "https://" + host
	}

	endpoint, err := url.ParseRequestURI(host)
	if err != nil {
		return nil, fmt.Errorf("could not parse ybm server url (%s): %w", host, err)
	}
	return endpoint, err
}

func (a *AuthApiClient) CreateMetricsExporterConfig() ybmclient.ApiCreateMetricsExporterConfigRequest {
	return a.ApiClient.MetricsExporterConfigApi.CreateMetricsExporterConfig(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) CreateIntegration() ybmclient.ApiCreateTelemetryProviderRequest {
	return a.ApiClient.TelemetryProviderApi.CreateTelemetryProvider(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ValidateIntegration() ybmclient.ApiValidateTelemetryProviderRequest {
	return a.ApiClient.TelemetryProviderApi.ValidateTelemetryProvider(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListMetricsExporterConfigs() ybmclient.ApiListMetricsExporterConfigsRequest {
	return a.ApiClient.MetricsExporterConfigApi.ListMetricsExporterConfigs(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListIntegrations() ybmclient.ApiListTelemetryProvidersRequest {
	return a.ApiClient.TelemetryProviderApi.ListTelemetryProviders(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListDbAuditExporterConfig(clusterId string) ybmclient.ApiListDbAuditExporterConfigRequest {
	return a.ApiClient.ClusterApi.ListDbAuditExporterConfig(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) UnassignDbAuditLogsExportConfig(configId string, integrationId string) ybmclient.ApiRemoveDbAuditLogExporterConfigRequest {
	return a.ApiClient.ClusterApi.RemoveDbAuditLogExporterConfig(a.ctx, a.AccountID, a.ProjectID, configId, integrationId)
}

func (a *AuthApiClient) DeleteMetricsExporterConfig(configId string) ybmclient.ApiDeleteMetricsExporterConfigRequest {
	return a.ApiClient.MetricsExporterConfigApi.DeleteMetricsExporterConfig(a.ctx, a.AccountID, a.ProjectID, configId)
}

func (a *AuthApiClient) DeleteIntegration(configId string) ybmclient.ApiDeleteTelemetryProviderRequest {
	return a.ApiClient.TelemetryProviderApi.DeleteTelemetryProvider(a.ctx, a.AccountID, a.ProjectID, configId)
}

func (a *AuthApiClient) RemoveMetricsExporterConfigFromCluster(clusterId string) ybmclient.ApiRemoveMetricsExporterConfigFromClusterRequest {
	return a.ApiClient.MetricsExporterConfigApi.RemoveMetricsExporterConfigFromCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) AssociateMetricsExporterWithCluster(clusterId string) ybmclient.ApiAddMetricsExporterConfigToClusterRequest {
	return a.ApiClient.MetricsExporterConfigApi.AddMetricsExporterConfigToCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) StopMetricsExporter(clusterId string) ybmclient.ApiStopMetricsExporterRequest {
	return a.ApiClient.MetricsExporterConfigApi.StopMetricsExporter(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) UpdateMetricsExporterConfig(configId string) ybmclient.ApiUpdateMetricsExporterConfigRequest {
	return a.ApiClient.MetricsExporterConfigApi.UpdateMetricsExporterConfig(a.ctx, a.AccountID, a.ProjectID, configId)
}

func (a *AuthApiClient) UpdateIntegration(configId string) ybmclient.ApiUpdateTelemetryProviderRequest {
	return a.ApiClient.TelemetryProviderApi.UpdateTelemetryProvider(a.ctx, a.AccountID, a.ProjectID, configId)
}

func (a *AuthApiClient) GetConfigByName(configName string) (*ybmclient.MetricsExporterConfigurationData, error) {
	resp, r, err := a.ListMetricsExporterConfigs().Execute()

	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		return nil, err
	}

	for _, metricsExporter := range resp.Data {
		if metricsExporter.GetSpec().Name == configName {
			return &metricsExporter, nil
		}
	}

	return nil, fmt.Errorf("could not find config with name %s", configName)
}

func (a *AuthApiClient) GetIntegrationByName(configName string) (*ybmclient.TelemetryProviderData, error) {
	resp, r, err := a.ListIntegrations().Execute()

	if err != nil {
		logrus.Debugf("Full HTTP response: %v", r)
		return nil, err
	}

	for _, tp := range resp.Data {
		if tp.GetSpec().Name == configName {
			return &tp, nil
		}
	}

	return nil, fmt.Errorf("could not find config with name %s", configName)
}

func (a *AuthApiClient) GetClusterNamespaces(clusterID string) ybmclient.ApiGetClusterNamespacesRequest {
	return a.ApiClient.ClusterApi.GetClusterNamespaces(a.ctx, a.AccountID, a.ProjectID, clusterID)
}

func (a *AuthApiClient) ListClusterPitrConfigs(clusterId string) ybmclient.ApiListClusterPitrConfigsRequest {
	return a.ApiClient.ClusterApi.ListClusterPitrConfigs(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) CreatePitrConfigSpec(databaseName string, databaseType string, retentionPeriodInDays int32) (*ybmclient.DatabasePitrConfigSpec, error) {
	pitrConfigSpec := ybmclient.NewDatabasePitrConfigSpec(ybmclient.YbApiEnum(databaseType), databaseName, retentionPeriodInDays)
	return pitrConfigSpec, nil
}

func (a *AuthApiClient) CreatePitrConfig(clusterId string) ybmclient.ApiCreateDatabasePitrConfigRequest {
	return a.ApiClient.ClusterApi.CreateDatabasePitrConfig(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) CreateRestoreViaPitrConfigSpec(restoreAtMilis int64) (*ybmclient.DatabaseRestoreViaPitrSpec, error) {
	restoreViaPitrConfigSpec := ybmclient.NewDatabaseRestoreViaPitrSpec(restoreAtMilis)
	return restoreViaPitrConfigSpec, nil
}

func (a *AuthApiClient) RestoreViaPitrConfig(clusterId string, pitrConfigId string) ybmclient.ApiRestoreDatabaseViaPitrRequest {
	return a.ApiClient.ClusterApi.RestoreDatabaseViaPitr(a.ctx, a.AccountID, a.ProjectID, clusterId, pitrConfigId)
}

func (a *AuthApiClient) GetPitrConfig(clusterId string, pitrConfigId string) ybmclient.ApiGetDatabasePitrConfigRequest {
	return a.ApiClient.ClusterApi.GetDatabasePitrConfig(a.ctx, a.AccountID, a.ProjectID, clusterId, pitrConfigId)
}

func (a *AuthApiClient) DeletePitrConfig(clusterId string, pitrConfigId string) ybmclient.ApiRemoveDatabasePitrConfigRequest {
	return a.ApiClient.ClusterApi.RemoveDatabasePitrConfig(a.ctx, a.AccountID, a.ProjectID, clusterId, pitrConfigId)
}

func (a *AuthApiClient) PerformConnectionPoolingOperation(clusterId string) ybmclient.ApiPerformConnectionPoolingOperationRequest {
	return a.ApiClient.ClusterApi.PerformConnectionPoolingOperation(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) GetDbLoggingConfig(clusterId string) ybmclient.ApiListPgLogExporterConfigsRequest {
	return a.ApiClient.ClusterApi.ListPgLogExporterConfigs(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) EnableDbQueryLogging(clusterId string) ybmclient.ApiAssociatePgLogExporterConfigRequest {
	return a.ApiClient.ClusterApi.AssociatePgLogExporterConfig(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) EditDbQueryLoggingConfig(clusterId string, exporterConfigId string) ybmclient.ApiUpdatePgLogExporterConfigRequest {
	return a.ApiClient.ClusterApi.UpdatePgLogExporterConfig(a.ctx, a.AccountID, a.ProjectID, clusterId, exporterConfigId)
}

func (a *AuthApiClient) RemoveDbQueryLoggingConfig(clusterId string, exporterConfigId string) ybmclient.ApiRemovePgLogExporterConfigRequest {
	return a.ApiClient.ClusterApi.RemovePgLogExporterConfig(a.ctx, a.AccountID, a.ProjectID, clusterId, exporterConfigId)
}

func (authApi *AuthApiClient) GetIntegrationIdFromName(integrationName string) (string, error) {
	integration, _, err := authApi.ListIntegrations().Name(integrationName).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get integration by name %s: %w", integrationName, err)
	}

	integrationData := integration.GetData()

	if len(integrationData) == 0 {
		return "", fmt.Errorf("no integrations found with name: %s%s", integrationName, "\n")
	}

	return integrationData[0].GetInfo().Id, nil
}

func (a *AuthApiClient) CreateXClusterDr(clusterId string) ybmclient.ApiCreateXClusterDrRequest {
	return a.ApiClient.XclusterDrApi.CreateXClusterDr(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) GetXClusterDr(clusterId string, drId string) ybmclient.ApiGetXClusterDrRequest {
	return a.ApiClient.XclusterDrApi.GetXClusterDr(a.ctx, a.AccountID, a.ProjectID, clusterId, drId)
}

func (a *AuthApiClient) ListXClusterDr(clusterId string) ybmclient.ApiListXClusterDrRequest {
	return a.ApiClient.XclusterDrApi.ListXClusterDr(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) DeleteXClusterDr(clusterId string, drId string) ybmclient.ApiDeleteXClusterDrRequest {
	return a.ApiClient.XclusterDrApi.DeleteXClusterDr(a.ctx, a.AccountID, a.ProjectID, clusterId, drId)
}
