package connector

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

func (g *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement
	opts := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(UserResourceType),
		entitlement.WithDisplayName(fmt.Sprintf("%s %s", resource.DisplayName, "Member")),
		entitlement.WithDescription(fmt.Sprintf("Is a %s of %s group in Metabase", "Member", resource.DisplayName)),
	}
	rv = append(rv, entitlement.NewAssignmentEntitlement(resource, MemberPermission, opts...))

	if g.client.IsPaidPlan() {
		opts := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(UserResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s %s", resource.DisplayName, "Manager")),
			entitlement.WithDescription(fmt.Sprintf("Is a %s of %s group in Metabase", "Manager", resource.DisplayName)),
		}
		rv = append(rv, entitlement.NewAssignmentEntitlement(resource, ManagerPermission, opts...))
	}

	return rv, "", nil, nil
}

// Grants is intentionally empty because group membership grants are computed in the userBuilder.
func (g *groupBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (g *groupBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	ann := annotations.New()

	groupID, err := strconv.Atoi(entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, fmt.Errorf("invalid group id %q: %w", entitlement.Resource.Id.Resource, err)
	}

	userIDStr := principal.Id.Resource
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id %q: %w", principal.Id.Resource, err)
	}

	memberships, rateLimitDesc, err := g.client.ListMemberships(ctx)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return ann, fmt.Errorf("failed to list memberships: %w", err)
	}

	userMemberships := memberships[userIDStr]
	for _, m := range userMemberships {
		if m.GroupID == groupID {
			return annotations.New(&v2.GrantAlreadyExists{}), nil
		}
	}

	var isManager bool
	switch {
	case strings.HasSuffix(entitlement.Id, ":"+ManagerPermission) || entitlement.Id == ManagerPermission:
		isManager = true
	case strings.HasSuffix(entitlement.Id, ":"+MemberPermission) || entitlement.Id == MemberPermission:
		isManager = false
	default:
		return nil, fmt.Errorf("unsupported entitlement id %q", entitlement.Id)
	}

	reqBody := &client.Membership{
		GroupID:        groupID,
		UserID:         userID,
		IsGroupManager: isManager,
	}

	rateLimitDesc, err = g.client.AddUserToGroup(ctx, reqBody)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return ann, fmt.Errorf("failed to grant user %d to group %d: %w", userID, groupID, err)
	}

	return ann, nil
}

func (g *groupBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	ann := annotations.New()

	groupID, err := strconv.Atoi(grant.Entitlement.Resource.Id.Resource)
	if err != nil {
		return nil, fmt.Errorf("invalid group id %q: %w", grant.Entitlement.Resource.Id.Resource, err)
	}

	userIDStr := grant.Principal.Id.Resource
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id %q: %w", grant.Principal.Id.Resource, err)
	}

	memberships, rateLimitDesc, err := g.client.ListMemberships(ctx)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return ann, fmt.Errorf("failed to list memberships: %w", err)
	}

	userMemberships := memberships[userIDStr]
	var targetMembership *client.Membership
	for _, m := range userMemberships {
		if m.GroupID == groupID {
			targetMembership = m
			break
		}
	}

	if targetMembership == nil {
		return annotations.New(&v2.GrantAlreadyRevoked{}), nil
	}

	rateLimitDesc, err = g.client.RemoveUserFromGroup(ctx, strconv.Itoa(targetMembership.MembershipID))
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return ann, fmt.Errorf("failed to revoke user %d from group %d: %w", userID, groupID, err)
	}

	return ann, nil
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
