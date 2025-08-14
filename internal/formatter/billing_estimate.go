package formatter

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ybmclient "github.com/yugabyte/yugabytedb-managed-go-client-internal"
)

const (
	defaultBillingEstimateListing = "table {{.AccountName}}\t{{.TotalAmount}}"
	billingSummaryListing         = "table {{.StartDate}}\t{{.EndDate}}\t{{.TotalAmount}}"
	accountNameHeader             = "Account Name"
	totalAmountHeader             = "Amount"
)

type BillingEstimateContext struct {
	HeaderContext
	Context
	c ybmclient.BillingEstimateAccountInfo
}

type BillingSummaryContext struct {
	HeaderContext
	Context
	data ybmclient.BillingEstimateData
}

func NewBillingEstimateFormat() Format {
	source := viper.GetString("output")
	switch source {
	case "table", "":
		format := defaultBillingEstimateListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewBillingSummaryFormat() Format {
	source := viper.GetString("output")
	switch source {
	case "table", "":
		format := billingSummaryListing
		return Format(format)
	default: // custom format or json or pretty
		return Format(source)
	}
}

func NewBillingEstimateContext() *BillingEstimateContext {
	billingEstimateCtx := BillingEstimateContext{}
	billingEstimateCtx.Header = SubHeaderContext{
		"AccountName": accountNameHeader,
		"TotalAmount": totalAmountHeader,
	}
	return &billingEstimateCtx
}

func NewBillingSummaryContext() *BillingSummaryContext {
	billingSummaryCtx := BillingSummaryContext{}
	billingSummaryCtx.Header = SubHeaderContext{
		"StartDate":   "Start Date",
		"EndDate":     "End Date",
		"TotalAmount": "Total Amount",
	}
	return &billingSummaryCtx
}

func billingSummaryWrite(ctx Context, billingEstimateData ybmclient.BillingEstimateData) error {
	render := func(format func(subContext SubContext) error) error {
		err := format(&BillingSummaryContext{data: billingEstimateData})
		if err != nil {
			logrus.Debugf("Error rendering billing summary: %v", err)
			return err
		}
		return nil
	}
	return ctx.Write(NewBillingSummaryContext(), render)
}

func billingEstimateAccountsWrite(ctx Context, accounts []ybmclient.BillingEstimateAccountInfo) error {
	render := func(format func(subContext SubContext) error) error {
		for _, account := range accounts {
			err := format(&BillingEstimateContext{c: account})
			if err != nil {
				logrus.Debugf("Error rendering billing estimate account: %v", err)
				return err
			}
		}
		return nil
	}
	return ctx.Write(NewBillingEstimateContext(), render)
}

func BillingEstimateWriteFull(billingEstimateData ybmclient.BillingEstimateData) {
	ctx := Context{
		Output: os.Stdout,
		Format: NewBillingSummaryFormat(),
	}

	err := billingSummaryWrite(ctx, billingEstimateData)
	if err != nil {
		logrus.Fatal(err.Error())
	}
	ctx.Output.Write([]byte("\n"))

	accounts := billingEstimateData.GetAccounts()
	if len(accounts) == 0 {
		if viper.GetString("output") == "table" {
			fmt.Fprintf(ctx.Output, "No account data available.\n")
		}
		return
	}

	if viper.GetString("output") == "table" {
		ctx = Context{
			Output: os.Stdout,
			Format: NewBillingEstimateFormat(),
		}

		err = billingEstimateAccountsWrite(ctx, accounts)
		if err != nil {
			logrus.Fatal(err.Error())
		}
	}
}

func (c *BillingEstimateContext) AccountName() string {
	return c.c.GetAccountName()
}

func (c *BillingEstimateContext) TotalAmount() string {
	return fmt.Sprintf("$%.2f", c.c.GetTotalAmount())
}

func (c *BillingEstimateContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.c)
}

func (c *BillingSummaryContext) StartDate() string {
	return c.data.GetStartDate()
}

func (c *BillingSummaryContext) EndDate() string {
	return c.data.GetEndDate()
}

func (c *BillingSummaryContext) TotalAmount() string {
	return fmt.Sprintf("$%.2f", c.data.GetTotalAmount())
}

func (c *BillingSummaryContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.data)
}
