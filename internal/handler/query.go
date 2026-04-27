package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"sqlmg/internal/dbengine"
	"sqlmg/internal/metastore"
)

type QueryHandler struct {
	manager *dbengine.Manager
	store   *metastore.Store
}

func NewQueryHandler(manager *dbengine.Manager, store *metastore.Store) *QueryHandler {
	return &QueryHandler{manager: manager, store: store}
}

type QueryRequest struct {
	SQL      string `json:"sql"`
	Database string `json:"database"`
	Limit    int    `json:"limit"`
}

func (h *QueryHandler) Execute(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SQL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sql is required"})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 1000
	}

	connID := c.Param("id")
	session, ok := h.manager.GetSession(connID)
	if !ok {
		var err error
		session, err = h.manager.Connect(c.Request.Context(), connID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
	}

	start := time.Now()
	result, err := session.Driver.Query(c.Request.Context(), session.DB, req.SQL)
	duration := time.Since(start)

	status := "success"
	var errMsg string
	if err != nil {
		status = "error"
		errMsg = err.Error()
		h.store.DB().Exec(
			"INSERT INTO query_history (connection_id, database, sql, duration_ms, status, error_message) VALUES (?, ?, ?, ?, ?, ?)",
			connID, req.Database, req.SQL, duration.Milliseconds(), status, errMsg,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg, "durationMs": duration.Milliseconds()})
		return
	}

	h.store.DB().Exec(
		"INSERT INTO query_history (connection_id, database, sql, duration_ms, row_count, status) VALUES (?, ?, ?, ?, ?, ?)",
		connID, req.Database, req.SQL, duration.Milliseconds(), result.Total, status,
	)

	c.JSON(http.StatusOK, gin.H{
		"columns":    result.Columns,
		"rows":       result.Rows,
		"total":      result.Total,
		"durationMs": duration.Milliseconds(),
	})
}

func (h *QueryHandler) Explain(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SQL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sql is required"})
		return
	}

	connID := c.Param("id")
	session, ok := h.manager.GetSession(connID)
	if !ok {
		var err error
		session, err = h.manager.Connect(c.Request.Context(), connID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
	}

	explainSQL := "EXPLAIN " + req.SQL
	result, err := session.Driver.Query(c.Request.Context(), session.DB, explainSQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
