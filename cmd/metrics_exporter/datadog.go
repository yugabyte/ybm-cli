package metrics_exporter

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func ValidateDatadogApiKey(apiKey string, site string) (bool, string) {
	//For CI we will skip validation are we don't want to use a valid DD API KEY for you test
	if strings.ToLower(os.Getenv("YBM_CI")) == "true" {
		return true, ""
	}
	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: apiKey,
			},
		},
	)

	ctx = context.WithValue(ctx,
		datadog.ContextServerVariables,
		map[string]string{
			"site": site,
		})

	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewAuthenticationApi(apiClient)
	_, r, err := api.Validate(ctx)

	if r.StatusCode == http.StatusForbidden {
		return false, fmt.Sprintf("Datadog api key is not valid on %s : Forbidden", site)
	}

	if err != nil {
		return false, fmt.Sprintf("Datadog api key is not valid on %s :%s", site, GetApiErrorDetails(err))
	}

	return true, ""
}

func GetApiErrorDetails(err error) string {
	switch castedError := err.(type) {
	case datadog.GenericOpenAPIError:
		return fmt.Sprintf("%s%s", castedError.ErrorModel, "\n")
	}
	return err.Error()

}
