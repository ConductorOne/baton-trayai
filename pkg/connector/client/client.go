package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

// Params is the parameters used to init a tray.io client.
type Params struct {
	HttpClient *uhttp.BaseHttpClient
}

// Client is used to interact with Tray.io.
type Client struct {
	httpClient *uhttp.BaseHttpClient
}

// NewClient initializes a new tray.ai Client.
func NewClient(p Params) *Client {
	return &Client{
		httpClient: p.HttpClient,
	}
}

// ListParams are the params passed to all `list` endpoint.
type ListParams struct {
	Cursor string
	First  int // page size.
	Last   int

	// ListUsers params.
	Email string

	// ListWorkspaces params.
	WorkspaceType string // one of [Embedded, Organization, Personal, PersonalExternal, Shared]
}

// ListResp is the response returned from all `list` endpoint.
type ListResp struct {
	Elements []Element `json:"elements"`
	Page     PageInfo  `json:"pageInfo"`
}

// ListUsers list all the users from tray.ai.
func (c *Client) ListUsers(ctx context.Context, params ListParams) (*ListResp, error) {
	urlpath, err := url.Parse(basePath + listUsersPath)
	if err != nil {
		return nil, err
	}

	urlpath.RawQuery = toQuery(urlpath, params)

	req, err := c.httpClient.NewRequest(ctx,
		http.MethodGet,
		urlpath,
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
	)

	if err != nil {
		return nil, err
	}

	var resp *ListResp
	rawResp, err := c.httpClient.Do(req, uhttp.WithJSONResponse(&resp))
	if err != nil {
		return nil, err
	}

	defer rawResp.Body.Close()
	return resp, nil
}

func toQuery(url *url.URL, p ListParams) string {
	q := url.Query()
	if p.Cursor != "" {
		q.Set("cursor", p.Cursor)
	}
	if p.First != 0 {
		q.Set("first", strconv.Itoa(p.First))
	}
	if p.Last != 0 {
		q.Set("last", strconv.Itoa(p.Last))
	}
	if p.Email != "" {
		q.Set("email", p.Email)
	}
	if p.WorkspaceType != "" {
		q.Set("type", p.WorkspaceType)
	}
	return q.Encode()
}

func (c *Client) ListWorkspaces(ctx context.Context, params ListParams) (*ListResp, error) {
	urlpath, err := url.Parse(basePath + listWorkspacesPath)
	if err != nil {
		return nil, err
	}

	urlpath.RawQuery = toQuery(urlpath, params)

	req, err := c.httpClient.NewRequest(ctx,
		http.MethodGet,
		urlpath,
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
	)

	if err != nil {
		return nil, err
	}

	var resp *ListResp
	rawResp, err := c.httpClient.Do(req, uhttp.WithJSONResponse(&resp))
	if err != nil {
		return nil, err
	}

	defer rawResp.Body.Close()
	return resp, nil
}
