package export

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"sqlmg/internal/dbengine"
)

type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatJSON ExportFormat = "json"
	FormatSQL  ExportFormat = "sql"
)

type ExportOptions struct {
	Format      ExportFormat `json:"format"`
	Schema      string       `json:"schema"`
	Table       string       `json:"table"`
	Query       string       `json:"query,omitempty"`
	Delimiter   string       `json:"delimiter,omitempty"`
	IncludeHeader bool       `json:"includeHeader"`
	Encoding    string       `json:"encoding,omitempty"`
	InsertBatch int          `json:"insertBatch,omitempty"`
}

func ExportCSV(writer io.Writer, columns []dbengine.ColumnMeta, rows []map[string]any, delimiter string, includeHeader bool) error {
	if delimiter == "" {
		delimiter = ","
	}
	w := csv.NewWriter(writer)
	w.Comma = []rune(delimiter)[0]

	if includeHeader {
		header := make([]string, len(columns))
		for i, col := range columns {
			header[i] = col.Name
		}
		if err := w.Write(header); err != nil {
			return err
		}
	}

	for _, row := range rows {
		record := make([]string, len(columns))
		for i, col := range columns {
			val := row[col.Name]
			if val == nil {
				record[i] = "NULL"
			} else {
				record[i] = fmt.Sprintf("%v", val)
			}
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}

func ExportJSON(writer io.Writer, columns []dbengine.ColumnMeta, rows []map[string]any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	result := make([]map[string]any, len(rows))
	for i, row := range rows {
		cleanRow := make(map[string]any, len(columns))
		for _, col := range columns {
			val := row[col.Name]
			if b, ok := val.([]byte); ok {
				cleanRow[col.Name] = string(b)
			} else {
				cleanRow[col.Name] = val
			}
		}
		result[i] = cleanRow
	}

	return encoder.Encode(result)
}

func ExportSQL(writer io.Writer, columns []dbengine.ColumnMeta, rows []map[string]any, schema, table string, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	fmt.Fprintf(writer, "-- SQLMG Export: %s.%s\n", schema, table)
	fmt.Fprintf(writer, "-- Total rows: %d\n\n", len(rows))

	colNames := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = fmt.Sprintf("`%s`", col.Name)
	}
	colList := strings.Join(colNames, ", ")

	batch := make([]string, 0, batchSize)
	for rowIdx, row := range rows {
		vals := make([]string, len(columns))
		for i, col := range columns {
			val := row[col.Name]
			if val == nil {
				vals[i] = "NULL"
			} else {
				switch v := val.(type) {
				case string:
					escaped := strings.ReplaceAll(v, "'", "''")
					vals[i] = fmt.Sprintf("'%s'", escaped)
				case []byte:
					escaped := strings.ReplaceAll(string(v), "'", "''")
					vals[i] = fmt.Sprintf("'%s'", escaped)
				case int, int32, int64, float32, float64:
					vals[i] = fmt.Sprintf("%v", v)
				case bool:
					if v {
						vals[i] = "1"
					} else {
						vals[i] = "0"
					}
				default:
					escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
					vals[i] = fmt.Sprintf("'%s'", escaped)
				}
			}
		}
		batch = append(batch, fmt.Sprintf("(%s)", strings.Join(vals, ", ")))

		if len(batch) >= batchSize || rowIdx == len(rows)-1 {
			fmt.Fprintf(writer, "INSERT INTO `%s`.`%s` (%s) VALUES\n", schema, table, colList)
			fmt.Fprintf(writer, "  %s;\n\n", strings.Join(batch, ",\n  "))
			batch = batch[:0]
		}
	}

	return nil
}

func ExportQueryResult(ctx context.Context, db *sql.DB, driver dbengine.IDriver, query string, writer io.Writer, opts ExportOptions) error {
	result, err := driver.Query(ctx, db, query)
	if err != nil {
		return err
	}

	switch opts.Format {
	case FormatCSV:
		return ExportCSV(writer, result.Columns, result.Rows, opts.Delimiter, opts.IncludeHeader)
	case FormatJSON:
		return ExportJSON(writer, result.Columns, result.Rows)
	case FormatSQL:
		return ExportSQL(writer, result.Columns, result.Rows, opts.Schema, opts.Table, opts.InsertBatch)
	default:
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

func DetectBatchSize(totalRows int) int {
	switch {
	case totalRows <= 100:
		return 50
	case totalRows <= 1000:
		return 100
	case totalRows <= 10000:
		return 500
	default:
		return 1000
	}
}

func SanitizeFilename(name string) string {
	replaced := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			return r
		}
		return '_'
	}, name)
	if len(replaced) > 200 {
		replaced = replaced[:200]
	}
	return replaced
}

func ParseInt(s string, defaultVal int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}
