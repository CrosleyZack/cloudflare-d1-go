package client

import (
	"context"
	"errors"
	"fmt"

	cloudflared1 "github.com/crosleyzack/cloudflare-d1-go"
	"github.com/crosleyzack/cloudflare-d1-go/utils"
)

type Client struct {
	AccountID string
	APIToken  string
	// track map of dbName->dbID to facilitate lookups by name
	NameIDMap map[string]string
}

var _ cloudflared1.CloudflareD1 = (*Client)(nil)

// NewClient creates a client for communicating with Cloudflare D1
func NewClient(accountID, apiToken string) (*Client, error) {
	if accountID == "" || apiToken == "" {
		return nil, errors.New("Invalid account ID and/or API Token")
	}
	return &Client{
		AccountID: accountID,
		APIToken:  apiToken,
		NameIDMap: map[string]string{},
	}, nil
}

// CreateDB create a new database with the given name in the cloudflare account.
func (c *Client) CreateDB(_ context.Context, dbName string) (*utils.APIResponse[cloudflared1.D1Database], error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database", c.AccountID)
	body := map[string]any{
		"name": dbName,
	}
	res, err := utils.DoRequest[cloudflared1.D1Database]("POST", url, body, c.APIToken)
	if err != nil {
		return nil, err
	}
	c.NameIDMap[dbName] = res.Result.UUID.String()
	return res, err
}

// DeleteDB delete a database by ID in the cloudflare account.
func (c *Client) DeleteDB(_ context.Context, dbID string) (*utils.APIResponse[cloudflared1.DeleteResult], error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s", c.AccountID, dbID)
	return utils.DoRequest[cloudflared1.DeleteResult]("DELETE", url, nil, c.APIToken)
}

// UpdateDB update the database settings by ID in the cloudflare account.
func (c *Client) UpdateDB(_ context.Context, dbID string, settings cloudflared1.DBSettings) (*utils.APIResponse[cloudflared1.D1Database], error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s", c.AccountID, dbID)
	body := map[string]any{
		"read_replication": map[string]any{
			"mode": settings.Replication.String(),
		},
	}
	return utils.DoRequest[cloudflared1.D1Database]("PATCH", url, body, c.APIToken)
}

// GetDB retrieve information on a database by id in the cloudflare account.
func (c *Client) GetDB(_ context.Context, dbID string) (*utils.APIResponse[cloudflared1.D1Database], error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s", c.AccountID, dbID)
	return utils.DoRequest[cloudflared1.D1Database]("GET", url, nil, c.APIToken)
}

// ListDB list all databases in the cloudflare account.
func (c *Client) ListDB(_ context.Context) (*utils.APIResponse[cloudflared1.D1DatabaseList], error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database", c.AccountID)
	return utils.DoRequest[cloudflared1.D1DatabaseList]("GET", url, nil, c.APIToken)
}

// QueryDB execute a SQL query on the D1 database with parameters
func (c *Client) QueryDB(_ context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]cloudflared1.QueryResult[any]], error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/query", c.AccountID, dbID)
	body := map[string]any{
		"sql":    query,
		"params": params,
	}
	return utils.DoRequest[[]cloudflared1.QueryResult[any]]("POST", url, body, c.APIToken)
}

// QueryDBRaw execute a SQL query on the D1 database with parameters
func (c *Client) QueryDBRaw(_ context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]cloudflared1.QueryResult[any]], error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/raw", c.AccountID, dbID)
	body := map[string]any{
		"sql":    query,
		"params": params,
	}
	return utils.DoRequest[[]cloudflared1.QueryResult[any]]("POST", url, body, c.APIToken)
}
