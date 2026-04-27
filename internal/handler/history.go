package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sqlmg/internal/metastore"
)

type HistoryHandler struct {
	store *metastore.Store
}

func NewHistoryHandler(store *metastore.Store) *HistoryHandler {
	return &HistoryHandler{store: store}
}

func (h *HistoryHandler) List(c *gin.Context) {
	connID := c.Param("id")
	limit := c.DefaultQuery("limit", "100")

	rows, err := h.store.DB().Query(
		"SELECT id, connection_id, database, sql, duration_ms, row_count, status, error_message, created_at FROM query_history WHERE connection_id = ? ORDER BY created_at DESC LIMIT ?",
		connID, limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type HistoryEntry struct {
		ID           int64  `json:"id"`
		ConnectionID string `json:"connectionId"`
		Database     string `json:"database"`
		SQL          string `json:"sql"`
		DurationMs   int64  `json:"duration_ms"`
		RowCount     int64  `json:"row_count"`
		Status       string `json:"status"`
		ErrorMessage string `json:"error_message"`
		CreatedAt    string `json:"created_at"`
	}

	var result []HistoryEntry
	for rows.Next() {
		var entry HistoryEntry
		if err := rows.Scan(&entry.ID, &entry.ConnectionID, &entry.Database, &entry.SQL,
			&entry.DurationMs, &entry.RowCount, &entry.Status, &entry.ErrorMessage, &entry.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result = append(result, entry)
	}

	if result == nil {
		result = []HistoryEntry{}
	}
	c.JSON(http.StatusOK, result)
}

func (h *HistoryHandler) Clear(c *gin.Context) {
	connID := c.Param("id")
	_, err := h.store.DB().Exec("DELETE FROM query_history WHERE connection_id = ?", connID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cleared"})
}

func (h *HistoryHandler) SaveFavorite(c *gin.Context) {
	var req struct {
		Title      string `json:"title"`
		SQL        string `json:"sql"`
		Database   string `json:"database"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	connID := c.Param("id")
	_, err := h.store.DB().Exec(
		"INSERT INTO query_favorites (connection_id, database, title, sql) VALUES (?, ?, ?, ?)",
		connID, req.Database, req.Title, req.SQL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "saved"})
}
