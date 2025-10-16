package connector

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/conductorone/baton-metabase-v049/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func newTestDatabaseBuilder() (*databaseBuilder, *client.MockService) {
	mockClient := &client.MockService{}
	builder := newDatabaseBuilder(mockClient)
	return builder, mockClient
}

func TestDatabasesList(t *testing.T) {
	ctx := context.Background()

	t.Run("should get databases with rate limit", func(t *testing.T) {
		dbBuilder, mockClient := newTestDatabaseBuilder()
		rl := &v2.RateLimitDescription{Limit: 100, Remaining: 10}

		mockClient.ListDatabasesFunc = func(ctx context.Context) ([]*client.Database, *v2.RateLimitDescription, error) {
			return []*client.Database{{ID: 1, Name: "SalesDB"}}, rl, nil
		}

		resources, nextPageToken, ann, err := dbBuilder.List(ctx, nil, &pagination.Token{})
		require.NoError(t, err)
		require.Len(t, resources, 1)
		require.Equal(t, "SalesDB", resources[0].DisplayName)
		require.Empty(t, nextPageToken)
		require.NotEmpty(t, ann)
	})

	t.Run("should return empty list if no databases", func(t *testing.T) {
		dbBuilder, mockClient := newTestDatabaseBuilder()
		mockClient.ListDatabasesFunc = func(ctx context.Context) ([]*client.Database, *v2.RateLimitDescription, error) {
			return []*client.Database{}, nil, nil
		}

		resources, nextPageToken, ann, err := dbBuilder.List(ctx, nil, &pagination.Token{})
		require.NoError(t, err)
		require.Empty(t, resources)
		require.Empty(t, nextPageToken)
		require.Empty(t, ann)
	})

	t.Run("should return error if ListDatabases fails", func(t *testing.T) {
		dbBuilder, mockClient := newTestDatabaseBuilder()
		mockClient.ListDatabasesFunc = func(ctx context.Context) ([]*client.Database, *v2.RateLimitDescription, error) {
			return nil, nil, fmt.Errorf("API error")
		}

		_, _, _, err := dbBuilder.List(ctx, nil, &pagination.Token{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list databases: API error")
	})
}

func TestDatabasesGrants(t *testing.T) {
	ctx := context.Background()
	dbResource := &v2.Resource{
		Id:          &v2.ResourceId{ResourceType: databaseResourceType.Id, Resource: "1"},
		DisplayName: "SalesDB",
	}

	t.Run("should handle rate limit in GetDBPermissions", func(t *testing.T) {
		dbBuilder, mockClient := newTestDatabaseBuilder()
		rl := &v2.RateLimitDescription{Limit: 50, Remaining: 0}

		mockClient.GetDBPermissionsFunc = func(ctx context.Context, dbID string) (map[string]map[string]*client.GroupPermission, *v2.RateLimitDescription, error) {
			return nil, rl, fmt.Errorf("rate limit error")
		}

		grants, _, ann, err := dbBuilder.Grants(ctx, dbResource, &pagination.Token{})
		require.Nil(t, grants)
		require.Error(t, err)
		require.NotEmpty(t, ann)
	})

	t.Run("should return access and write grants correctly", func(t *testing.T) {
		dbBuilder, mockClient := newTestDatabaseBuilder()
		mockClient.GetDBPermissionsFunc = func(ctx context.Context, dbID string) (map[string]map[string]*client.GroupPermission, *v2.RateLimitDescription, error) {
			return map[string]map[string]*client.GroupPermission{
				"group3": {"1": {Data: &client.DataAccessDetails{NativePermission: "write"}}},
				"group4": {"1": {Data: &client.DataAccessDetails{NativePermission: ""}}},
			}, nil, nil
		}

		grants, _, ann, err := dbBuilder.Grants(ctx, dbResource, &pagination.Token{})
		require.NoError(t, err)
		require.Empty(t, ann)

		var g3Access, g3Write, g4Access, g4Write bool
		for _, g := range grants {
			switch g.Principal.Id.Resource {
			case "group3":
				if strings.Contains(g.Entitlement.Id, "access") {
					g3Access = true
				}
				if strings.Contains(g.Entitlement.Id, "write") {
					g3Write = true
				}
			case "group4":
				if strings.Contains(g.Entitlement.Id, "access") {
					g4Access = true
				}
				if strings.Contains(g.Entitlement.Id, "write") {
					g4Write = true
				}
			}
		}

		require.True(t, g3Access)
		require.True(t, g3Write)
		require.True(t, g4Access)
		require.False(t, g4Write)
	})

	t.Run("should return error if GetDBPermissions fails", func(t *testing.T) {
		dbBuilder, mockClient := newTestDatabaseBuilder()
		mockClient.GetDBPermissionsFunc = func(ctx context.Context, dbID string) (map[string]map[string]*client.GroupPermission, *v2.RateLimitDescription, error) {
			return nil, nil, fmt.Errorf("API error")
		}

		_, _, _, err := dbBuilder.Grants(ctx, dbResource, &pagination.Token{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list database permissions: API error")
	})
}
