package connector

import (
	"context"

	baseConnector "github.com/conductorone/baton-metabase/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/actions"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

func (c *Connector) RegisterActionManager(ctx context.Context) (connectorbuilder.CustomActionManager, error) {
	actionManager := actions.NewActionManager(ctx)

	err := actionManager.RegisterAction(ctx, baseConnector.EnableUserAction.Name, baseConnector.EnableUserAction, c.vBaseConnector.EnableUser)
	if err != nil {
		return nil, err
	}

	err = actionManager.RegisterAction(ctx, baseConnector.DisableUserAction.Name, baseConnector.DisableUserAction, c.vBaseConnector.DisableUser)
	if err != nil {
		return nil, err
	}

	return actionManager, nil
}
