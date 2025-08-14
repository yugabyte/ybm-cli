package billing

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
)

var BillingCmd = &cobra.Command{
	Use:   "billing",
	Short: "Billing operations for YugabyteDB Aeon",
	Long:  "Billing operations for YugabyteDB Aeon",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var billingEstimateCmd = &cobra.Command{
	Use:   "estimate",
	Short: "Get billing estimate for accounts",
	Long:  "Get billing estimate for one or more accounts within a specified date range. Results may not reflect real-time usage, please verify against your end-of-month invoice for accurate billing.",
	Run: func(cmd *cobra.Command, args []string) {
		authApi, err := ybmAuthClient.NewAuthApiClient()
		if err != nil {
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}
		authApi.GetInfo("", "")

		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		accountNames, _ := cmd.Flags().GetStringSlice("account-names")

		resp, r, err := authApi.GetBillingEstimate(startDate, endDate, accountNames).Execute()
		if err != nil {
			logrus.Debugf("Full HTTP response: %v", r)
			logrus.Fatalf(ybmAuthClient.GetApiErrorDetails(err))
		}

		billingEstimateData := resp.GetData()
		formatter.BillingEstimateWriteFull(billingEstimateData)
	},
}

func init() {
	BillingCmd.AddCommand(billingEstimateCmd)

	billingEstimateCmd.Flags().StringSlice("account-names", []string{}, "[OPTIONAL] Comma-separated list of account names to fetch billing information for. Defaults to all accounts of user.")
	billingEstimateCmd.Flags().String("start-date", "", "[OPTIONAL] Start date(format yyyy-MM-dd) for billing estimate (inclusive). Defaults to 1st day of current month if not provided.")
	billingEstimateCmd.Flags().String("end-date", "", "[OPTIONAL] End date(format yyyy-MM-dd) for billing estimate (inclusive). Defaults to current date if not provided.")
}
