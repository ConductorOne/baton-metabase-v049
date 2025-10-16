package connector

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/conductorone/baton-metabase-v049/pkg/client"
	cfg "github.com/conductorone/baton-metabase-v049/pkg/config"
	baseClient "github.com/conductorone/baton-metabase/pkg/client"
	baseConfig "github.com/conductorone/baton-metabase/pkg/config"
	baseConnector "github.com/conductorone/baton-metabase/pkg/connector"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type Connector struct {
	vBaseConnector *baseConnector.Connector
	vBaseClient    *baseClient.MetabaseClient
	v049Client     *client.MetabaseV049Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	syncers := d.vBaseConnector.ResourceSyncers(ctx)
	syncers = append(syncers,
		newDatabaseBuilder(d.v049Client),
	)

	return syncers
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(_ context.Context, _ *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(_ context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Metabase v0-49",
		Description: "Metabase connector v0-49 to sync users, groups and databases",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	ann := annotations.New()

	versionResp, rateLimitDesc, err := d.vBaseClient.GetVersion(ctx)
	if rateLimitDesc != nil {
		ann.WithRateLimiting(rateLimitDesc)
	}
	if err != nil {
		l.Error("failed to fetch Metabase version", zap.Error(err))
		return ann, fmt.Errorf("failed to fetch Metabase version: %w", err)
	}

	tag := strings.TrimPrefix(versionResp.Tag, "v")
	parts := strings.Split(tag, ".")
	if len(parts) < 2 {
		return ann, fmt.Errorf("unexpected version format: %s", versionResp.Tag)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return ann, fmt.Errorf("invalid major version in tag %s: %w", versionResp.Tag, err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return ann, fmt.Errorf("invalid minor version in tag %s: %w", versionResp.Tag, err)
	}

	// Metabase v0.49.x only
	if major != 0 || minor < 49 || minor >= 50 {
		return ann, fmt.Errorf("unsupported Metabase version: %s (this connector supports only Metabase v0.49.x)", versionResp.Tag)
	}

	return ann, nil
}

func New(ctx context.Context, config *cfg.MetabaseV049) (*Connector, error) {
	l := ctxzap.Extract(ctx)

	baseCfg := &baseConfig.Metabase{
		MetabaseBaseUrl: config.MetabaseBaseUrl,
		MetabaseApiKey:  config.MetabaseApiKey,
	}

	vBaseConnector, err := baseConnector.New(ctx, baseCfg)
	if err != nil {
		l.Error("failed to create base Metabase connector", zap.Error(err))
		return nil, err
	}

	v049Client, err := client.NewV049Client(ctx, config.MetabaseBaseUrl, config.MetabaseApiKey)
	if err != nil {
		l.Error("failed to create extended Metabase v0.49 client", zap.Error(err))
		return nil, err
	}

	vBaseClient, err := baseClient.New(ctx, config.MetabaseBaseUrl, config.MetabaseApiKey)
	if err != nil {
		l.Error("failed to create base Metabase client", zap.Error(err))
		return nil, err
	}

	return &Connector{
		vBaseConnector: vBaseConnector,
		vBaseClient:    vBaseClient,
		v049Client:     v049Client,
	}, nil
}
