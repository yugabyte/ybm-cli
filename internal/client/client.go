package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	openapi "github.com/yugabyte/yugabytedb-managed-go-client-internal"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

// AuthApiClient is a auth YBM Client
type AuthApiClient struct {
	ApiClient *ybmclient.APIClient
	AccountID string
	ProjectID string
	ctx       context.Context
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
	return &AuthApiClient{
		apiClient,
		"",
		"",
		context.Background(),
	}, nil
}

func (a *AuthApiClient) GetAccountID(accountID string) (string, error) {
	//If an account ID is provided then we use this one
	if len(accountID) > 0 {
		return accountID, nil
	}
	accountResp, resp, err := a.ApiClient.AccountApi.ListAccounts(a.ctx).Execute()
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

func (a *AuthApiClient) GetProjectID(projectID string, providedAccountID string) (string, error) {
	// If a projectID is specified then we use this one.
	if len(projectID) > 0 {
		return projectID, nil
	}
	accountId, err := a.GetAccountID(providedAccountID)
	if err != nil {
		return "", err
	}

	projectResp, resp, err := a.ApiClient.ProjectApi.ListProjects(a.ctx, accountId).Execute()
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

func (a *AuthApiClient) GetInfo(providedAccountID string, providedProjectID string) {
	var err error
	a.AccountID, err = a.GetAccountID(providedAccountID)
	if err != nil {
		logrus.Errorf("could not initiate api client: ", err.Error())
		os.Exit(1)
	}
	a.ProjectID, err = a.GetProjectID(providedProjectID, a.AccountID)
	if err != nil {
		logrus.Errorf("could not initiate api client: ", err.Error())
		os.Exit(1)
	}
}

func (a *AuthApiClient) GetClusterID(clusterName string) (string, error) {
	clusterResp, resp, err := a.ApiClient.ClusterApi.ListClusters(a.ctx, a.AccountID, a.ProjectID).Name(clusterName).Execute()
	if err != nil {
		b, _ := httputil.DumpResponse(resp, true)
		logrus.Debug(string(b))
		return "", fmt.Errorf("could not find cluster id with name: %s", clusterName)
	}
	clusterData := clusterResp.GetData()

	if len(clusterData) != 0 {
		return clusterData[0].Info.GetId(), nil
	}

	return "", fmt.Errorf("could no get cluster data for cluster name: %s", clusterName)
}

func (a *AuthApiClient) ListClusters() openapi.ApiListClustersRequest {
	return a.ApiClient.ClusterApi.ListClusters(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) DeleteCluster(clusterId string) openapi.ApiDeleteClusterRequest {
	return a.ApiClient.ClusterApi.DeleteCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) PauseCluster(clusterId string) openapi.ApiPauseClusterRequest {
	return a.ApiClient.ClusterApi.PauseCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) ResumeCluster(clusterId string) openapi.ApiResumeClusterRequest {
	return a.ApiClient.ClusterApi.ResumeCluster(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) CreateReadReplica(clusterId string) openapi.ApiCreateReadReplicaRequest {
	return a.ApiClient.ReadReplicaApi.CreateReadReplica(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) ListReadReplicas(clusterId string) openapi.ApiListReadReplicasRequest {
	return a.ApiClient.ReadReplicaApi.ListReadReplicas(a.ctx, a.AccountID, a.ProjectID, clusterId)
}

func (a *AuthApiClient) ListSingleTenantVpcs() openapi.ApiListSingleTenantVpcsRequest {
	return a.ApiClient.NetworkApi.ListSingleTenantVpcs(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListSingleTenantVpcsByName(name string) openapi.ApiListSingleTenantVpcsRequest {
	if name == "" {
		return a.ListSingleTenantVpcs()
	}
	return a.ListSingleTenantVpcs().Name(name)
}

func (a *AuthApiClient) DeleteVpc(vpcId string) openapi.ApiDeleteVpcRequest {
	return a.ApiClient.NetworkApi.DeleteVpc(a.ctx, a.AccountID, a.ProjectID, vpcId)
}

func (a *AuthApiClient) CreateVpcPeering() openapi.ApiCreateVpcPeeringRequest {
	return a.ApiClient.NetworkApi.CreateVpcPeering(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListVpcPeerings() openapi.ApiListVpcPeeringsRequest {
	return a.ApiClient.NetworkApi.ListVpcPeerings(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) DeleteVpcPeering(vpcPeeringId string) openapi.ApiDeleteVpcPeeringRequest {
	return a.ApiClient.NetworkApi.DeleteVpcPeering(a.ctx, a.AccountID, a.ProjectID, vpcPeeringId)
}

func (a *AuthApiClient) CreateNetworkAllowList() openapi.ApiCreateNetworkAllowListRequest {
	return a.ApiClient.NetworkApi.CreateNetworkAllowList(a.ctx, a.AccountID, a.ProjectID)
}
func (a *AuthApiClient) DeleteNetworkAllowList(allowListId string) openapi.ApiDeleteNetworkAllowListRequest {
	return a.ApiClient.NetworkApi.DeleteNetworkAllowList(a.ctx, a.AccountID, a.ProjectID, allowListId)
}
func (a *AuthApiClient) ListNetworkAllowLists() openapi.ApiListNetworkAllowListsRequest {
	return a.ApiClient.NetworkApi.ListNetworkAllowLists(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) ListBackups() openapi.ApiListBackupsRequest {
	return a.ApiClient.BackupApi.ListBackups(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) RestoreBackup() openapi.ApiRestoreBackupRequest {
	return a.ApiClient.BackupApi.RestoreBackup(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) CreateBackup() openapi.ApiCreateBackupRequest {
	return a.ApiClient.BackupApi.CreateBackup(a.ctx, a.AccountID, a.ProjectID)
}

func (a *AuthApiClient) DeleteBackup(backupId string) openapi.ApiDeleteBackupRequest {
	return a.ApiClient.BackupApi.DeleteBackup(a.ctx, a.AccountID, a.ProjectID, backupId)
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
