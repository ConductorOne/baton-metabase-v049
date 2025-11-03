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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

const (
	// Headers.
	headerAPIKey = "X-API-KEY"

	// Endpoints.
	// The permissions required for these endpoints to function correctly are determined by the group (administrators) attached to the creation of the API Key.
	// For more information, please refer to docs-info.md or README.md.

	// https://www.metabase.com/docs/latest/api#tag/apipermissions/get/api/permissions/group
	getGroups = "/api/permissions/group"

	// https://www.metabase.com/docs/latest/api#tag/apiuser/get/api/user/
	getUsers = "/api/user"

	// https://www.metabase.com/docs/latest/api#tag/apiuser/get/api/user/{id}
	getUserByID = "/api/user"

	// https://www.metabase.com/docs/latest/api#tag/apiuser/post/api/user/
	createUser = "/api/user"

	// https://www.metabase.com/docs/latest/api#tag/apiuser/put/api/user/{id}/reactivate
	activateUser = "/api/user/%s/reactivate"

	// https://www.metabase.com/docs/latest/api#tag/apiuser/delete/api/user/{id}
	deactivateUser = "/api/user/%s"

	// https://www.metabase.com/docs/latest/api#tag/apipermissions/post/api/permissions/membership
	getMemberships = "/api/permissions/membership"

	// https://www.metabase.com/docs/latest/api#tag/apipermissions/post/api/permissions/membership
	addUserToGroup = "/api/permissions/membership"

	// https://www.metabase.com/docs/latest/api#tag/apipermissions/delete/api/permissions/membership/{id}
	removeUserFromGroup = "/api/permissions/membership/%s"
)

type MetabaseClient struct {
	client     *uhttp.BaseHttpClient
	baseURL    *url.URL
	apiKey     string
	isPaidPlan bool
}

func New(ctx context.Context, rawBaseURL string, apiKey string, isPaidPlan bool) (*MetabaseClient, error) {
	l := ctxzap.Extract(ctx)

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

	if baseURL.Scheme != "https" {
		l.Warn("Metabase connector is using HTTP. Make sure this instance is running in a trusted or on-premise environment.")
	}

	return &MetabaseClient{
		client:     httpClient,
		baseURL:    baseURL,
		apiKey:     apiKey,
		isPaidPlan: isPaidPlan,
	}, nil
}

func (c *MetabaseClient) doRequest(ctx context.Context, method string, url *url.URL, target interface{}, body interface{}, opts ...ReqOpt) (*http.Header, *v2.RateLimitDescription, error) {
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
	response, err := c.client.Do(request, uhttp.WithRatelimitData(&rateLimitData))
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

	if target != nil {
		if err := json.NewDecoder(response.Body).Decode(target); err != nil {
			return nil, &rateLimitData, fmt.Errorf("failed to decode JSON response: %w", err)
		}
	}

	return &response.Header, &rateLimitData, nil
}

func (c *MetabaseClient) ListUsers(ctx context.Context, options PageOptions) ([]*User, string, *v2.RateLimitDescription, error) {
	var res UsersQueryResponse

	queryUrl := c.baseURL.JoinPath(getUsers)

	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodGet, queryUrl, &res, nil,
		withLimitParam(options.Limit),
		withOffsetParam(options.Offset),
		withStatusAllParam())
	if err != nil {
		return nil, "", rateLimitDesc, fmt.Errorf("failed to fetch users: %w", err)
	}

	nextToken := getNextPageToken(res.Offset, res.Limit, res.Total)

	return res.Data, nextToken, rateLimitDesc, nil
}

func (c *MetabaseClient) CreateUser(ctx context.Context, request *CreateUserRequest) (*User, *v2.RateLimitDescription, error) {
	queryUrl := c.baseURL.JoinPath(createUser)

	var user User
	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodPost, queryUrl, &user, request)
	if err != nil {
		return nil, rateLimitDesc, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, rateLimitDesc, nil
}

func (c *MetabaseClient) GetUserByID(ctx context.Context, userID string) (*User, *v2.RateLimitDescription, error) {
	queryUrl := c.baseURL.JoinPath(getUserByID, url.PathEscape(userID))

	var user User
	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodGet, queryUrl, &user, nil)
	if err != nil {
		return nil, rateLimitDesc, fmt.Errorf("failed to fetch user by ID %s: %w", userID, err)
	}

	return &user, rateLimitDesc, nil
}

func (c *MetabaseClient) UpdateUserActiveStatus(ctx context.Context, userID string, active bool) (*User, *v2.RateLimitDescription, error) {
	var (
		queryUrl *url.URL
		method   string
	)

	if active {
		method = http.MethodPut
		queryUrl = c.baseURL.JoinPath(fmt.Sprintf(activateUser, url.PathEscape(userID)))
	} else {
		method = http.MethodDelete
		queryUrl = c.baseURL.JoinPath(fmt.Sprintf(deactivateUser, url.PathEscape(userID)))
	}

	var user User
	_, rateLimitDesc, err := c.doRequest(ctx, method, queryUrl, &user, nil)
	if err != nil {
		return nil, rateLimitDesc, fmt.Errorf("failed to update user active status in Metabase: %w", err)
	}

	return &user, rateLimitDesc, nil
}

func (c *MetabaseClient) ListGroups(ctx context.Context) ([]*Group, *v2.RateLimitDescription, error) {
	var resp []*Group

	queryUrl := c.baseURL.JoinPath(getGroups)

	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodGet, queryUrl, &resp, nil)
	if err != nil {
		return nil, rateLimitDesc, fmt.Errorf("failed to fetch groups: %w", err)
	}

	return resp, rateLimitDesc, nil
}

func (c *MetabaseClient) ListMemberships(ctx context.Context) (map[string][]*Membership, *v2.RateLimitDescription, error) {
	var membershipResponse map[string][]*Membership

	queryUrl := c.baseURL.JoinPath(getMemberships)

	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodGet, queryUrl, &membershipResponse, nil)
	if err != nil {
		return nil, rateLimitDesc, fmt.Errorf("failed to fetch memberships: %w", err)
	}

	return membershipResponse, rateLimitDesc, nil
}

func (c *MetabaseClient) AddUserToGroup(ctx context.Context, request *Membership) (*v2.RateLimitDescription, error) {
	queryUrl := c.baseURL.JoinPath(addUserToGroup)

	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodPost, queryUrl, nil, request)
	if err != nil {
		return rateLimitDesc, fmt.Errorf("failed to add user %d to group %d: %w", request.UserID, request.GroupID, err)
	}

	return rateLimitDesc, nil
}

func (c *MetabaseClient) RemoveUserFromGroup(ctx context.Context, membershipID string) (*v2.RateLimitDescription, error) {
	queryUrl := c.baseURL.JoinPath(fmt.Sprintf(removeUserFromGroup, url.PathEscape(membershipID)))

	_, rateLimitDesc, err := c.doRequest(ctx, http.MethodDelete, queryUrl, nil, nil)
	if err != nil {
		return rateLimitDesc, fmt.Errorf("failed to remove membership %s from group: %w", membershipID, err)
	}

	return rateLimitDesc, nil
}

func (c *MetabaseClient) IsPaidPlan() bool {
	return c.isPaidPlan
}
