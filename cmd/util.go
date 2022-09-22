package cmd

import (
	"context"
	"fmt"
	"net/http/httputil"
	"os"
	"strconv"

	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func prettyPrintJson(data interface{}) {
	//b, _ := json.MarshalIndent(data, "", "  ")
	b, _ := prettyjson.Marshal(data)
	fmt.Println(string(b))
}

func getHostOrDefault(ctx context.Context) string {
	host := os.Getenv("YBM_HOST")
	if host == "" {
		host = "devcloud.yugabyte.com"
	}
	return host
}

func getApiClient(ctx context.Context) (*ybmclient.APIClient, error) {
	configuration := ybmclient.NewConfiguration()
	//Configure the client

	configuration.Host = getHostOrDefault(ctx)
	configuration.Scheme = "https"
	apiClient := ybmclient.NewAPIClient(configuration)
	// authorize user with api key
	apiKeyBytes, _ := os.ReadFile("credentials")
	apiKey := string(apiKeyBytes)
	apiClient.GetConfig().AddDefaultHeader("Authorization", "Bearer "+apiKey)
	return apiClient, nil
}

func getClusterID(ctx context.Context, apiClient *ybmclient.APIClient, accountId string, projectId string, clusterName string) (clusterId string, clusterIdOk bool, errorMessage string) {
	clusterResp, resp, err := apiClient.ClusterApi.ListClusters(ctx, accountId, projectId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return "", false, string(b)
	}
	clusterData := clusterResp.GetData()

	for _, cluster := range clusterData {
		if cluster.Spec.GetName() == clusterName {
			return cluster.Info.GetId(), true, ""
		}
	}

	return "", false, "Couldn't find any cluster with the given name"
}

func getAccountID(ctx context.Context, apiClient *ybmclient.APIClient) (accountId string, accountIdOK bool, errorMessage string) {
	accountResp, resp, err := apiClient.AccountApi.ListAccounts(ctx).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return "", false, string(b)
	}
	accountData := accountResp.GetData()
	if len(accountData) == 0 {
		return "", false, "The user is not associated with any accounts."
	}
	if len(accountData) > 1 {
		return "", false, "The user is associated with multiple accounts, please provide an account ID."
	}
	accountId = accountData[0].Info.Id
	return accountId, true, ""
}

func getProjectID(ctx context.Context, apiClient *ybmclient.APIClient, accountId string) (projectId string, projectIdOK bool, errorMessage string) {
	projectResp, resp, err := apiClient.ProjectApi.ListProjects(ctx, accountId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return "", false, string(b)
	}
	projectData := projectResp.GetData()
	if len(projectData) == 0 {
		return "", false, "The account is not associated with any projects."
	}
	if len(projectData) > 1 {
		return "", false, "The account is associated with multiple projects, please provide a project ID."
	}

	projectId = projectData[0].Id
	return projectId, true, ""
}

func createClusterSpec(ctx context.Context, apiClient *ybmclient.APIClient, cmd *cobra.Command, accountId string, regionInfoList []map[string]string) (clusterSpec *ybmclient.ClusterSpec, clusterSpecOK bool, errorMessage string) {

	var diskSizeGb int32
	var diskSizeOK bool
	var memoryMb int32
	var memoryOK bool
	var trackId string
	var trackName string
	var trackIdOK bool
	var message string
	var regionInfoProvided bool

	clusterRegionInfo := []ybmclient.ClusterRegionInfo{}
	totalNodes := 0
	fmt.Println(regionInfoList)
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

	memoryMb, memoryOK, message = getMemoryFromInstanceType(ctx, apiClient, accountId, cloud, tier, region, int32(numCores))
	if !memoryOK {
		return nil, false, message
	}
	clusterInfo.NodeInfo.SetMemoryMb(memoryMb)

	// Computing the default disk size if it is not provided
	if diskSizeGb == 0 {
		diskSizeGb, diskSizeOK, message = getDiskSizeFromInstanceType(ctx, apiClient, accountId, cloud, tier, region, int32(numCores))
		if !diskSizeOK {
			return nil, false, message
		}
	}
	clusterInfo.NodeInfo.SetDiskSizeGb(diskSizeGb)

	if cmd.Flags().Changed("cluster-type") {
		clusterType, _ := cmd.Flags().GetString("cluster-type")
		clusterInfo.SetClusterType(ybmclient.ClusterType(clusterType))
	}
	networking := *ybmclient.NewNetworkingWithDefaults()

	// Compute track ID for database version
	softwareInfo := *ybmclient.NewSoftwareInfoWithDefaults()
	if cmd.Flags().Changed("database-track") {
		trackName, _ = cmd.Flags().GetString("database-track")
		trackId, trackIdOK, message = getTrackId(ctx, apiClient, accountId, trackName)
		if !trackIdOK {
			return nil, false, message
		}
		softwareInfo.SetTrackId(trackId)
	}

	clusterSpec = ybmclient.NewClusterSpec(
		clusterName,
		cloudInfo,
		clusterInfo,
		networking,
		softwareInfo)
	if regionInfoProvided {
		clusterSpec.SetClusterRegionInfo(clusterRegionInfo)
	}

	return clusterSpec, true, ""
}

func getTrackId(ctx context.Context, apiClient *ybmclient.APIClient, accountId string, trackName string) (trackId string, trackIdOK bool, errorMessage string) {
	tracksResp, resp, err := apiClient.SoftwareReleaseApi.ListTracks(ctx, accountId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return "", false, string(b)
	}
	tracksData := tracksResp.GetData()

	for _, track := range tracksData {
		if track.Spec.GetName() == trackName {
			return track.Info.GetId(), true, ""
		}
	}

	return "", false, "The database version doesn't exist."
}

func getMemoryFromInstanceType(ctx context.Context, apiClient *ybmclient.APIClient, accountId string, cloud string, tier string, region string, numCores int32) (memory int32, memoryOK bool, errorMessage string) {
	instanceResp, resp, err := apiClient.ClusterApi.GetSupportedInstanceTypes(context.Background()).AccountId(accountId).Cloud(cloud).Tier(tier).Region(region).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return 0, false, string(b)
	}
	instanceData := instanceResp.GetData()
	nodeConfigList, ok := instanceData[region]
	if !ok || len(nodeConfigList) == 0 {
		return 0, false, "No instances configured for the given region."
	}
	for _, nodeConfig := range nodeConfigList {
		if nodeConfig.GetNumCores() == numCores {
			memory = nodeConfig.GetMemoryMb()
			return memory, true, ""
		}
	}

	return 0, false, "Node with the given number of CPU cores doesn't exist in the given region."
}

func getDiskSizeFromInstanceType(ctx context.Context, apiClient *ybmclient.APIClient, accountId string, cloud string, tier string, region string, numCores int32) (diskSize int32, diskSizeOK bool, errorMessage string) {
	instanceResp, resp, err := apiClient.ClusterApi.GetSupportedInstanceTypes(context.Background()).AccountId(accountId).Cloud(cloud).Tier(tier).Region(region).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return 0, false, string(b)
	}
	instanceData := instanceResp.GetData()
	nodeConfigList, ok := instanceData[region]
	if !ok || len(nodeConfigList) == 0 {
		return 0, false, "No instances configured for the given region."
	}
	for _, nodeConfig := range nodeConfigList {
		if nodeConfig.GetNumCores() == numCores {
			diskSize = nodeConfig.GetIncludedDiskSizeGb()
			return diskSize, true, ""
		}
	}

	return 0, false, "Node with the given number of CPU cores doesn't exist in the given region."
}
