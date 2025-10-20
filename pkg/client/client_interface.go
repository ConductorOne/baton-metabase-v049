package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type ClientService interface {
	ListDatabases(ctx context.Context) ([]*Database, *v2.RateLimitDescription, error)
	GetDBPermissions(ctx context.Context, dbID string) (map[string]map[string]*GroupPermission, *v2.RateLimitDescription, error)
	GetVersion(ctx context.Context) (*VersionInfo, *v2.RateLimitDescription, error)
	IsPaidPlan() bool
}
