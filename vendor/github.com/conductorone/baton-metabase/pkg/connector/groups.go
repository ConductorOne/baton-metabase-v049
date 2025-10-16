package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-metabase/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const (
	MemberPermission  = "member"
	ManagerPermission = "manager"
)

type groupBuilder struct {
	client client.ClientService
}

func (g *groupBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return GroupResourceType
}

func (g *groupBuilder) List(ctx context.Context, _ *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	ann := annotations.New()

	groups, rateLimitDesc, err := g.client.ListGroups(ctx)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return nil, "", ann, fmt.Errorf("failed to list groups: %w", err)
	}

	outResources := make([]*v2.Resource, 0, len(groups))
	for _, group := range groups {
		res, err := g.parseIntoGroupResource(group)
		if err != nil {
			return nil, "", ann, err
		}
		outResources = append(outResources, res)
	}

	return outResources, "", ann, nil
}

var roles = []struct {
	ID          string
	DisplayName string
}{
	{ID: MemberPermission, DisplayName: "Member"},
	{ID: ManagerPermission, DisplayName: "Manager"},
}

func (g *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement
	for _, role := range roles {
		opts := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(UserResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s %s", resource.DisplayName, role.DisplayName)),
			entitlement.WithDescription(fmt.Sprintf("Is a %s of %s group in Metabase", role.DisplayName, resource.DisplayName)),
		}
		ent := entitlement.NewAssignmentEntitlement(resource, role.ID, opts...)
		rv = append(rv, ent)
	}

	return rv, "", nil, nil
}

// Grants is intentionally empty because group membership grants are computed in the userBuilder.
func (g *groupBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (g *groupBuilder) parseIntoGroupResource(group *client.Group) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"name":         group.Name,
		"member_count": group.MemberCount,
	}

	return resourceSdk.NewGroupResource(
		group.Name,
		GroupResourceType,
		group.ID,
		[]resourceSdk.GroupTraitOption{resourceSdk.WithGroupProfile(profile)},
	)
}

func newGroupBuilder(client client.ClientService) *groupBuilder {
	return &groupBuilder{
		client: client,
	}
}
