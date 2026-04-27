package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"sqlmg/internal/dbengine"
)

func init() {
	dbengine.Register("postgres", func() dbengine.IDriver { return &PostgresDriver{} })
}

type PostgresDriver struct{}

func (d *PostgresDriver) Name() string { return "postgres" }

func (d *PostgresDriver) Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return db, nil
}

func (d *PostgresDriver) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

func (d *PostgresDriver) Close(db *sql.DB) error {
	return db.Close()
}

func (d *PostgresDriver) GetDatabases(ctx context.Context, db *sql.DB) ([]dbengine.Database, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT datname FROM pg_database
		WHERE datistemplate = false AND datname NOT IN ('postgres')
		ORDER BY datname
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Database
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, dbengine.Database{Name: name})
	}
	return result, nil
}

func (d *PostgresDriver) GetSchemas(ctx context.Context, db *sql.DB, database string) ([]dbengine.Schema, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT schema_name FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY schema_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Schema
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, dbengine.Schema{Name: name})
	}
	return result, nil
}

func (d *PostgresDriver) GetTables(ctx context.Context, db *sql.DB, schema string) ([]dbengine.Table, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.QueryContext(ctx, `
		SELECT tablename, 'TABLE' as type
		FROM pg_tables WHERE schemaname = $1
		UNION ALL
		SELECT viewname, 'VIEW' as type
		FROM pg_views WHERE schemaname = $1
		ORDER BY name
	`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Table
	for rows.Next() {
		var t dbengine.Table
		if err := rows.Scan(&t.Name, &t.Type); err != nil {
			return nil, err
		}
		t.Schema = schema
		result = append(result, t)
	}
	return result, nil
}

func (d *PostgresDriver) GetColumns(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.Column, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.QueryContext(ctx, `
		SELECT c.column_name, c.data_type, c.is_nullable, c.column_default,
		       CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary,
		       c.column_comment
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku
			  ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
			  AND tc.table_schema = $1 AND tc.table_name = $2
		) pk ON c.column_name = pk.column_name
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Column
	for rows.Next() {
		var c dbengine.Column
		var nullable, defaultVal sql.NullString
		var isPrimary bool
		var comment sql.NullString
		if err := rows.Scan(&c.Name, &c.Type, &nullable, &defaultVal, &isPrimary, &comment); err != nil {
			return nil, err
		}
		c.Nullable = nullable.String == "YES"
		if defaultVal.Valid {
			c.DefaultValue = defaultVal.String
		}
		c.IsPrimary = isPrimary
		if comment.Valid {
			c.Comment = comment.String
		}
		result = append(result, c)
	}
	return result, nil
}

func (d *PostgresDriver) GetIndexes(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.Index, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.QueryContext(ctx, `
		SELECT i.relname as index_name,
		       array_to_string(ARRAY(
		         SELECT pg_get_indexdef(i.oid, k + 1, false)
		         FROM generate_subscripts(indkey, 1) k
		       ), ',') as columns,
		       indisunique as is_unique,
		       am.amname as index_type
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_am am ON am.oid = i.relam
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE n.nspname = $1 AND t.relname = $2
	`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Index
	for rows.Next() {
		var idx dbengine.Index
		var columns string
		var unique bool
		var idxType string
		if err := rows.Scan(&idx.Name, &columns, &unique, &idxType); err != nil {
			return nil, err
		}
		idx.Columns = strings.Split(columns, ",")
		idx.Unique = unique
		idx.Type = idxType
		result = append(result, idx)
	}
	return result, nil
}

func (d *PostgresDriver) GetForeignKeys(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.ForeignKey, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.QueryContext(ctx, `
		SELECT
			tc.constraint_name,
			string_agg(DISTINCT kcu.column_name, ',') AS columns,
			ccu.table_name AS ref_table,
			string_agg(DISTINCT ccu.column_name, ',') AS ref_columns,
			rc.delete_rule, rc.update_rule
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
		  ON tc.constraint_name = ccu.constraint_name AND tc.table_schema = ccu.table_schema
		JOIN information_schema.referential_constraints rc
		  ON tc.constraint_name = rc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = $1 AND tc.table_name = $2
		GROUP BY tc.constraint_name, ccu.table_name, rc.delete_rule, rc.update_rule
	`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.ForeignKey
	for rows.Next() {
		var fk dbengine.ForeignKey
		var cols, refCols string
		if err := rows.Scan(&fk.Name, &cols, &fk.RefTable, &refCols, &fk.OnDelete, &fk.OnUpdate); err != nil {
			return nil, err
		}
		fk.Columns = strings.Split(cols, ",")
		fk.RefColumns = strings.Split(refCols, ",")
		result = append(result, fk)
	}
	return result, nil
}

func (d *PostgresDriver) GetDDL(ctx context.Context, db *sql.DB, schema, table string) (string, error) {
	var ddl string
	err := db.QueryRowContext(ctx, fmt.Sprintf(
		"SELECT pg_get_tabledef('%s.%s')", schema, table,
	)).Scan(&ddl)
	if err != nil {
		return "", fmt.Errorf("pg_get_tabledef not available: %w", err)
	}
	return ddl, nil
}

func (d *PostgresDriver) Query(ctx context.Context, db *sql.DB, query string, args ...any) (*dbengine.ResultSet, error) {
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

func (d *PostgresDriver) QueryStream(ctx context.Context, db *sql.DB, query string, writer io.Writer) error {
	return fmt.Errorf("not implemented")
}

func (d *PostgresDriver) Exec(ctx context.Context, db *sql.DB, query string, args ...any) (*dbengine.ExecResult, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	affected, _ := result.RowsAffected()
	return &dbengine.ExecResult{RowsAffected: affected}, nil
}

func BuildDSN(host string, port int, username, password, database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=prefer", username, password, host, port, database)
}
