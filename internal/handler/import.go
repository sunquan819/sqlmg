package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"sqlmg/internal/dbengine"
	imp "sqlmg/internal/import"
	"sqlmg/internal/metastore"
)

type ImportHandler struct {
	manager *dbengine.Manager
	store   *metastore.Store
}

func NewImportHandler(manager *dbengine.Manager, store *metastore.Store) *ImportHandler {
	return &ImportHandler{manager: manager, store: store}
}

func (h *ImportHandler) getSession(c *gin.Context) (*dbengine.Session, bool) {
	connID := c.Param("id")
	session, ok := h.manager.GetSession(connID)
	if !ok {
		var err error
		session, err = h.manager.Connect(c.Request.Context(), connID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return nil, false
		}
	}
	return session, true
}

func (h *ImportHandler) ImportFile(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	schema := c.PostForm("schema")
	table := c.PostForm("table")
	format := c.PostForm("format")
	delimiter := c.PostForm("delimiter")
	hasHeader := c.PostForm("hasHeader") == "true"
	batchSize := imp.StrToInt(c.PostForm("batchSize"), 100)
	onError := c.PostForm("onError")
	columnMapJSON := c.PostForm("columnMap")

	if schema == "" {
		schema = "main"
	}
	if format == "" {
		format = "csv"
	}

	var columnMap map[string]string
	if columnMapJSON != "" {
		if err := json.Unmarshal([]byte(columnMapJSON), &columnMap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("列映射JSON无效: %s", err.Error())})
			return
		}
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("读取上传文件失败: %s", err.Error())})
		return
	}
	defer file.Close()

	opts := imp.ImportOptions{
		Format:    imp.ImportFormat(format),
		Schema:    schema,
		Table:     table,
		Delimiter: delimiter,
		HasHeader: hasHeader,
		BatchSize: batchSize,
		OnError:   onError,
		ColumnMap: columnMap,
	}

	var result *imp.ImportResult

	switch format {
	case "csv":
		result, err = imp.ImportCSV(session.DB, session.Driver, file, opts)
	case "json":
		result, err = imp.ImportJSON(session.DB, session.Driver, file, opts)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的导入格式: " + format})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("导入失败: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ImportHandler) PreviewImport(c *gin.Context) {
	format := c.PostForm("format")
	delimiter := c.PostForm("delimiter")
	hasHeader := c.PostForm("hasHeader") == "true"

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取上传文件失败"})
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(file, &buf)

	type PreviewResult struct {
		Headers []string     `json:"headers"`
		Rows    [][]string   `json:"rows"`
		Total   int          `json:"total"`
	}

	switch format {
	case "csv":
		headers, rows, err := imp.ParseCSV(tee, delimiter, hasHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		previewRows := rows
		if len(rows) > 10 {
			previewRows = rows[:10]
		}
		c.JSON(http.StatusOK, PreviewResult{Headers: headers, Rows: previewRows, Total: len(rows)})

	case "json":
		allData, err := io.ReadAll(tee)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
			return
		}
		var data []map[string]any
		if err := json.Unmarshal(allData, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON格式无效"})
			return
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

		previewRows := make([][]string, 0)
		limit := len(data)
		if limit > 10 {
			limit = 10
		}
		for i := 0; i < limit; i++ {
			row := make([]string, len(headers))
			for j, h := range headers {
				if val, ok := data[i][h]; ok {
					row[j] = fmt.Sprintf("%v", val)
				}
			}
			previewRows = append(previewRows, row)
		}

		c.JSON(http.StatusOK, PreviewResult{Headers: headers, Rows: previewRows, Total: len(data)})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的格式: " + format})
	}
}
