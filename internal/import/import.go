package imp

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"sqlmg/internal/dbengine"
)

type ImportFormat string

const (
	FormatCSV  ImportFormat = "csv"
	FormatJSON ImportFormat = "json"
)

type ImportOptions struct {
	Format      ImportFormat        `json:"format"`
	Schema      string              `json:"schema"`
	Table       string              `json:"table"`
	Delimiter   string              `json:"delimiter,omitempty"`
	HasHeader   bool                `json:"hasHeader"`
	Encoding    string              `json:"encoding,omitempty"`
	ColumnMap   map[string]string   `json:"columnMap,omitempty"`
	BatchSize   int                 `json:"batchSize"`
	OnError     string              `json:"onError,omitempty"`
}

type ImportResult struct {
	TotalRows    int      `json:"totalRows"`
	InsertedRows int      `json:"insertedRows"`
	FailedRows   int      `json:"failedRows"`
	Errors       []string `json:"errors,omitempty"`
}

func ParseCSV(reader io.Reader, delimiter string, hasHeader bool) (headers []string, rows [][]string, err error) {
	if delimiter == "" {
		delimiter = ","
	}
	r := csv.NewReader(reader)
	r.Comma = []rune(delimiter)[0]
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	if hasHeader {
		headers, err = r.Read()
		if err != nil {
			return nil, nil, fmt.Errorf("读取CSV表头失败: %w", err)
		}
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return headers, rows, fmt.Errorf("读取CSV行失败: %w", err)
		}
		rows = append(rows, record)
	}

	if !hasHeader && len(rows) > 0 {
		headers = make([]string, len(rows[0]))
		for i := range headers {
			headers[i] = fmt.Sprintf("col_%d", i+1)
		}
	}

	return headers, rows, nil
}

func ParseJSON(reader io.Reader) ([]string, []map[string]any, error) {
	var data []map[string]any
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&data); err != nil {
		return nil, nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if len(data) == 0 {
		return nil, nil, fmt.Errorf("JSON数据为空")
	}

	headerSet := make(map[string]bool)
	for _, row := range data {
		for k := range row {
			headerSet[k] = true
		}
	}

	headers := make([]string, 0, len(headerSet))
	for k := range headerSet {
		headers = append(headers, k)
	}

	return headers, data, nil
}

func ImportCSV(db *sql.DB, driver dbengine.IDriver, reader io.Reader, opts ImportOptions) (*ImportResult, error) {
	headers, rows, err := ParseCSV(reader, opts.Delimiter, opts.HasHeader)
	if err != nil {
		return nil, err
	}

	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}

	result := &ImportResult{TotalRows: len(rows)}

	columnMap := opts.ColumnMap
	if columnMap == nil {
		columnMap = make(map[string]string)
	}

	targetColumns := make([]string, len(headers))
	for i, h := range headers {
		if mapped, ok := columnMap[h]; ok && mapped != "" {
			targetColumns[i] = mapped
		} else {
			targetColumns[i] = h
		}
	}

	colPlaceholders := make([]string, len(headers))
	for i := range colPlaceholders {
		colPlaceholders[i] = "?"
	}

	quotedCols := make([]string, len(targetColumns))
	for i, c := range targetColumns {
		quotedCols[i] = fmt.Sprintf("`%s`", c)
	}

	insertSQL := fmt.Sprintf(
		"INSERT INTO `%s`.`%s` (%s) VALUES (%s)",
		opts.Schema, opts.Table,
		strings.Join(quotedCols, ", "),
		strings.Join(colPlaceholders, ", "),
	)

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return nil, fmt.Errorf("准备插入语句失败: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rowIdx, row := range rows {
		vals := make([]any, len(headers))
		for i := range headers {
			if i < len(row) {
				vals[i] = row[i]
			} else {
				vals[i] = nil
			}
		}

		_, err := stmt.Exec(vals...)
		if err != nil {
			result.FailedRows++
			errMsg := fmt.Sprintf("第%d行: %s", rowIdx+1, err.Error())
			result.Errors = append(result.Errors, errMsg)
			if opts.OnError == "abort" {
				break
			}
			continue
		}

		result.InsertedRows++
		count++

		if count >= opts.BatchSize {
			if err := tx.Commit(); err != nil {
				return result, fmt.Errorf("提交事务失败: %w", err)
			}
			tx, err = db.Begin()
			if err != nil {
				return result, err
			}
			stmt, err = tx.Prepare(insertSQL)
			if err != nil {
				return result, err
			}
			count = 0
		}
	}

	if count > 0 {
		if err := tx.Commit(); err != nil {
			return result, fmt.Errorf("提交事务失败: %w", err)
		}
	}

	return result, nil
}

func ImportJSON(db *sql.DB, driver dbengine.IDriver, reader io.Reader, opts ImportOptions) (*ImportResult, error) {
	headers, data, err := ParseJSON(reader)
	if err != nil {
		return nil, err
	}

	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}

	result := &ImportResult{TotalRows: len(data)}

	quotedCols := make([]string, len(headers))
	placeholders := make([]string, len(headers))
	for i, h := range headers {
		mapped := h
		if opts.ColumnMap != nil {
			if m, ok := opts.ColumnMap[h]; ok && m != "" {
				mapped = m
			}
		}
		quotedCols[i] = fmt.Sprintf("`%s`", mapped)
		placeholders[i] = "?"
	}

	insertSQL := fmt.Sprintf(
		"INSERT INTO `%s`.`%s` (%s) VALUES (%s)",
		opts.Schema, opts.Table,
		strings.Join(quotedCols, ", "),
		strings.Join(placeholders, ", "),
	)

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return nil, fmt.Errorf("准备插入语句失败: %w", err)
	}
	defer stmt.Close()

	count := 0
	for rowIdx, row := range data {
		vals := make([]any, len(headers))
		for i, h := range headers {
			val := row[h]
			if b, ok := val.([]byte); ok {
				vals[i] = string(b)
			} else {
				vals[i] = val
			}
		}

		_, err := stmt.Exec(vals...)
		if err != nil {
			result.FailedRows++
			errMsg := fmt.Sprintf("第%d行: %s", rowIdx+1, err.Error())
			result.Errors = append(result.Errors, errMsg)
			if opts.OnError == "abort" {
				break
			}
			continue
		}

		result.InsertedRows++
		count++

		if count >= opts.BatchSize {
			if err := tx.Commit(); err != nil {
				return result, fmt.Errorf("提交事务失败: %w", err)
			}
			tx, err = db.Begin()
			if err != nil {
				return result, err
			}
			stmt, err = tx.Prepare(insertSQL)
			if err != nil {
				return result, err
			}
			count = 0
		}
	}

	if count > 0 {
		if err := tx.Commit(); err != nil {
			return result, fmt.Errorf("提交事务失败: %w", err)
		}
	}

	return result, nil
}

func StrToInt(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
