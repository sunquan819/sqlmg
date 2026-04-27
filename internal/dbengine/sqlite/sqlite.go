package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	_ "modernc.org/sqlite"

	"sqlmg/internal/dbengine"
)

func init() {
	dbengine.Register("sqlite", func() dbengine.IDriver { return &SQLiteDriver{} })
}

type SQLiteDriver struct{}

func (d *SQLiteDriver) Name() string { return "sqlite" }

func (d *SQLiteDriver) Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("sqlite ping: %w", err)
	}
	return db, nil
}

func (d *SQLiteDriver) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

func (d *SQLiteDriver) Close(db *sql.DB) error {
	return db.Close()
}

func (d *SQLiteDriver) GetDatabases(ctx context.Context, db *sql.DB) ([]dbengine.Database, error) {
	return []dbengine.Database{{Name: "main"}}, nil
}

func (d *SQLiteDriver) GetSchemas(ctx context.Context, db *sql.DB, database string) ([]dbengine.Schema, error) {
	return []dbengine.Schema{{Name: "main"}}, nil
}

func (d *SQLiteDriver) GetTables(ctx context.Context, db *sql.DB, schema string) ([]dbengine.Table, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name, type FROM sqlite_master
		WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Table
	for rows.Next() {
		var t dbengine.Table
		var typ string
		if err := rows.Scan(&t.Name, &typ); err != nil {
			return nil, err
		}
		t.Schema = "main"
		if typ == "view" {
			t.Type = "VIEW"
		} else {
			t.Type = "TABLE"
		}
		result = append(result, t)
	}
	return result, nil
}

func (d *SQLiteDriver) GetColumns(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.Column, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(`%s`)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Column
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultVal sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultVal, &pk); err != nil {
			return nil, err
		}
		c := dbengine.Column{
			Name:      name,
			Type:      typ,
			Nullable:  notNull == 0,
			IsPrimary: pk > 0,
		}
		if defaultVal.Valid {
			c.DefaultValue = defaultVal.String
		}
		result = append(result, c)
	}
	return result, nil
}

func (d *SQLiteDriver) GetIndexes(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.Index, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA index_list(`%s`)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []struct {
		seq     int
		name    string
		unique  int
		origin  string
		partial int
	}
	for rows.Next() {
		var idx struct {
			seq     int
			name    string
			unique  int
			origin  string
			partial int
		}
		if err := rows.Scan(&idx.seq, &idx.name, &idx.unique, &idx.origin, &idx.partial); err != nil {
			return nil, err
		}
		indexes = append(indexes, idx)
	}

	var result []dbengine.Index
	for _, idx := range indexes {
		colRows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA index_info(`%s`)", idx.name))
		if err != nil {
			return nil, err
		}
		var cols []string
		for colRows.Next() {
			var seq, cid int
			var colName string
			if err := colRows.Scan(&seq, &cid, &colName); err != nil {
				colRows.Close()
				return nil, err
			}
			cols = append(cols, colName)
		}
		colRows.Close()

		if strings.HasPrefix(idx.name, "sqlite_autoindex_") {
			continue
		}

		result = append(result, dbengine.Index{
			Name:    idx.name,
			Columns: cols,
			Unique:  idx.unique == 1,
		})
	}
	return result, nil
}

func (d *SQLiteDriver) GetForeignKeys(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.ForeignKey, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA foreign_key_list(`%s`)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fkMap := make(map[string]*dbengine.ForeignKey)
	for rows.Next() {
		var id, seq int
		var refTable, from, to string
		var onDelete, onUpdate sql.NullString
		if err := rows.Scan(&id, &seq, &refTable, &from, &to, &onDelete, &onUpdate); err != nil {
			return nil, err
		}
		fkName := fmt.Sprintf("fk_%d", id)
		fk, ok := fkMap[fkName]
		if !ok {
			fk = &dbengine.ForeignKey{
				Name:     fkName,
				RefTable: refTable,
			}
			if onDelete.Valid {
				fk.OnDelete = onDelete.String
			}
			if onUpdate.Valid {
				fk.OnUpdate = onUpdate.String
			}
			fkMap[fkName] = fk
		}
		fk.Columns = append(fk.Columns, from)
		fk.RefColumns = append(fk.RefColumns, to)
	}

	var result []dbengine.ForeignKey
	for _, fk := range fkMap {
		result = append(result, *fk)
	}
	return result, nil
}

func (d *SQLiteDriver) GetDDL(ctx context.Context, db *sql.DB, schema, table string) (string, error) {
	var ddl string
	err := db.QueryRowContext(ctx, fmt.Sprintf(
		"SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", table,
	)).Scan(&ddl)
	if err != nil {
		return "", err
	}
	return ddl, nil
}

func (d *SQLiteDriver) Query(ctx context.Context, db *sql.DB, query string, args ...any) (*dbengine.ResultSet, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	columns := make([]dbengine.ColumnMeta, len(colTypes))
	for i, ct := range colTypes {
		columns[i] = dbengine.ColumnMeta{Name: ct.Name(), Type: ct.DatabaseTypeName()}
	}

	var resultRows []map[string]any
	for rows.Next() {
		values := make([]any, len(colTypes))
		ptrs := make([]any, len(colTypes))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(colTypes))
		for i, ct := range colTypes {
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				row[ct.Name()] = string(b)
			} else {
				row[ct.Name()] = val
			}
		}
		resultRows = append(resultRows, row)
	}

	if resultRows == nil {
		resultRows = []map[string]any{}
	}

	return &dbengine.ResultSet{
		Columns: columns,
		Rows:    resultRows,
		Total:   int64(len(resultRows)),
	}, nil
}

func (d *SQLiteDriver) QueryStream(ctx context.Context, db *sql.DB, query string, writer io.Writer) error {
	return fmt.Errorf("not implemented")
}

func (d *SQLiteDriver) Exec(ctx context.Context, db *sql.DB, query string, args ...any) (*dbengine.ExecResult, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	affected, _ := result.RowsAffected()
	lastID, _ := result.LastInsertId()
	return &dbengine.ExecResult{RowsAffected: affected, LastInsertID: lastID}, nil
}

func BuildDSN(path string) string {
	return path
}
