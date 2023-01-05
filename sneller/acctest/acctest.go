package acctest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"terraform-provider-sneller/sneller/api"
	"terraform-provider-sneller/sneller/provider"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// test.ProviderConfig is a shared configuration to combine with the actual
	// test configuration so the Sneller client is properly configured.
	// It is also possible to use the SNELLER_TOKEN and SNELLER_REGION
	// environment variables instead, such as updating the Makefile and
	// running the testing through that tool.
	ProviderConfig = `provider "sneller" {}` + "\n"
)

var (
	SnellerAccountID string
	TenantAccountID  string
	SnellerTenantID  string

	Bucket1Name  string
	Bucket2Name  string
	Role1ARN     string
	Role2ARN     string
	DatabaseName string
	TableName    string

	TestAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
)

func init() {
	ctx := context.Background()

	token := os.Getenv(api.EnvSnellerToken)
	if token == "" {
		panic(fmt.Sprintf("%s must be set for acceptance tests", api.EnvSnellerToken))
	}

	apiEndPoint := os.Getenv(api.EnvSnellerApiEndpoint)
	if apiEndPoint == "" {
		apiEndPoint = api.DefaultApiEndPoint
	}

	if TenantAccountID == "" {
		TenantAccountID = os.Getenv("TENANT_ACCOUNT_ID")
		if TenantAccountID == "" {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				panic(fmt.Sprintf("TENANT_ACCOUNT_ID not set and cannot load AWS configuration: %s", err.Error()))
			}
			resp, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
			if err != nil {
				panic(fmt.Sprintf("TENANT_ACCOUNT_ID not set and cannot determine current AWS account ID: %s", err.Error()))
			}
			TenantAccountID = *resp.Account
		}
	}

	apiURL, err := url.Parse(apiEndPoint)
	if err != nil {
		panic(fmt.Sprintf("The Sneller API url %q is invalid: %s", apiEndPoint, err.Error()))
	}

	c := api.Client{
		Token:         token,
		ApiURL:        apiURL,
		DefaultRegion: api.DefaultSnellerRegion,
	}
	tenantInfo, err := c.Tenant(ctx, "")
	if err != nil {
		panic(fmt.Sprintf("Unable to fetch tenant information: %s", err.Error()))
	}
	if tenantInfo.TenantState != "active" {
		panic(fmt.Sprintf("Tenant %q has invalid tenant state (got: %s, expected: active)", tenantInfo.TenantID, tenantInfo.TenantState))
	}

	tenantRoleARN, err := arn.Parse(tenantInfo.TenantRoleArn)
	if err != nil {
		panic(fmt.Sprintf("Unable to parse role ARN %q: %s", tenantInfo.TenantRoleArn, err.Error()))
	}

	SnellerTenantID = tenantInfo.TenantID
	SnellerAccountID = tenantRoleARN.AccountID

	Bucket1Name = fmt.Sprintf("sneller-cache1-%s-%s", strings.ToLower(SnellerTenantID), TenantAccountID)
	Bucket2Name = fmt.Sprintf("sneller-cache2-%s-%s", strings.ToLower(SnellerTenantID), TenantAccountID)
	Role1ARN = fmt.Sprintf("arn:aws:iam::%s:role/role1-%s", TenantAccountID, strings.ToLower(SnellerTenantID))
	Role2ARN = fmt.Sprintf("arn:aws:iam::%s:role/role2-%s", TenantAccountID, strings.ToLower(SnellerTenantID))
	DatabaseName = "test-db"
	TableName = "test-table"

	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"sneller": providerserver.NewProtocol6WithError(provider.New()),
	}
}
