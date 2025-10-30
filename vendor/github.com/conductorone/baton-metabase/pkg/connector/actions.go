package connector

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	config "github.com/conductorone/baton-sdk/pb/c1/config/v1"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/actions"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	ActionEnableUser  = "enable_user"
	ActionDisableUser = "disable_user"
)

var EnableUserAction = &v2.BatonActionSchema{
	Name: ActionEnableUser,
	Arguments: []*config.Field{
		{
			Name:        "userId",
			DisplayName: "User ID",
			Field:       &config.Field_StringField{},
			IsRequired:  true,
		},
	},
	ReturnTypes: []*config.Field{
		{
			Name:        "success",
			DisplayName: "Success",
			Field:       &config.Field_BoolField{},
		},
	},
	ActionType: []v2.ActionType{
		v2.ActionType_ACTION_TYPE_ACCOUNT,
		v2.ActionType_ACTION_TYPE_ACCOUNT_ENABLE,
	},
}

var DisableUserAction = &v2.BatonActionSchema{
	Name: ActionDisableUser,
	Arguments: []*config.Field{
		{
			Name:        "userId",
			DisplayName: "User ID",
			Field:       &config.Field_StringField{},
			IsRequired:  true,
		},
	},
	ReturnTypes: []*config.Field{
		{
			Name:        "success",
			DisplayName: "Success",
			Field:       &config.Field_BoolField{},
		},
	},
	ActionType: []v2.ActionType{
		v2.ActionType_ACTION_TYPE_ACCOUNT,
		v2.ActionType_ACTION_TYPE_ACCOUNT_DISABLE,
	},
}

func (c *Connector) RegisterActionManager(ctx context.Context) (connectorbuilder.CustomActionManager, error) {
	actionManager := actions.NewActionManager(ctx)

	err := actionManager.RegisterAction(ctx, EnableUserAction.Name, EnableUserAction, c.EnableUser)
	if err != nil {
		return nil, err
	}

	err = actionManager.RegisterAction(ctx, DisableUserAction.Name, DisableUserAction, c.DisableUser)
	if err != nil {
		return nil, err
	}

	return actionManager, nil
}

func (c *Connector) EnableUser(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	ann := annotations.New()

	if args == nil {
		return nil, nil, fmt.Errorf("arguments cannot be nil")
	}

	if args.Fields == nil {
		return nil, nil, fmt.Errorf("arguments fields cannot be nil")
	}

	userId, ok := args.Fields["userId"]
	if !ok {
		return nil, nil, fmt.Errorf("missing required argument userId")
	}

	if userId == nil {
		return nil, nil, fmt.Errorf("userId value cannot be nil")
	}

	userIdStr := userId.GetStringValue()
	if userIdStr == "" {
		return nil, nil, fmt.Errorf("userId cannot be empty")
	}

	l.Info("enabling user", zap.String("userId", userIdStr))

	updatedUser, rateLimitDesc, err := c.client.UpdateUserActiveStatus(ctx, userIdStr, true)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		l.Error("failed to enable user", zap.String("userId", userIdStr), zap.Error(err))
		return nil, ann, fmt.Errorf("failed to enable user %s: %w", userIdStr, err)
	}

	success := updatedUser.IsActive
	if !success {
		l.Warn("user enable operation completed but user is still inactive", zap.String("userId", userIdStr), zap.Bool("active", updatedUser.IsActive))
	}

	response := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": structpb.NewBoolValue(success),
		},
	}
	return response, ann, nil
}

func (c *Connector) DisableUser(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	ann := annotations.New()

	if args == nil {
		return nil, nil, fmt.Errorf("arguments cannot be nil")
	}

	if args.Fields == nil {
		return nil, nil, fmt.Errorf("arguments fields cannot be nil")
	}

	userId, ok := args.Fields["userId"]
	if !ok {
		return nil, nil, fmt.Errorf("missing required argument userId")
	}

	if userId == nil {
		return nil, nil, fmt.Errorf("userId value cannot be nil")
	}

	userIdStr := userId.GetStringValue()
	if userIdStr == "" {
		return nil, nil, fmt.Errorf("userId cannot be empty")
	}

	l.Info("disabling user", zap.String("userId", userIdStr))

	updatedUser, rateLimitDesc, err := c.client.UpdateUserActiveStatus(ctx, userIdStr, false)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		l.Error("failed to disable user", zap.String("userId", userIdStr), zap.Error(err))
		return nil, ann, fmt.Errorf("failed to disable user %s: %w", userIdStr, err)
	}

	success := !updatedUser.IsActive
	if !success {
		l.Warn("user disable operation completed but user is still active", zap.String("userId", userIdStr), zap.Bool("active", updatedUser.IsActive))
	}

	response := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": structpb.NewBoolValue(success),
		},
	}
	return response, ann, nil
}
