package mock

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cloudflared1 "github.com/crosleyzack/cloudflare-d1-go"
	"github.com/crosleyzack/cloudflare-d1-go/utils"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

const (
	dbtype = "sqlite"
)

type MockClient struct {
	dbpath    string
	NameIDMap map[string]string
	ConnMap   map[string]*sql.DB
}

var _ cloudflared1.CloudflareD1 = (*MockClient)(nil)

// NewMockClient creates a new client for interfacing with local sqlite
func NewMockClient(dbpath string) (*MockClient, error) {
	if len(dbpath) == 0 {
		return nil, errors.New("DBPath cannot be empty")
	}
	p, err := filepath.Abs(dbpath)
	if err != nil {
		return nil, err
	}
	return &MockClient{
		dbpath:    p,
		NameIDMap: map[string]string{},
		ConnMap:   map[string]*sql.DB{},
	}, nil
}

func (m *MockClient) Close() error {
	for _, conn := range m.ConnMap {
		err := conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateDB create a new database in the local sqlite database
func (m *MockClient) CreateDB(ctx context.Context, name string) (*utils.APIResponse[cloudflared1.D1Database], error) {
	uid := uuid.New()
	path := m.getDBPath(uid.String())
	db, err := sql.Open(dbtype, path)
	if err != nil {
		return nil, err
	}
	// Mock implementation - create a basic D1Database response
	database := cloudflared1.D1Database{
		CreatedAt: time.Now().Format(time.RFC3339),
		FileSize:  0,
		Name:      name,
		NumTables: 0,
		ReadReplication: cloudflared1.ReadReplication{
			Mode: cloudflared1.ReadReplicationModeDisabled,
		},
		UUID:    uid,
		Version: "1.0.0",
	}
	m.NameIDMap[name] = uid.String()
	m.ConnMap[uid.String()] = db

	return &utils.APIResponse[cloudflared1.D1Database]{
		Result:  database,
		Success: true,
		Errors:  nil,
	}, nil
}

// DeleteDB delete a new database in the local sqlite database by id
func (m *MockClient) DeleteDB(_ context.Context, dbID string) (*utils.APIResponse[cloudflared1.DeleteResult], error) {
	conn, ok := m.ConnMap[dbID]
	if ok {
		delete(m.ConnMap, dbID)
	}
	if err := os.Remove(m.getDBPath(dbID)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	defer conn.Close()
	return &utils.APIResponse[cloudflared1.DeleteResult]{
		Result:  cloudflared1.DeleteResult{},
		Success: true,
		Errors:  nil,
	}, nil
}

// UpdateDB fulfills interface however there is no relevant operation for a local sqlite
func (m *MockClient) UpdateDB(_ context.Context, dbID string, settings cloudflared1.DBSettings) (*utils.APIResponse[cloudflared1.D1Database], error) {
	return &utils.APIResponse[cloudflared1.D1Database]{
		Result:  cloudflared1.D1Database{},
		Success: false,
		Errors: []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			{
				Code:    1000,
				Message: "There is no concept of replication modes in local sqlite",
			},
		},
	}, nil
}

// GetDB Retrieve database information for local sqlite db
func (m *MockClient) GetDB(_ context.Context, dbID string) (*utils.APIResponse[cloudflared1.D1Database], error) {
	db, ok := m.ConnMap[dbID]
	if !ok {
		return nil, fmt.Errorf("Invalid db id: %s", dbID)
	}
	var dbSize int64
	var dbTime string
	if fi, err := os.Stat(m.getDBPath(dbID)); err == nil {
		dbTime = fi.ModTime().Format(time.RFC3339)
		dbSize = fi.Size()
	}
	// get uuid from database id for return
	uid, err := uuid.Parse(dbID)
	if err != nil {
		return nil, err
	}
	// get name out of the map
	dbname := ""
	for name, id := range m.NameIDMap {
		if id == dbID {
			dbname = name
			break
		}
	}
	// get tables in database
	rows, err := db.Query("SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name != 'android_metadata' AND name != 'sqlite_sequence';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return nil, err
		}
	}
	// Mock implementation - return a mock database
	database := cloudflared1.D1Database{
		CreatedAt: dbTime,
		FileSize:  dbSize,
		Name:      dbname,
		NumTables: count,
		ReadReplication: cloudflared1.ReadReplication{
			Mode: cloudflared1.ReadReplicationModeDisabled,
		},
		UUID:    uid,
		Version: "1.0.0",
	}

	return &utils.APIResponse[cloudflared1.D1Database]{
		Result:  database,
		Success: true,
		Errors:  nil,
	}, nil
}

func (m *MockClient) ListDB(ctx context.Context) (*utils.APIResponse[cloudflared1.D1DatabaseList], error) {
	l := make(cloudflared1.D1DatabaseList, 0, len(m.NameIDMap))
	for _, id := range m.NameIDMap {
		res, err := m.GetDB(ctx, id)
		if err != nil {
			// just skip this one
			continue
		}
		l = append(l, res.Result)
	}

	return &utils.APIResponse[cloudflared1.D1DatabaseList]{
		Result:  l,
		Success: true,
		Errors:  nil,
	}, nil
}

// QueryDB execute a query on the local sqlite db
func (m *MockClient) QueryDB(ctx context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]cloudflared1.QueryResult[any]], error) {
	// local sqlite db separates operations that retrieve and alter data.
	// check which we are doing and perform the appropriate operation
	if strings.Contains(strings.ToLower(query), "select") {
		return m.query(ctx, dbID, query, params...)
	} else {
		return m.exec(ctx, dbID, query, params...)
	}
}

// query helper to retrieve information from the local sql db
func (m *MockClient) query(_ context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]cloudflared1.QueryResult[any]], error) {
	db, ok := m.ConnMap[dbID]
	if !ok {
		return nil, fmt.Errorf("Invalid db id: %s", dbID)
	}
	rows, err := db.Query(query, params...)
	if err != nil {
		return &utils.APIResponse[[]cloudflared1.QueryResult[any]]{
			Result:  []cloudflared1.QueryResult[any]{},
			Success: false,
			Errors: []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{{Code: 1000, Message: err.Error()}},
		}, nil
	}
	defer rows.Close()

	// Convert rows to result format
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []any
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Convert to map with column names as keys
		row := make(map[string]interface{})
		for i, col := range columns {
			// Convert []byte to string for better JSON marshaling
			if b, ok := values[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = values[i]
			}
		}
		results = append(results, row)
	}

	meta := cloudflared1.Meta{
		ChangedDB:       false,
		Changes:         0,
		Duration:        0.1,
		LastRowID:       0,
		RowsRead:        len(results),
		RowsWritten:     0,
		ServedByPrimary: true,
		ServedByRegion:  "mock-region",
		SizeAfter:       1024,
		Timings: &cloudflared1.Timings{
			SQLDurationMS: 0.1,
		},
	}

	queryResult := cloudflared1.QueryResult[any]{
		Meta:    meta,
		Results: results,
		Success: true,
	}

	return &utils.APIResponse[[]cloudflared1.QueryResult[any]]{
		Result:  []cloudflared1.QueryResult[any]{queryResult},
		Success: true,
		Errors:  nil,
	}, nil
}

// exec helper to alter the local sql db
func (m *MockClient) exec(_ context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]cloudflared1.QueryResult[any]], error) {
	db, ok := m.ConnMap[dbID]
	if !ok {
		return nil, fmt.Errorf("Invalid db id: %s", dbID)
	}
	result, err := db.Exec(query, params...)
	if err != nil {
		return &utils.APIResponse[[]cloudflared1.QueryResult[any]]{
			Result:  []cloudflared1.QueryResult[any]{},
			Success: false,
			Errors: []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{{Code: 1000, Message: err.Error()}},
		}, nil
	}
	last, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	meta := cloudflared1.Meta{
		ChangedDB:       false,
		Changes:         0,
		Duration:        0.1,
		LastRowID:       last,
		RowsRead:        int(affected),
		RowsWritten:     int(affected),
		ServedByPrimary: true,
		ServedByRegion:  "mock-region",
		SizeAfter:       1024,
		Timings: &cloudflared1.Timings{
			SQLDurationMS: 0.1,
		},
	}

	queryResult := cloudflared1.QueryResult[any]{
		Meta:    meta,
		Results: nil,
		Success: true,
	}

	return &utils.APIResponse[[]cloudflared1.QueryResult[any]]{
		Result:  []cloudflared1.QueryResult[any]{queryResult},
		Success: true,
		Errors:  nil,
	}, nil
}

// QueryDBRaw execute a query on the local sqlite db
func (m *MockClient) QueryDBRaw(ctx context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]cloudflared1.QueryResult[any]], error) {
	// For mock, QueryDBRaw behaves the same as QueryDB
	return m.QueryDB(ctx, dbID, query, params...)
}

func (m *MockClient) getDBPath(id string) string {
	return filepath.Join(m.dbpath, fmt.Sprintf("%s.db", id))
}
