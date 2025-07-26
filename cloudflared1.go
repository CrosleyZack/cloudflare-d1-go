package cloudflared1

import (
	"context"
	"encoding/json"

	"github.com/crosleyzack/cloudflare-d1-go/utils"
	"github.com/google/uuid"
)

// https://developers.cloudflare.com/api/resources/d1/subresources/database/
type CloudflareD1 interface {
	CreateDB(ctx context.Context, dbName string) (*utils.APIResponse[D1Database], error)
	DeleteDB(ctx context.Context, dbID string) (*utils.APIResponse[DeleteResult], error)
	UpdateDB(ctx context.Context, dbID string, settings DBSettings) (*utils.APIResponse[D1Database], error)
	GetDB(ctx context.Context, dbID string) (*utils.APIResponse[D1Database], error)
	ListDB(ctx context.Context) (*utils.APIResponse[D1DatabaseList], error)
	// NOTE: It would be awesome to template this interface so we could specify the exact format of the data expected with each call.
	// unforutnately, that is not allowed in golang
	QueryDB(ctx context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]QueryResult[any]], error)
	QueryDBRaw(ctx context.Context, dbID string, query string, params ...any) (*utils.APIResponse[[]QueryResult[any]], error)
}

type ReadReplicationMode int

const (
	ReadReplicationModeAuto ReadReplicationMode = iota
	ReadReplicationModeDisabled
)

func (r ReadReplicationMode) String() string {
	switch r {
	case ReadReplicationModeAuto:
		return "auto"
	default:
		return "disabled"
	}
}

func (r ReadReplicationMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *ReadReplicationMode) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "auto":
		*r = ReadReplicationModeAuto
	default:
		*r = ReadReplicationModeDisabled
	}
	return nil
}

type ReadReplication struct {
	Mode ReadReplicationMode `json:"mode"`
}

type DBSettings struct {
	Replication ReadReplicationMode `json:"replication"`
}

type D1Database struct {
	CreatedAt       string          `json:"created_at"`
	FileSize        int64           `json:"file_size"`
	Name            string          `json:"name"`
	NumTables       int             `json:"num_tables"`
	ReadReplication ReadReplication `json:"read_replication"`
	UUID            uuid.UUID       `json:"uuid"`
	Version         string          `json:"version"`
}

type Meta struct {
	ChangedDB       bool     `json:"changed_db"`
	Changes         int      `json:"changes"`
	Duration        float64  `json:"duration"`
	LastRowID       int64    `json:"last_row_id"`
	RowsRead        int      `json:"rows_read"`
	RowsWritten     int      `json:"rows_written"`
	ServedByPrimary bool     `json:"served_by_primary"`
	ServedByRegion  string   `json:"served_by_region"`
	SizeAfter       int64    `json:"size_after"`
	Timings         *Timings `json:"timings"`
}

type Timings struct {
	SQLDurationMS float64 `json:"sql_duration_ms"`
}

type QueryResult[T any] struct {
	Meta    Meta `json:"meta"`
	Results T    `json:"results"`
	Success bool `json:"success"`
}

type DeleteResult struct{}

type D1DatabaseList []D1Database
