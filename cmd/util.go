package cmd

import (
	"context"
	"fmt"
	"net/http/httputil"
	"os"

	"github.com/hokaccha/go-prettyjson"
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
