package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sqlmg/internal/dbengine"
	"sqlmg/internal/metastore"
)

type ConnectionHandler struct {
	store   *metastore.Store
	manager *dbengine.Manager
}

func NewConnectionHandler(store *metastore.Store, manager *dbengine.Manager) *ConnectionHandler {
	return &ConnectionHandler{store: store, manager: manager}
}

type Connection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Options  string `json:"options"`
}

type ConnectionResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Driver    string `json:"driver"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Database  string `json:"database"`
	Options   string `json:"options"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func (h *ConnectionHandler) List(c *gin.Context) {
	rows, err := h.store.DB().Query(
		"SELECT id, name, driver, host, port, username, database, options, created_at, updated_at FROM connections ORDER BY created_at",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var result []ConnectionResponse
	for rows.Next() {
		var conn ConnectionResponse
		if err := rows.Scan(&conn.ID, &conn.Name, &conn.Driver, &conn.Host, &conn.Port,
			&conn.Username, &conn.Database, &conn.Options, &conn.CreatedAt, &conn.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		result = append(result, conn)
	}

	if result == nil {
		result = []ConnectionResponse{}
	}
	c.JSON(http.StatusOK, result)
}

func (h *ConnectionHandler) Create(c *gin.Context) {
	var req Connection
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	_, err := h.store.DB().Exec(
		"INSERT INTO connections (id, name, driver, host, port, username, password, database, options, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, req.Name, req.Driver, req.Host, req.Port, req.Username, req.Password, req.Database, req.Options, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *ConnectionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var conn ConnectionResponse
	err := h.store.DB().QueryRow(
		"SELECT id, name, driver, host, port, username, database, options, created_at, updated_at FROM connections WHERE id = ?",
		id,
	).Scan(&conn.ID, &conn.Name, &conn.Driver, &conn.Host, &conn.Port, &conn.Username, &conn.Database, &conn.Options, &conn.CreatedAt, &conn.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conn)
}

func (h *ConnectionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req Connection
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().Format(time.RFC3339)
	result, err := h.store.DB().Exec(
		"UPDATE connections SET name=?, driver=?, host=?, port=?, username=?, password=?, database=?, options=?, updated_at=? WHERE id=?",
		req.Name, req.Driver, req.Host, req.Port, req.Username, req.Password, req.Database, req.Options, now, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *ConnectionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	result, err := h.store.DB().Exec("DELETE FROM connections WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}

	h.manager.Disconnect(id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConnectionHandler) Test(c *gin.Context) {
	id := c.Param("id")

	var driverName, host, username, password, database string
	var port int
	err := h.store.DB().QueryRow(
		"SELECT driver, host, port, username, password, database FROM connections WHERE id = ?",
		id,
	).Scan(&driverName, &host, &port, &username, &password, &database)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "connection not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	session, err := h.manager.Connect(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "status": "failed"})
		return
	}

	if err := session.Driver.Ping(c.Request.Context(), session.DB); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "status": "failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "driver": driverName})
}

func (h *ConnectionHandler) Connect(c *gin.Context) {
	id := c.Param("id")

	session, err := h.manager.Connect(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "connected",
		"driver": session.DriverName,
	})
}
