package connector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/conductorone/baton-metabase/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/crypto"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	resourceSdk "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	client client.ClientService
}

func (u *userBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return UserResourceType
}

func (u *userBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	opts, err := getPageOptions(pToken, client.ItemsPerPage)
	if err != nil {
		return nil, "", nil, err
	}

	ann := annotations.New()

	users, nextPageToken, rateLimitDesc, err := u.client.ListUsers(ctx, opts)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return nil, "", ann, fmt.Errorf("failed to list users: %w", err)
	}

	outResources := make([]*v2.Resource, 0, len(users))
	for _, user := range users {
		res, err := u.parseIntoUserResource(user)
		if err != nil {
			return nil, "", ann, err
		}
		outResources = append(outResources, res)
	}

	return outResources, nextPageToken, ann, nil
}

// Entitlements always returns an empty slice for users.
func (u *userBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants returns the user's memberships as grants to groups.
// We implement it here in users (instead of groups) to avoid inefficient lookups:
// for each user we already have their memberships, so we can generate grants directly.
// Placing this in groups would require iterating over all users for each group,
// which is costly and unnecessary.
func (u *userBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	ann := annotations.New()
	allMemberships, rateLimitDesc, err := u.client.ListMemberships(ctx)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return nil, "", ann, fmt.Errorf("failed to list memberships: %w", err)
	}

	userMemberships, ok := allMemberships[resource.Id.Resource]
	if !ok {
		return nil, "", ann, nil
	}

	grants := make([]*v2.Grant, 0, len(userMemberships))
	for _, membership := range userMemberships {
		groupResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: GroupResourceType.Id,
				Resource:     strconv.Itoa(membership.GroupID),
			},
		}

		role := MemberPermission
		if membership.IsGroupManager {
			role = ManagerPermission
		}

		grants = append(grants, grant.NewGrant(
			groupResource,
			role,
			resource.Id,
		))
	}

	return grants, "", ann, nil
}

func (u *userBuilder) CreateAccountCapabilityDetails(
	_ context.Context,
) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
	}, nil, nil
}

func (u *userBuilder) CreateAccount(
	ctx context.Context,
	accountInfo *v2.AccountInfo,
	_ *v2.LocalCredentialOptions,
) (
	connectorbuilder.CreateAccountResponse,
	[]*v2.PlaintextData,
	annotations.Annotations,
	error,
) {
	ann := annotations.New()
	profile := accountInfo.GetProfile().AsMap()

	email, ok := profile["email"].(string)
	if !ok || email == "" {
		return nil, nil, nil, fmt.Errorf("missing required field: email")
	}

	firstName, ok := profile["first_name"].(string)
	if !ok || firstName == "" {
		return nil, nil, nil, fmt.Errorf("missing required field: first_name")
	}

	lastName, ok := profile["last_name"].(string)
	if !ok || lastName == "" {
		return nil, nil, nil, fmt.Errorf("missing required field: last_name")
	}

	password, err := crypto.GenerateRandomPassword(&v2.LocalCredentialOptions_RandomPassword{
		Length: 12,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate password: %w", err)
	}

	createReq := &client.CreateUserRequest{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	}

	user, rateLimitDesc, err := u.client.CreateUser(ctx, createReq)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		return nil, nil, ann, err
	}

	userResource, err := u.parseIntoUserResource(user)
	if err != nil {
		return nil, nil, ann, err
	}

	resp := &v2.CreateAccountResponse_SuccessResult{
		Resource:              userResource,
		IsCreateAccountResult: true,
	}

	plaintexts := []*v2.PlaintextData{
		{
			Name:  "password",
			Bytes: []byte(password),
		},
	}

	return resp, plaintexts, ann, nil
}

func (u *userBuilder) parseIntoUserResource(user *client.User) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}

	traitOptions := []resourceSdk.UserTraitOption{
		resourceSdk.WithEmail(user.Email, true),
		resourceSdk.WithUserLogin(user.Email),
		resourceSdk.WithUserProfile(profile),
	}

	if user.LastLogin != nil {
		traitOptions = append(traitOptions, resourceSdk.WithLastLogin(*user.LastLogin))
	}

	if user.IsActive {
		traitOptions = append(traitOptions, resourceSdk.WithStatus(v2.UserTrait_Status_STATUS_ENABLED))
	} else {
		traitOptions = append(traitOptions, resourceSdk.WithStatus(v2.UserTrait_Status_STATUS_DISABLED))
	}

	return resourceSdk.NewUserResource(
		fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		UserResourceType,
		user.ID,
		traitOptions,
	)
}

func newUserBuilder(client client.ClientService) *userBuilder {
	return &userBuilder{
		client: client,
	}
}
