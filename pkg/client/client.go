package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const (
	// Headers.
	headerAPIKey = "X-API-KEY"

	// Endpoints.
	// The permissions required for these endpoints to function correctly are determined by the group (administrators) attached to the creation of the API Key.
	// For more information, please refer to docs-info.md or README.md.

	// https://www.metabase.com/docs/latest/api#tag/apisetting/get/api/setting/{key}
	getVersion = "/api/setting/version"

	// https://www.metabase.com/docs/latest/api#tag/apidatabase/post/api/database/
	getDatabases = "/api/database"

	// https://www.metabase.com/docs/latest/api#tag/apipermissions/get/api/permissions/graph/db/{db-id}
	getDBPermissions = "/api/permissions/graph/db/%s"
	// Example JSON response version 0.49:
	/*
	   {
	       "revision": 13,
	       "groups": {
	           "1": {
	               "3": {
	                   "download": {
	                       "native": "full",
	                       "schemas": "full"
	                   }
	               }
	           },
	           "3": {
	               "3": {
	                   "data": {
	                       "schemas": "all",
	                       "native": "write"
	                   }
	               }
	           },
	           "4": {
	               "3": {
	                   "data": {
	                       "schemas": "all"
	                   }
	               }
	           }
	       }
	   }
	*/
)

type MetabaseV049Client struct {
	client     *uhttp.BaseHttpClient
	baseURL    *url.URL
	apiKey     string
	isPaidPlan bool
}

func NewV049Client(ctx context.Context, rawBaseURL string, apiKey string, isPaidPlan bool) (*MetabaseV049Client, error) {
	client, err := uhttp.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, client)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	return &MetabaseV049Client{
		client:     httpClient,
		baseURL:    baseURL,
		apiKey:     apiKey,
		isPaidPlan: isPaidPlan,
	}, nil
}

type ReqOpt func(reqURL *url.URL)

func (c *MetabaseV049Client) doRequest(ctx context.Context, method string, url *url.URL, target interface{}, body interface{}, opts ...ReqOpt) (*http.Header, *v2.RateLimitDescription, error) {
	for _, opt := range opts {
		opt(url)
	}

	var requestOptions []uhttp.RequestOption
	requestOptions = append(requestOptions,
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithHeader(headerAPIKey, c.apiKey))
	if body != nil {
		requestOptions = append(requestOptions, uhttp.WithContentTypeJSONHeader(), uhttp.WithJSONBody(body))
	}

	request, err := c.client.NewRequest(ctx, method, url, requestOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	var rateLimitData v2.RateLimitDescription
	doOptions := []uhttp.DoOption{
		uhttp.WithRatelimitData(&rateLimitData),
	}
	if target != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(target))
	}

	response, err := c.client.Do(request, doOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, response.Body)
		closeErr := response.Body.Close()
		if closeErr != nil {
			log.Printf("warning: failed to close response body: %v", closeErr)
		}
	}()

	// The Metabase API does not always return a JSON-formatted error body,
	// so we first read the raw response and try to parse it as JSON.
	// If parsing fails or the response is empty, we fall back to using the HTTP status text.
	if response.StatusCode >= 300 {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, &rateLimitData, fmt.Errorf("failed to read response body: %w", err)
		}
		bodyStr := strings.TrimSpace(string(bodyBytes))

		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.MessageText != "" {
			bodyStr = errResp.Message()
		}

		if bodyStr == "" {
			bodyStr = http.StatusText(response.StatusCode)
		}

		return nil, &rateLimitData, fmt.Errorf("metabase API error: status %d %s: %s",
			response.StatusCode, response.Status, bodyStr)
	}

	return &response.Header, &rateLimitData, nil
}

func (c *MetabaseV049Client) ListDatabases(ctx context.Context) ([]*Database, *v2.RateLimitDescription, error) {
	var dbResponse DatabaseAPIResponse

	queryUrl := c.baseURL.JoinPath(getDatabases)

	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodGet, queryUrl, &dbResponse, nil)
	if err != nil {
		return nil, rateLimitDesc, fmt.Errorf("failed to fetch databases: %w", err)
	}

	return dbResponse.Data, rateLimitDesc, nil
}

func (c *MetabaseV049Client) GetDBPermissions(ctx context.Context, dbID string) (map[string]map[string]*GroupPermission, *v2.RateLimitDescription, error) {
	var dbPermissions DBPermissionGraph

	queryUrl := c.baseURL.JoinPath(fmt.Sprintf(getDBPermissions, dbID))

	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodGet, queryUrl, &dbPermissions, nil)
	if err != nil {
		return nil, rateLimitDesc, fmt.Errorf("failed to fetch database permissions %s: %w", dbID, err)
	}

	return dbPermissions.Groups, rateLimitDesc, nil
}

func (c *MetabaseV049Client) GetVersion(ctx context.Context) (*VersionInfo, *v2.RateLimitDescription, error) {
	var utilInfo VersionInfo

	queryUrl := c.baseURL.JoinPath(getVersion)

	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodGet, queryUrl, &utilInfo, nil)
	if err != nil {
		return nil, rateLimitDesc, fmt.Errorf("failed to fetch Metabase version: %w", err)
	}

	return &utilInfo, rateLimitDesc, nil
}

func (c *MetabaseV049Client) IsPaidPlan() bool {
	return c.isPaidPlan
}
