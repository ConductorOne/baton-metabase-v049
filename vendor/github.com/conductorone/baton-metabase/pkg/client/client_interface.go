package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type ClientService interface {
	ListUsers(ctx context.Context, options PageOptions) ([]*User, string, *v2.RateLimitDescription, error)
	ListGroups(ctx context.Context) ([]*Group, *v2.RateLimitDescription, error)
	ListMemberships(ctx context.Context) (map[string][]*Membership, *v2.RateLimitDescription, error)
	GetVersion(ctx context.Context) (*VersionInfo, *v2.RateLimitDescription, error)
}
