package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestNewClient(t *testing.T) {
	// Empty path
	_, err := NewMockClient("")
	assert.Error(t, err)
	// non empty path
	_, err = NewMockClient(".")
	assert.NoError(t, err)
}

func TestCreateDeleteUpdateDB(t *testing.T) {
	client, err := NewMockClient(".")
	ctx := context.Background()
	newDBName := fmt.Sprintf("test-%s", randomString(4))
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
	getResult, err := client.GetDB(ctx, dbID.String())
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
	client, err := NewMockClient(".")
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
	assert.Equal(t, entries[0].Name, "alice")
	assert.Equal(t, entries[1].Name, "bob")

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
	assert.Equal(t, entries[0].Name, "alice")
	assert.Equal(t, entries[1].Name, "bob")

	// delete items
	sql = `DELETE FROM test_table`
	resp, err = client.QueryDB(ctx, dbID.String(), sql)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Errors, 0)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Messages, 0)

	// select
	sql = `SELECT * FROM test_table`
	resp, err = client.QueryDB(ctx, dbID.String(), sql)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Errors, 0)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Result, 1)
	assert.Len(t, resp.Result[0].Results, 0)

	// remove table
	dropSQL := `DROP TABLE IF EXISTS test_table`
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
