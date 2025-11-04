package connector

import (
	"context"
	"fmt"

	baseConnector "github.com/conductorone/baton-metabase/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/actions"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (c *Connector) RegisterActionManager(ctx context.Context) (connectorbuilder.CustomActionManager, error) {
	actionManager := actions.NewActionManager(ctx)

	err := actionManager.RegisterAction(ctx, baseConnector.EnableUserAction.Name, baseConnector.EnableUserAction, c.EnableUserV049)
	if err != nil {
		return nil, err
	}

	err = actionManager.RegisterAction(ctx, baseConnector.DisableUserAction.Name, baseConnector.DisableUserAction, c.DisableUserV049)
	if err != nil {
		return nil, err
	}

	return actionManager, nil
}

func (c *Connector) EnableUserV049(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	ann := annotations.New()

	userIdField, ok := args.Fields["userId"]
	if !ok || userIdField == nil {
		return nil, ann, fmt.Errorf("userId field is required")
	}
	userId := userIdField.GetStringValue()
	if userId == "" {
		return nil, ann, fmt.Errorf("userId cannot be empty")
	}

	_, rateLimitDesc, err := c.vBaseClient.GetUserByID(ctx, userId)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		// GetUserByID only retrieves enabled users.
		// If the user is disabled in Metabase, the API returns a 404 (NotFound),
		// so we treat this as a signal to re-enable the user instead of a real "not found" error.
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			resp, ann2, err := c.vBaseConnector.EnableUser(ctx, args)
			if ann2 != nil {
				ann.Merge(ann2...)
			}
			if err != nil {
				return nil, ann, fmt.Errorf("failed to enable user %s: %w", userId, err)
			}
			return resp, ann, nil
		}
		return nil, ann, fmt.Errorf("failed to fetch user %s: %w", userId, err)
	}

	l.Debug("user already active, skipping enable", zap.String("userId", userId))
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": structpb.NewBoolValue(true),
		},
	}, ann, nil
}

func (c *Connector) DisableUserV049(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	ann := annotations.New()

	userIdField, ok := args.Fields["userId"]
	if !ok || userIdField == nil {
		return nil, ann, fmt.Errorf("userId field is required")
	}
	userId := userIdField.GetStringValue()
	if userId == "" {
		return nil, ann, fmt.Errorf("userId cannot be empty")
	}

	user, rateLimitDesc, err := c.vBaseClient.GetUserByID(ctx, userId)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		// GetUserByID only retrieves enabled users.
		// If the user is disabled in Metabase, the API returns a 404 (NotFound),
		// In that case, the user is already disabled, so we skip the operation and return success.
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			l.Debug("user not found (already disabled), skipping disable", zap.String("userId", userId))
			return &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"success": structpb.NewBoolValue(true),
				},
			}, ann, nil
		}
		return nil, ann, fmt.Errorf("failed to fetch user %s: %w", userId, err)
	}

	if user.IsActive {
		resp, ann2, err := c.vBaseConnector.DisableUser(ctx, args)
		if ann2 != nil {
			ann.Merge(ann2...)
		}
		if err != nil {
			return nil, ann, fmt.Errorf("failed to disable user %s: %w", userId, err)
		}
		return resp, ann, nil
	}

	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": structpb.NewBoolValue(true),
		},
	}, ann, nil
}
