package cloudflared1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	cloudflared1 "github.com/crosleyzack/cloudflare-d1-go"
	cloudflare_d1_go "github.com/crosleyzack/cloudflare-d1-go/client"
	"github.com/stretchr/testify/assert"
)

var (
	// get required values from environment
	accountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken  = os.Getenv("CLOUDFLARE_API_TOKEN")
)

func TestNewClient(t *testing.T) {
	// missing account id
	_, err := cloudflare_d1_go.NewClient("", apiToken)
	assert.Error(t, err)
	// missing token
	_, err = cloudflare_d1_go.NewClient(accountID, "")
	assert.Error(t, err)
	// valid values
	client, err := cloudflare_d1_go.NewClient(accountID, apiToken)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, client.APIToken, apiToken)
	assert.Equal(t, client.AccountID, accountID)
}

func TestCreateDeleteUpdateDB(t *testing.T) {
	newDBName := fmt.Sprintf("test-%s", randomString(4))
	client, err := cloudflare_d1_go.NewClient(accountID, apiToken)
	ctx := context.Background()
	createResult, err := client.CreateDB(ctx, newDBName)
	assert.NoError(t, err)
	assert.NotNil(t, createResult)
	assert.Len(t, createResult.Errors, 0)
	assert.True(t, createResult.Success)
	assert.Equal(t, newDBName, createResult.Result.Name)
	dbID := createResult.Result.UUID
	assert.Equal(t, dbID.String(), client.NameIDMap[newDBName])

	if !createResult.Success {
		panic("Test cannot continue, db creation failed")
	}

	// get db
	getResult, err := client.GetDB(ctx, newDBName)
	assert.NoError(t, err)
	assert.NotNil(t, getResult)
	assert.Len(t, getResult.Errors, 0)
	assert.True(t, getResult.Success)
	assert.Equal(t, newDBName, createResult.Result.Name)

	// list db
	listResult, err := client.ListDB(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, listResult)
	assert.Len(t, listResult.Errors, 0)
	assert.Greater(t, len(listResult.Result), 0)
	assert.True(t, listResult.Success)
	assert.Equal(t, newDBName, listResult.Result[0].Name)

	// update db
	updateRes, err := client.UpdateDB(ctx, dbID.String(), cloudflared1.DBSettings{
		Replication: cloudflared1.ReadReplicationModeAuto,
	})
	assert.NoError(t, err)
	assert.NotNil(t, updateRes)
	assert.Len(t, updateRes.Errors, 0)
	assert.True(t, updateRes.Success)
	assert.Equal(t, newDBName, updateRes.Result.Name)
	assert.Equal(t, cloudflared1.ReadReplicationModeAuto, updateRes.Result.ReadReplication.Mode)

	// delete database
	deleteRes, err := client.DeleteDB(ctx, dbID.String())
	assert.NoError(t, err)
	assert.NotNil(t, deleteRes)
	assert.True(t, deleteRes.Success)
	assert.Len(t, deleteRes.Errors, 0)
}

func TestEditDBRecords(t *testing.T) {
	// create db for testing
	newDBName := fmt.Sprintf("test-%s", randomString(4))
	client, err := cloudflare_d1_go.NewClient(accountID, apiToken)
	ctx := context.Background()
	createResult, err := client.CreateDB(ctx, newDBName)
	assert.NoError(t, err)
	assert.NotNil(t, createResult)
	assert.Len(t, createResult.Errors, 0)
	assert.True(t, createResult.Success)
	assert.Equal(t, newDBName, createResult.Result.Name)
	dbID := createResult.Result.UUID
	assert.Equal(t, dbID.String(), client.NameIDMap[newDBName])

	if !createResult.Success {
		panic("Test cannot continue, db creation failed")
	}

	// create a new table
	sql := `
	CREATE TABLE IF NOT EXISTS test_table (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		verified INTEGER DEFAULT 0
	);`
	resp, err := client.QueryDB(ctx, dbID.String(), sql)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Errors, 0)

	// add row
	sql = `
	INSERT INTO 'test_table' ('name', 'verified') VALUES
    ('alice', 'TRUE'),
    ('bob', 'FALSE');`
	resp, err = client.QueryDB(ctx, dbID.String(), sql)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Errors, 0)

	// invalid insert
	sql = `
	INSERT INTO 'test_table' ('foo', 'bar', 'baz') VALUES
    ('alice', 'TRUE', 1),
    ('bob', 'FALSE', 9);`
	resp, err = client.QueryDB(ctx, dbID.String(), sql)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Len(t, resp.Errors, 1)

	// select
	sql = `SELECT * FROM test_table`
	resp, err = client.QueryDB(ctx, dbID.String(), sql)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Errors, 0)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Result, 1)
	assert.Len(t, resp.Result[0].Results, 2)
	type entry struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Verified string `json:"verified"`
	}
	entries := make([]entry, 2)
	resultsJSON, _ := json.Marshal(resp.Result[0].Results)
	err = json.Unmarshal(resultsJSON, &entries)
	assert.NoError(t, err)

	// raw select
	sql = `SELECT * FROM test_table`
	rawResp, err := client.QueryDBRaw(ctx, dbID.String(), sql)
	assert.NoError(t, err)
	assert.Len(t, rawResp.Errors, 0)
	assert.True(t, rawResp.Success)
	assert.NotNil(t, rawResp)
	j, _ := json.Marshal(resp.Result[0].Results)
	err = json.Unmarshal(j, &entries)
	assert.NoError(t, err)

	// remove table
	dropSQL := `DROP TABLE test_table`
	out, err := client.QueryDB(ctx, dbID.String(), dropSQL)
	assert.NoError(t, err)
	assert.Len(t, out.Errors, 0)
	assert.True(t, out.Success)
	assert.NotNil(t, out.Result)

	// delete database
	deleteRes, err := client.DeleteDB(ctx, dbID.String())
	assert.NoError(t, err)
	assert.NotNil(t, deleteRes)
	assert.True(t, deleteRes.Success)
	assert.Len(t, deleteRes.Errors, 0)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
