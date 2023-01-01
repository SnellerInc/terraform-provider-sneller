package sneller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

func (c *Client) client() *http.Client {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	return client
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
	resp, err := c.client().Do(req)
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

	resp, err := c.client().Do(req)
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

type User struct {
	UserID      string
	Email       string
	IsEnabled   bool
	IsFederated bool
	Locale      string
	GivenName   string
	FamilyName  string
	Picture     string
	Groups      []string // only returned when fetching user details
}

func (c *Client) Users(ctx context.Context) ([]User, error) {
	resp, err := c.client().Do(c.url(ctx, http.MethodGet, "", "/user"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) User(ctx context.Context, userID string) (*User, error) {
	resp, err := c.client().Do(c.url(ctx, http.MethodGet, "", fmt.Sprintf("/user/%s", userID)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) CreateUser(ctx context.Context, email string, isAdmin bool, locale, givenName, familyName *string) (string, error) {
	req := c.url(ctx, http.MethodPost, "", "/user")
	q, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return "", err
	}
	q.Set("email", email)
	if isAdmin {
		q.Set("isAdmin", "true")
	}
	if locale != nil && *locale != "" {
		q.Set("locale", *locale)
	}
	if givenName != nil && *givenName != "" {
		q.Set("givenName", *givenName)
	}
	if familyName != nil && *familyName != "" {
		q.Set("familyName", *familyName)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	var userID string
	if err := json.NewDecoder(resp.Body).Decode(&userID); err != nil {
		return "", err
	}

	return userID, nil
}

func (c *Client) UpdateUser(ctx context.Context, userID string, email *string, isEnabled, isAdmin *bool, locale, givenName, familyName *string) error {
	req := c.url(ctx, http.MethodPatch, "", fmt.Sprintf("/user/%s", userID))
	q, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return err
	}
	if isEnabled != nil {
		q.Set("isEnabled", strconv.FormatBool(*isEnabled))
	}
	if isAdmin != nil {
		q.Set("isAdmin", strconv.FormatBool(*isAdmin))
	}
	if email != nil {
		q.Set("email", *email)
	}
	if locale != nil {
		q.Set("locale", *locale)
	}
	if givenName != nil {
		q.Set("givenName", *givenName)
	}
	if familyName != nil {
		q.Set("familyName", *familyName)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.client().Do(req)
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

func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	resp, err := c.client().Do(c.url(ctx, http.MethodDelete, "", fmt.Sprintf("/user/%s", userID)))
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

	resp, err := c.client().Do(req)
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

	resp, err := c.client().Do(req)
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
	resp, err := c.client().Do(c.url(ctx, http.MethodGet, region, "/db"))
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
	resp, err := c.client().Do(c.url(ctx, http.MethodGet, region, fmt.Sprintf("/db/%s/table", database)))
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
	Name         string           `json:"name"`
	Input        []TableInput     `json:"input"`
	Partitions   []TablePartition `json:"partitions,omitempty"`
	BetaFeatures []string         `json:"beta_features,omitempty"`
	SkipBackfill bool             `json:"skip_backfill,omitempty"`
}

type TableInput struct {
	Pattern string `json:"pattern"`
	Format  string `json:"format"`
}

type TablePartition struct {
	Field string `json:"field"`
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

func (c *Client) SetTable(ctx context.Context, region, database string, table TableDescription) error {
	req := c.url(ctx, http.MethodPut, region, fmt.Sprintf("/db/%s/table/%s/definition", database, table.Name))
	body, err := json.Marshal(table)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(body))

	resp, err := c.client().Do(req)
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

	resp, err := c.client().Do(req)
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
	resp, err := c.client().Do(c.url(ctx, http.MethodGet, region, fmt.Sprintf("/db/%s/table/%s/definition", database, table)))
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

func (c *Client) ElasticProxyConfig(ctx context.Context, region string) (*ElasticProxyConfig, error) {
	resp, err := c.client().Do(c.url(ctx, http.MethodGet, region, "/elasticproxy/config"))
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP status %d: %s", resp.StatusCode, msg)
	}

	var config ElasticProxyConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("error reading elastic proxy configuration: %s", err.Error())
	}

	return &config, nil
}

func (c *Client) SetElasticProxyConfig(ctx context.Context, region string, config ElasticProxyConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req := c.url(ctx, http.MethodPut, region, "/elasticproxy/config")
	req.Header.Add("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(data))
	resp, err := c.client().Do(req)
	if err != nil {
		return err
	}

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

func (c *Client) DeleteElasticProxyConfig(ctx context.Context, region string) error {
	resp, err := c.client().Do(c.url(ctx, http.MethodDelete, region, "/elasticproxy/config"))
	if err != nil {
		return err
	}

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
