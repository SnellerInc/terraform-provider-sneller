package api

import "time"

type TenantRegionInfo struct {
	Bucket                string
	RegionRoleArn         string
	RegionExternalID      string
	SqsArn                string
	MaxScanBytes          *uint64
	EffectiveMaxScanBytes uint64
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
