package cmd

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

	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

func prettyPrintJson(data interface{}) {
	//b, _ := json.MarshalIndent(data, "", "  ")
	b, _ := prettyjson.Marshal(data)
	fmt.Println(string(b))
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

func getApiClient(ctx context.Context, cmd *cobra.Command) (*ybmclient.APIClient, error) {
	configuration := ybmclient.NewConfiguration()
	//Configure the client

	url, err := parseURL(viper.GetString("host"))
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		os.Exit(1)
	}

	configuration.Host = url.Host
	configuration.Scheme = url.Scheme
	apiClient := ybmclient.NewAPIClient(configuration)
	apiKey := viper.GetString("apiKey")
	apiClient.GetConfig().AddDefaultHeader("Authorization", "Bearer "+apiKey)
	return apiClient, nil
}
func getClusterID(ctx context.Context, apiClient *ybmclient.APIClient, accountId string, projectId string, clusterName string) (clusterId string, clusterIdOk bool, errorMessage string) {
	clusterResp, resp, err := apiClient.ClusterApi.ListClusters(ctx, accountId, projectId).Name(clusterName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return "", false, string(b)
	}
	clusterData := clusterResp.GetData()

	if len(clusterData) != 0 {
		return clusterData[0].Info.GetId(), true, ""
	}

	return "", false, "Couldn't find any cluster with the given name"
}

func getAccountID(ctx context.Context, apiClient *ybmclient.APIClient) (accountId string, accountIdOK bool, errorMessage string) {
	accountResp, resp, err := apiClient.AccountApi.ListAccounts(ctx).Execute()
	if err != nil {
		errMsg := getErrorMessage(resp, err)
		if strings.Contains(err.Error(), "is not a valid") {
			fmt.Println("The deserialization of the response failed due to following error. "+
				"Skipping as this should not impact the functionality of the provider.",
				map[string]interface{}{"errMsg": err.Error()})
		} else {
			return "", false, errMsg
		}
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
		errMsg := getErrorMessage(resp, err)
		if strings.Contains(err.Error(), "is not a valid") {
			fmt.Println("The deserialization of the response failed due to following error. "+
				"Skipping as this should not impact the functionality of the provider.",
				map[string]interface{}{"errMsg": err.Error()})
		} else {
			return "", false, errMsg
		}
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

func getCdcSinkID(ctx context.Context, apiClient *ybmclient.APIClient, accountId string, cdcSinkName string) (sinkId string, sinkIdOk bool, errorMessage string) {
	sinkResp, resp, err := apiClient.CdcApi.ListCdcSinks(ctx, accountId).Name(cdcSinkName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return "", false, string(b)
	}

	sinkData := sinkResp.GetData()

	if len(sinkData) != 0 {
		return sinkData[0].Info.GetId(), true, ""
	}

	return "", false, "Couldn't find any cdcSink with the given name"
}

func getCdcStreamID(ctx context.Context, apiClient *ybmclient.APIClient, accountId string, cdcStreamName string) (streamId string, streamIdOk bool, errorMessage string) {
	streamResp, resp, err := apiClient.CdcApi.ListCdcStreamsForAccount(ctx, accountId).Name(cdcStreamName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return "", false, string(b)
	}
	streamData := streamResp.GetData()

	if len(streamData) != 0 {
		return streamData[0].Info.GetId(), true, ""
	}

	return "", false, "Couldn't find any cdcStream with the given name"
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
		clusterInfo,
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

func getTrackName(ctx context.Context, apiClient *ybmclient.APIClient, accountId string, trackId string) (trackName string, trackNameOK bool, errorMessage string) {

	trackNameResp, resp, err := apiClient.SoftwareReleaseApi.GetTrack(ctx, accountId, trackId).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		return "", false, string(b)
	}
	trackData := trackNameResp.GetData()
	trackName = trackData.Spec.GetName()

	return trackName, true, ""
}
func parseURL(host string) (*url.URL, error) {
	endpoint, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("Could not parse ybm server url (%s): %w", host, err)
	}
	if endpoint.Scheme == "" {
		endpoint.Scheme = "https"
	}
	return endpoint, err
}
