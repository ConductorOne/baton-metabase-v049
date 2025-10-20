package connector

import (
	"context"
	"fmt"
	"testing"

	"github.com/conductorone/baton-metabase/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/stretchr/testify/require"
)

func newTestClient() *client.MockService {
	return &client.MockService{}
}

func TestGetVersion(t *testing.T) {
	ctx := context.Background()
	mockClient := newTestClient()

	t.Run("should return version successfully", func(t *testing.T) {
		mockClient.GetVersionFunc = func(ctx context.Context) (*client.VersionInfo, *v2.RateLimitDescription, error) {
			return &client.VersionInfo{Tag: "v0.49.2"}, nil, nil
		}

		versionInfo, rateLimit, err := mockClient.GetVersion(ctx)
		require.NoError(t, err)
		require.NotNil(t, versionInfo)
		require.Equal(t, "v0.49.2", versionInfo.Tag)
		require.Nil(t, rateLimit)
	})

	t.Run("should return error if API fails", func(t *testing.T) {
		mockClient.GetVersionFunc = func(ctx context.Context) (*client.VersionInfo, *v2.RateLimitDescription, error) {
			return nil, nil, fmt.Errorf("API error")
		}

		versionInfo, rateLimit, err := mockClient.GetVersion(ctx)
		require.Error(t, err)
		require.Nil(t, versionInfo)
		require.Nil(t, rateLimit)
		require.Contains(t, err.Error(), "API error")
	})
}
