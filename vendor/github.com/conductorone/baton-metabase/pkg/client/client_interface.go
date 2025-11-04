package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type ClientService interface {
	ListUsers(ctx context.Context, options PageOptions) ([]*User, string, *v2.RateLimitDescription, error)
	ListGroups(ctx context.Context) ([]*Group, *v2.RateLimitDescription, error)
	ListMemberships(ctx context.Context) (map[string][]*Membership, *v2.RateLimitDescription, error)
	IsPaidPlan() bool
	CreateUser(ctx context.Context, payload *CreateUserRequest) (*User, *v2.RateLimitDescription, error)
	UpdateUserActiveStatus(ctx context.Context, userId string, active bool) (*User, *v2.RateLimitDescription, error)
	AddUserToGroup(ctx context.Context, request *Membership) (*v2.RateLimitDescription, error)
	RemoveUserFromGroup(ctx context.Context, membershipID string) (*v2.RateLimitDescription, error)
	GetUserByID(ctx context.Context, userID string) (*User, *v2.RateLimitDescription, error)
}
