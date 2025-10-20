package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type MockService struct {
	ListUsersFunc       func(ctx context.Context, options PageOptions) ([]*User, string, *v2.RateLimitDescription, error)
	ListGroupsFunc      func(ctx context.Context) ([]*Group, *v2.RateLimitDescription, error)
	ListMembershipsFunc func(ctx context.Context) (map[string][]*Membership, *v2.RateLimitDescription, error)
	GetVersionFunc      func(ctx context.Context) (*VersionInfo, *v2.RateLimitDescription, error)
	IsPaidPlanFunc      func() bool
}

func (m *MockService) ListUsers(ctx context.Context, options PageOptions) ([]*User, string, *v2.RateLimitDescription, error) {
	return m.ListUsersFunc(ctx, options)
}

func (m *MockService) ListGroups(ctx context.Context) ([]*Group, *v2.RateLimitDescription, error) {
	return m.ListGroupsFunc(ctx)
}

func (m *MockService) ListMemberships(ctx context.Context) (map[string][]*Membership, *v2.RateLimitDescription, error) {
	return m.ListMembershipsFunc(ctx)
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
