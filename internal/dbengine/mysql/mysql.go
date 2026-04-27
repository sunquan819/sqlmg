package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"sqlmg/internal/dbengine"
)

func init() {
	dbengine.Register("mysql", func() dbengine.IDriver { return &MySQLDriver{} })
}

type MySQLDriver struct{}

func (d *MySQLDriver) Name() string { return "mysql" }

func (d *MySQLDriver) Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	if !strings.Contains(dsn, "?") {
		dsn += "?parseTime=true&charset=utf8mb4&loc=Local"
	} else {
		if !strings.Contains(dsn, "parseTime") {
			dsn += "&parseTime=true"
		}
		if !strings.Contains(dsn, "charset") {
			dsn += "&charset=utf8mb4"
		}
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mysql ping: %w", err)
	}
	return db, nil
}

func (d *MySQLDriver) Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

func (d *MySQLDriver) Close(db *sql.DB) error {
	return db.Close()
}

func (d *MySQLDriver) GetDatabases(ctx context.Context, db *sql.DB) ([]dbengine.Database, error) {
	rows, err := db.QueryContext(ctx, "SHOW DATABASES")
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
		if name == "information_schema" || name == "performance_schema" || name == "mysql" || name == "sys" {
			continue
		}
		result = append(result, dbengine.Database{Name: name})
	}
	return result, nil
}

func (d *MySQLDriver) GetSchemas(ctx context.Context, db *sql.DB, database string) ([]dbengine.Schema, error) {
	return []dbengine.Schema{{Name: database}}, nil
}

func (d *MySQLDriver) GetTables(ctx context.Context, db *sql.DB, schema string) ([]dbengine.Table, error) {
	if schema != "" {
		if _, err := db.ExecContext(ctx, "USE `"+schema+"`"); err != nil {
			return nil, err
		}
	}

	rows, err := db.QueryContext(ctx, `
		SELECT TABLE_NAME, TABLE_TYPE, TABLE_COMMENT, TABLE_ROWS
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Table
	for rows.Next() {
		var t dbengine.Table
		var tableType, comment string
		var rowCount sql.NullInt64
		if err := rows.Scan(&t.Name, &tableType, &comment, &rowCount); err != nil {
			return nil, err
		}
		t.Schema = schema
		if strings.Contains(tableType, "VIEW") {
			t.Type = "VIEW"
		} else {
			t.Type = "TABLE"
		}
		t.Comment = comment
		if rowCount.Valid {
			t.RowCount = rowCount.Int64
		}
		result = append(result, t)
	}
	return result, nil
}

func (d *MySQLDriver) GetColumns(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.Column, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_KEY,
		       EXTRA, COLUMN_COMMENT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Column
	for rows.Next() {
		var c dbengine.Column
		var nullable, defaultVal, key, extra, comment string
		if err := rows.Scan(&c.Name, &c.Type, &nullable, &defaultVal, &key, &extra, &comment); err != nil {
			return nil, err
		}
		c.Nullable = nullable == "YES"
		c.DefaultValue = defaultVal
		c.IsPrimary = key == "PRI"
		c.AutoIncrement = strings.Contains(extra, "auto_increment")
		c.Comment = comment
		result = append(result, c)
	}
	return result, nil
}

func (d *MySQLDriver) GetIndexes(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.Index, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT INDEX_NAME, GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) AS columns,
		       NON_UNIQUE, INDEX_TYPE
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		GROUP BY INDEX_NAME, NON_UNIQUE, INDEX_TYPE
	`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dbengine.Index
	for rows.Next() {
		var idx dbengine.Index
		var columns string
		var nonUnique int
		var idxType string
		if err := rows.Scan(&idx.Name, &columns, &nonUnique, &idxType); err != nil {
			return nil, err
		}
		idx.Columns = strings.Split(columns, ",")
		idx.Unique = nonUnique == 0
		idx.Type = idxType
		result = append(result, idx)
	}
	return result, nil
}

func (d *MySQLDriver) GetForeignKeys(ctx context.Context, db *sql.DB, schema, table string) ([]dbengine.ForeignKey, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT kcu.CONSTRAINT_NAME,
		       GROUP_CONCAT(kcu.COLUMN_NAME ORDER BY kcu.ORDINAL_POSITION),
		       kcu.REFERENCED_TABLE_NAME,
		       GROUP_CONCAT(kcu.REFERENCED_COLUMN_NAME ORDER BY kcu.ORDINAL_POSITION),
		       rc.DELETE_RULE, rc.UPDATE_RULE
		FROM information_schema.KEY_COLUMN_USAGE kcu
		JOIN information_schema.REFERENTIAL_CONSTRAINTS rc
		  ON kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME
		 AND kcu.TABLE_SCHEMA = rc.CONSTRAINT_SCHEMA
		WHERE kcu.TABLE_SCHEMA = ? AND kcu.TABLE_NAME = ?
		  AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
		GROUP BY kcu.CONSTRAINT_NAME, kcu.REFERENCED_TABLE_NAME, rc.DELETE_RULE, rc.UPDATE_RULE
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

func (d *MySQLDriver) GetDDL(ctx context.Context, db *sql.DB, schema, table string) (string, error) {
	var ddl string
	err := db.QueryRowContext(ctx, "SHOW CREATE TABLE `"+schema+"`.`"+table+"`").Scan(nil, &ddl)
	if err != nil {
		return "", err
	}
	return ddl, nil
}

func (d *MySQLDriver) Query(ctx context.Context, db *sql.DB, query string, args ...any) (*dbengine.ResultSet, error) {
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

func (d *MySQLDriver) QueryStream(ctx context.Context, db *sql.DB, query string, writer io.Writer) error {
	return fmt.Errorf("not implemented")
}

func (d *MySQLDriver) Exec(ctx context.Context, db *sql.DB, query string, args ...any) (*dbengine.ExecResult, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	affected, _ := result.RowsAffected()
	lastID, _ := result.LastInsertId()
	return &dbengine.ExecResult{RowsAffected: affected, LastInsertID: lastID}, nil
}

func BuildDSN(host string, port int, username, password, database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
}
