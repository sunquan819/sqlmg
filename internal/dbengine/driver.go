package dbengine

import (
	"context"
	"database/sql"
	"io"
)

type ColumnType struct {
	DatabaseType string
	Name         string
	Nullable     bool
	Length       int64
	Precision    int64
	Scale        int64
}

type Column struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"defaultValue,omitempty"`
	IsPrimary    bool   `json:"isPrimary"`
	AutoIncrement bool  `json:"autoIncrement,omitempty"`
	Comment      string `json:"comment,omitempty"`
}

type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type,omitempty"`
}

type ForeignKey struct {
	Name           string `json:"name"`
	Columns        []string `json:"columns"`
	RefTable       string `json:"refTable"`
	RefColumns     []string `json:"refColumns"`
	OnDelete       string `json:"onDelete,omitempty"`
	OnUpdate       string `json:"onUpdate,omitempty"`
}

type Table struct {
	Name    string `json:"name"`
	Schema  string `json:"schema,omitempty"`
	Type    string `json:"type"`
	Comment string `json:"comment,omitempty"`
	RowCount int64 `json:"rowCount,omitempty"`
}

type Database struct {
	Name string `json:"name"`
}

type Schema struct {
	Name string `json:"name"`
}

type ResultSet struct {
	Columns []ColumnMeta     `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Total   int64            `json:"total"`
}

type ColumnMeta struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ExecResult struct {
	RowsAffected int64 `json:"rowsAffected"`
	LastInsertID int64 `json:"lastInsertId,omitempty"`
}

type DDLResult struct {
	DDL string `json:"ddl"`
}

type IDriver interface {
	Name() string
	Connect(ctx context.Context, dsn string) (*sql.DB, error)
	Ping(ctx context.Context, db *sql.DB) error
	Close(db *sql.DB) error

	GetDatabases(ctx context.Context, db *sql.DB) ([]Database, error)
	GetSchemas(ctx context.Context, db *sql.DB, database string) ([]Schema, error)
	GetTables(ctx context.Context, db *sql.DB, schema string) ([]Table, error)
	GetColumns(ctx context.Context, db *sql.DB, schema, table string) ([]Column, error)
	GetIndexes(ctx context.Context, db *sql.DB, schema, table string) ([]Index, error)
	GetForeignKeys(ctx context.Context, db *sql.DB, schema, table string) ([]ForeignKey, error)
	GetDDL(ctx context.Context, db *sql.DB, schema, table string) (string, error)

	Query(ctx context.Context, db *sql.DB, query string, args ...any) (*ResultSet, error)
	QueryStream(ctx context.Context, db *sql.DB, query string, writer io.Writer) error
	Exec(ctx context.Context, db *sql.DB, query string, args ...any) (*ExecResult, error)
}

type DriverFactory func() IDriver

var registry = map[string]DriverFactory{}

func Register(name string, factory DriverFactory) {
	registry[name] = factory
}

func Get(name string) (IDriver, bool) {
	factory, ok := registry[name]
	if !ok {
		return nil, false
	}
	return factory(), true
}

func Drivers() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
