package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
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

// NewAuthClient function is returning a new AuthApiClient Client
func NewAuthApiClient() (*AuthApiClient, error) {
	configuration := ybmclient.NewConfiguration()
	//Configure the client

	url, err := parseURL(viper.GetString("host"))
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	configuration.Host = url.Host
	configuration.Scheme = url.Scheme
	apiClient := ybmclient.NewAPIClient(configuration)
	apiKey := viper.GetString("apiKey")
	apiClient.GetConfig().AddDefaultHeader("Authorization", "Bearer "+apiKey)
	apiClient.GetConfig().UserAgent = "ybm-cli/" + cliVersion

	return &AuthApiClient{
		apiClient,
		"",
		"",
		context.Background(),
	}, nil
}

func (a *AuthApiClient) ListAccounts() ybmclient.ApiListAccountsRequest {
	return a.ApiClient.AccountApi.ListAccounts(a.ctx)
}

func (a *AuthApiClient) ListProjects() ybmclient.ApiListProjectsRequest {
	return a.ApiClient.ProjectApi.ListProjects(a.ctx, a.AccountID)
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

	clusterRegionInfo := []ybmclient.ClusterRegionInfo{}
	totalNodes := 0
	for _, regionInfo := range regionInfoList {
		numNodes, _ := strconv.ParseInt(regionInfo["num_nodes"], 10, 32)
		regionNodes := int32(numNodes)
		region := regionInfo["region"]
		totalNodes += int(regionNodes)
		cloudInfo := *ybmclient.NewCloudInfoWithDefaults()
		cloudInfo.SetRegion(region)
		if cmd.Flags().Changed("cloud-type") {
			cloudType, _ := cmd.Flags().GetString("cloud-type")
			cloudInfo.SetCode(ybmclient.CloudEnum(cloudType))
		}
		info := *ybmclient.NewClusterRegionInfo(
			*ybmclient.NewPlacementInfo(cloudInfo, int32(regionNodes)),
		)
		if vpcID, ok := regionInfo["vpc_id"]; ok {
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
	cloudInfo := *ybmclient.NewCloudInfoWithDefaults()
	if cmd.Flags().Changed("cloud-type") {
		cloudType, _ := cmd.Flags().GetString("cloud-type")
		cloudInfo.SetCode(ybmclient.CloudEnum(cloudType))
	}
	if regionInfoProvided {
		cloudInfo.SetRegion(region)
	}

	clusterInfo := *ybmclient.NewClusterInfoWithDefaults()
	if cmd.Flags().Changed("cluster-tier") {
		clusterTier, _ := cmd.Flags().GetString("cluster-tier")
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
		numCores := nodeConfig["num_cores"]

		clusterInfo.NodeInfo.SetNumCores(int32(numCores))

		if diskSize, ok := nodeConfig["disk_size_gb"]; ok {
			diskSizeGb = int32(diskSize)
		}

	}

	cloud := string(cloudInfo.GetCode())
	region = cloudInfo.GetRegion()
	tier := string(clusterInfo.GetClusterTier())
	numCores := clusterInfo.NodeInfo.GetNumCores()

	memoryMb, err := a.GetFromInstanceType("memory", cloud, tier, region, int32(numCores))
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
	if cmd.Flags().Changed("database-track") {
		trackName, _ = cmd.Flags().GetString("database-track")
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
	if regionInfoProvided {
		clusterSpec.SetClusterRegionInfo(clusterRegionInfo)
	}

	return clusterSpec, nil
}

func (a *AuthApiClient) GetInfo(providedAccountID string, providedProjectID string) {
	var err error
	a.AccountID, err = a.GetAccountID(providedAccountID)
	if err != nil {
		logrus.Errorf("could not initiate api client: ", err.Error())
		os.Exit(1)
	}
	a.ProjectID, err = a.GetProjectID(providedProjectID)
	if err != nil {
		logrus.Errorf("could not initiate api client: ", err.Error())
		os.Exit(1)
	}
}

func (a *AuthApiClient) GetClusterIdByName(clusterName string) (string, error) {
	clusterResp, resp, err := a.ListClusters().Name(clusterName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return "", err
	}
	clusterData := clusterResp.GetData()

	if len(clusterData) != 0 {
		return clusterData[0].Info.GetId(), nil
	}

	return "", fmt.Errorf("could no get cluster data for cluster name: %s", clusterName)
}

func (a *AuthApiClient) CreateCluster() ybmclient.ApiCreateClusterRequest {
	return a.ApiClient.ClusterApi.CreateCluster(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) GetCluster(clusterId string) ybmclient.ApiGetClusterRequest {
	return a.ApiClient.ClusterApi.GetCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
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

func (a *AuthApiClient) GetSupportedInstanceTypes(cloud string, tier string, region string, numCores int32) ybmclient.ApiGetSupportedInstanceTypesRequest {
	return a.ApiClient.ClusterApi.GetSupportedInstanceTypes(a.ctx).AccountId(a.AccountID).Cloud(cloud).Tier(tier).Region(region)
}
func (a *AuthApiClient) GetFromInstanceType(resource string, cloud string, tier string, region string, numCores int32) (int32, error) {
	instanceResp, resp, err := a.GetSupportedInstanceTypes(cloud, tier, region, numCores).Execute()
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

func parseURL(host string) (*url.URL, error) {
	endpoint, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("could not parse ybm server url (%s): %w", host, err)
	}
	if endpoint.Scheme == "" {
		endpoint.Scheme = "https"
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
