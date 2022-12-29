package sneller

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	initialized bool

	snellerAccountID string
	tenantAccountID  string
	snellerTenantID  string

	bucket1Name  string
	bucket2Name  string
	role1ARN     string
	role2ARN     string
	databaseName string
	tableName    string
)

func initVars(t *testing.T) {
	if !initialized {
		ctx := context.Background()

		token := os.Getenv(envSnellerToken)
		if token == "" {
			t.Fatalf("%s must be set for acceptance tests", envSnellerToken)
		}

		apiEndPoint := os.Getenv(envSnellerApiEndpoint)
		if apiEndPoint == "" {
			apiEndPoint = defaultApiEndPoint
		}

		if tenantAccountID == "" {
			tenantAccountID = os.Getenv("TENANT_ACCOUNT_ID")
			if tenantAccountID == "" {
				cfg, err := config.LoadDefaultConfig(ctx)
				if err != nil {
					t.Fatalf("TENANT_ACCOUNT_ID not set and cannot load AWS configuration: %s", err.Error())
				}
				resp, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
				if err != nil {
					t.Fatalf("TENANT_ACCOUNT_ID not set and cannot determine current AWS account ID: %s", err.Error())
				}
				tenantAccountID = *resp.Account
			}
		}

		apiURL, err := url.Parse(apiEndPoint)
		if err != nil {
			t.Fatalf("The Sneller API url %q is invalid: %s", apiEndPoint, err.Error())
		}

		c := Client{
			Token:         token,
			apiURL:        apiURL,
			DefaultRegion: defaultSnellerRegion,
		}
		tenantInfo, err := c.Tenant(ctx, "")
		if err != nil {
			t.Fatalf("Unable to fetch tenant information: %s", err.Error())
		}
		if tenantInfo.TenantState != "active" {
			t.Fatalf("Tenant %q has invalid tenant state (got: %s, expected: active)", tenantInfo.TenantID, tenantInfo.TenantState)
		}

		tenantRoleARN, err := arn.Parse(tenantInfo.TenantRoleArn)
		if err != nil {
			t.Fatalf("Unable to parse role ARN %q: %s", tenantInfo.TenantRoleArn, err.Error())
		}

		snellerTenantID = tenantInfo.TenantID
		snellerAccountID = tenantRoleARN.AccountID

		bucket1Name = fmt.Sprintf("sneller-cache1-%s-%s", strings.ToLower(snellerTenantID), tenantAccountID)
		bucket2Name = fmt.Sprintf("sneller-cache2-%s-%s", strings.ToLower(snellerTenantID), tenantAccountID)
		role1ARN = fmt.Sprintf("arn:aws:iam::%s:role/role1-%s", tenantAccountID, strings.ToLower(snellerTenantID))
		role2ARN = fmt.Sprintf("arn:aws:iam::%s:role/role2-%s", tenantAccountID, strings.ToLower(snellerTenantID))
		databaseName = "test-db"
		tableName = "test-table"

		initialized = true
	}

	t.Logf("using tenant %s and AWS account %s", snellerTenantID, snellerAccountID)
}
