package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-metabase-v049/pkg/client"
	baseConnector "github.com/conductorone/baton-metabase/pkg/connector"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

const (
	accessPermission = "access"
	writePermission  = "write"
)

type databaseBuilder struct {
	client client.ClientService
}

func (d *databaseBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return databaseResourceType
}

func (d *databaseBuilder) List(ctx context.Context, _ *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	ann := annotations.New()

	databases, rateLimitDesc, err := d.client.ListDatabases(ctx)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return nil, "", ann, err
	}

	outResources := make([]*v2.Resource, 0, len(databases))
	for _, database := range databases {
		res, err := d.parseIntoDatabaseResource(database)
		if err != nil {
			return nil, "", ann, err
		}
		outResources = append(outResources, res)
	}

	return outResources, "", ann, nil
}

var databasePermissions = []struct {
	ID          string
	DisplayName string
}{
	{ID: accessPermission, DisplayName: "Access"},
	{ID: writePermission, DisplayName: "Write"},
}

func (d *databaseBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	for _, permission := range databasePermissions {
		opts := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(baseConnector.GroupResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s %s", resource.DisplayName, permission.DisplayName)),
			entitlement.WithDescription(fmt.Sprintf("Grants %s permission on the %s database", permission.DisplayName, resource.DisplayName)),
		}
		ent := entitlement.NewPermissionEntitlement(resource, permission.ID, opts...)
		rv = append(rv, ent)
	}

	return rv, "", nil, nil
}

func (d *databaseBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	dbID := resource.Id.Resource
	ann := annotations.New()

	groups, rateLimitDesc, err := d.client.GetDBPermissions(ctx, dbID)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return nil, "", ann, err
	}

	var grants []*v2.Grant
	for groupIDStr, dbPermissions := range groups {
		permissions, ok := dbPermissions[dbID]
		if !ok {
			continue
		}

		groupResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: baseConnector.GroupResourceType.Id,
				Resource:     groupIDStr,
			},
		}

		entitlementIDs := []string{
			fmt.Sprintf("%s:%s:%s", baseConnector.GroupResourceType.Id, groupIDStr, baseConnector.MemberPermission),
		}

		if d.client.IsPaidPlan() {
			entitlementIDs = append(entitlementIDs,
				fmt.Sprintf("%s:%s:%s", baseConnector.GroupResourceType.Id, groupIDStr, baseConnector.ManagerPermission),
			)
		}

		if permissions.Data != nil {
			grants = append(grants, grant.NewGrant(resource,
				accessPermission,
				groupResource,
				grant.WithAnnotation(&v2.GrantExpandable{
					EntitlementIds: entitlementIDs,
				}),
			))

			if permissions.Data.NativePermission == writePermission {
				grants = append(grants, grant.NewGrant(
					resource,
					writePermission,
					groupResource,
					grant.WithAnnotation(&v2.GrantExpandable{
						EntitlementIds: entitlementIDs,
					}),
				))
			}
		}
	}

	return grants, "", ann, nil
}

func (d *databaseBuilder) parseIntoDatabaseResource(database *client.Database) (*v2.Resource, error) {
	return resourceSdk.NewResource(
		database.Name,
		databaseResourceType,
		database.ID,
	)
}

func newDatabaseBuilder(client client.ClientService) *databaseBuilder {
	return &databaseBuilder{
		client: client,
	}
}
