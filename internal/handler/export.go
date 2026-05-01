package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"sqlmg/internal/dbengine"
	"sqlmg/internal/export"
	"sqlmg/internal/metastore"
)

type ExportHandler struct {
	manager *dbengine.Manager
	store   *metastore.Store
}

func NewExportHandler(manager *dbengine.Manager, store *metastore.Store) *ExportHandler {
	return &ExportHandler{manager: manager, store: store}
}

func (h *ExportHandler) getSession(c *gin.Context) (*dbengine.Session, bool) {
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

func (h *ExportHandler) ExportTable(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	schema := c.Param("schema")
	table := c.Param("table")
	format := c.DefaultQuery("format", "csv")
	delimiter := c.DefaultQuery("delimiter", ",")
	includeHeader := c.DefaultQuery("header", "true") == "true"
	batchSize, _ := strconv.Atoi(c.DefaultQuery("batch", "100"))

	opts := export.ExportOptions{
		Format:        export.ExportFormat(format),
		Schema:        schema,
		Table:         table,
		Delimiter:     delimiter,
		IncludeHeader: includeHeader,
		InsertBatch:   batchSize,
	}

	query := fmt.Sprintf("SELECT * FROM `%s`.`%s`", schema, table)
	if session.DriverName == "sqlite" || session.DriverName == "postgres" {
		query = fmt.Sprintf("SELECT * FROM \"%s\"", table)
	}

	filename := export.SanitizeFilename(fmt.Sprintf("%s_%s", table, format))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", filename, format))

	switch format {
	case "csv":
		c.Header("Content-Type", "text/csv; charset=utf-8")
	case "json":
		c.Header("Content-Type", "application/json; charset=utf-8")
	case "sql":
		c.Header("Content-Type", "text/plain; charset=utf-8")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的导出格式: " + format})
		return
	}

	export.ExportQueryResult(c.Request.Context(), session.DB, session.Driver, query, c.Writer, opts)
}

func (h *ExportHandler) ExportQuery(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	var req struct {
		SQL          string `json:"sql"`
		Format       string `json:"format"`
		Schema       string `json:"schema"`
		Table        string `json:"table"`
		Delimiter    string `json:"delimiter"`
		IncludeHeader bool   `json:"includeHeader"`
		InsertBatch  int    `json:"insertBatch"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SQL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SQL不能为空"})
		return
	}

	if req.Format == "" {
		req.Format = "csv"
	}
	if req.Schema == "" {
		req.Schema = "main"
	}
	if req.Table == "" {
		req.Table = "query_result"
	}

	opts := export.ExportOptions{
		Format:        export.ExportFormat(req.Format),
		Schema:        req.Schema,
		Table:         req.Table,
		Delimiter:     req.Delimiter,
		IncludeHeader: req.IncludeHeader,
		InsertBatch:   req.InsertBatch,
	}

	filename := export.SanitizeFilename(fmt.Sprintf("%s_%s", req.Table, req.Format))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", filename, req.Format))

	switch req.Format {
	case "csv":
		c.Header("Content-Type", "text/csv; charset=utf-8")
	case "json":
		c.Header("Content-Type", "application/json; charset=utf-8")
	case "sql":
		c.Header("Content-Type", "text/plain; charset=utf-8")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的导出格式: " + req.Format})
		return
	}

	export.ExportQueryResult(c.Request.Context(), session.DB, session.Driver, req.SQL, c.Writer, opts)
}
