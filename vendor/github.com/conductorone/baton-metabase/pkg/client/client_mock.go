package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type MockService struct {
	ListUsersFunc              func(ctx context.Context, options PageOptions) ([]*User, string, *v2.RateLimitDescription, error)
	ListGroupsFunc             func(ctx context.Context) ([]*Group, *v2.RateLimitDescription, error)
	ListMembershipsFunc        func(ctx context.Context) (map[string][]*Membership, *v2.RateLimitDescription, error)
	IsPaidPlanFunc             func() bool
	CreateUserFunc             func(ctx context.Context, request *CreateUserRequest) (*User, *v2.RateLimitDescription, error)
	UpdateUserActiveStatusFunc func(ctx context.Context, userId string, active bool) (*User, *v2.RateLimitDescription, error)
	AddUserToGroupFunc         func(ctx context.Context, request *Membership) (*v2.RateLimitDescription, error)
	RemoveUserFromGroupFunc    func(ctx context.Context, membershipID string) (*v2.RateLimitDescription, error)
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

func (m *MockService) IsPaidPlan() bool {
	if m.IsPaidPlanFunc != nil {
		return m.IsPaidPlanFunc()
	}
	return false
}

func (m *MockService) CreateUser(ctx context.Context, request *CreateUserRequest) (*User, *v2.RateLimitDescription, error) {
	return m.CreateUserFunc(ctx, request)
}

func (m *MockService) UpdateUserActiveStatus(ctx context.Context, userId string, active bool) (*User, *v2.RateLimitDescription, error) {
	return m.UpdateUserActiveStatusFunc(ctx, userId, active)
}

func (m *MockService) AddUserToGroup(ctx context.Context, request *Membership) (*v2.RateLimitDescription, error) {
	return m.AddUserToGroupFunc(ctx, request)
}

func (m *MockService) RemoveUserFromGroup(ctx context.Context, membershipID string) (*v2.RateLimitDescription, error) {
	return m.RemoveUserFromGroupFunc(ctx, membershipID)
}
