package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type MockService struct {
	ListDatabasesFunc    func(ctx context.Context) ([]*Database, *v2.RateLimitDescription, error)
	GetDBPermissionsFunc func(ctx context.Context, dbID string) (map[string]map[string]*GroupPermission, *v2.RateLimitDescription, error)
	GetVersionFunc       func(ctx context.Context) (*VersionInfo, *v2.RateLimitDescription, error)
	IsPaidPlanFunc       func() bool
}

func (m *MockService) ListDatabases(ctx context.Context) ([]*Database, *v2.RateLimitDescription, error) {
	return m.ListDatabasesFunc(ctx)
}

func (m *MockService) GetDBPermissions(ctx context.Context, dbID string) (map[string]map[string]*GroupPermission, *v2.RateLimitDescription, error) {
	return m.GetDBPermissionsFunc(ctx, dbID)
}

func (m *MockService) GetVersion(ctx context.Context) (*VersionInfo, *v2.RateLimitDescription, error) {
	return m.GetVersionFunc(ctx)
}

func (m *MockService) IsPaidPlan() bool {
	if m.IsPaidPlanFunc != nil {
		return m.IsPaidPlanFunc()
	}
	return false
}
