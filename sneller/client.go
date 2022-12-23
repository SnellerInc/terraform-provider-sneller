package sneller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Client        *http.Client
	tenantID      string
	Token         string
	DefaultRegion string
	apiURL        *url.URL
}

func (c *Client) url(ctx context.Context, method, region, path string) *http.Request {
	effectiveRegion := region
	if effectiveRegion == "" {
		effectiveRegion = c.DefaultRegion
	}
	tenantID := c.tenantID
	if tenantID == "" {
		tenantID = "me"
	}
	url := *c.apiURL
	url.Host = strings.ReplaceAll(url.Host, "__REGION__", effectiveRegion)
	url.Path = fmt.Sprintf("/tenant/%s%s", tenantID, path)
	req, err := http.NewRequestWithContext(ctx, method, url.String(), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	return req
}

func (c *Client) Ping(ctx context.Context, region string) error {
	req := c.url(ctx, http.MethodGet, region, "")
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}
	return nil
}

type MfaRequirement string

const (
	MfaOff      = MfaRequirement("off")
	MfaOptional = MfaRequirement("optional")
	MfaRequired = MfaRequirement("required")
)

type TenantRegionInfo struct {
	Bucket           string
	RegionRoleArn    string
	RegionExternalID string
	SqsArn           string
}

type TenantInfo struct {
	TenantID      string
	TenantState   string
	TenantName    string
	HomeRegion    string
	Email         string
	TenantRoleArn string
	Mfa           MfaRequirement
	CreatedAt     time.Time
	ActivatedAt   *time.Time
	DeactivatedAt *time.Time
	Regions       map[string]TenantRegionInfo `json:",omitempty"`
}

func (c *Client) Tenant(ctx context.Context, region string) (*TenantInfo, error) {
	req := c.url(ctx, http.MethodGet, "", "")
	if region != "" {
		q, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			return nil, err
		}
		q.Set("regions", region)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	var tenantInfo TenantInfo
	if err := json.NewDecoder(resp.Body).Decode(&tenantInfo); err != nil {
		return nil, err
	}

	if tenantInfo.Regions == nil {
		tenantInfo.Regions = make(map[string]TenantRegionInfo)
	}

	return &tenantInfo, nil
}

func (c *Client) SetBucket(ctx context.Context, region, bucket, roleARN string) error {
	req := c.url(ctx, http.MethodPatch, region, "")

	q, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return err
	}
	q.Set("operation", "setBucket")
	q.Set("bucket", bucket)
	q.Set("roleArn", roleARN)
	req.URL.RawQuery = q.Encode()

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	return nil
}

func (c *Client) ResetBucket(ctx context.Context, region string) error {
	req := c.url(ctx, http.MethodPatch, region, "")

	q, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return err
	}
	q.Set("operation", "resetBucket")
	req.URL.RawQuery = q.Encode()

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	return nil
}

func (c *Client) Databases(ctx context.Context, region string) ([]string, error) {
	resp, err := c.Client.Do(c.url(ctx, http.MethodGet, region, "/db"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	var databases []string
	if err := json.NewDecoder(resp.Body).Decode(&databases); err != nil {
		return nil, err
	}

	return databases, nil
}

type TableInfo struct {
	Name          string
	HasDefinition bool
	HasIndex      bool
}

func (c *Client) Database(ctx context.Context, region, database string) ([]TableInfo, error) {
	resp, err := c.Client.Do(c.url(ctx, http.MethodGet, region, fmt.Sprintf("/db/%s/table", database)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("Database %q not found", database)
		}
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	var tables []TableInfo
	if err := json.NewDecoder(resp.Body).Decode(&tables); err != nil {
		return nil, err
	}

	return tables, nil
}

type TableDescription struct {
	Name  string       `json:"name"`
	Input []TableInput `json:"input"`
}

type TableInput struct {
	Pattern string `json:"pattern"`
	Format  string `json:"format"`
}

func (c *Client) SetTable(ctx context.Context, region, database, table string, input []TableInput) error {
	req := c.url(ctx, http.MethodPut, region, fmt.Sprintf("/db/%s/table/%s/definition", database, table))
	body, err := json.Marshal(TableDescription{
		Name:  table,
		Input: input,
	})
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	return nil
}

func (c *Client) DeleteTable(ctx context.Context, region, database, table string, all bool) error {
	req := c.url(ctx, http.MethodDelete, region, fmt.Sprintf("/db/%s/table/%s/definition", database, table))
	if all {
		q, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			return err
		}
		q.Set("all", "")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	return nil
}

func (c *Client) Table(ctx context.Context, region, database, table string) (*TableDescription, error) {
	resp, err := c.Client.Do(c.url(ctx, http.MethodGet, region, fmt.Sprintf("/db/%s/table/%s/definition", database, table)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("Table %q in database %q not found", table, database)
		}
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	var tableDescription TableDescription
	if err := json.NewDecoder(resp.Body).Decode(&tableDescription); err != nil {
		return nil, err
	}

	return &tableDescription, nil
}
