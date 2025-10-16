package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type MockService struct {
	ListDatabasesFunc    func(ctx context.Context) ([]*Database, *v2.RateLimitDescription, error)
	GetDBPermissionsFunc func(ctx context.Context, dbID string) (map[string]map[string]*GroupPermission, *v2.RateLimitDescription, error)
}

func (m *MockService) ListDatabases(ctx context.Context) ([]*Database, *v2.RateLimitDescription, error) {
	return m.ListDatabasesFunc(ctx)
}

func (m *MockService) GetDBPermissions(ctx context.Context, dbID string) (map[string]map[string]*GroupPermission, *v2.RateLimitDescription, error) {
	return m.GetDBPermissionsFunc(ctx, dbID)
}
