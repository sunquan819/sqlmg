package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sqlmg/internal/dbengine"
)

type MetadataHandler struct {
	manager *dbengine.Manager
}

func NewMetadataHandler(manager *dbengine.Manager) *MetadataHandler {
	return &MetadataHandler{manager: manager}
}

func (h *MetadataHandler) getSession(c *gin.Context) (*dbengine.Session, bool) {
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

func (h *MetadataHandler) GetDatabases(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	dbs, err := session.Driver.GetDatabases(c.Request.Context(), session.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dbs)
}

func (h *MetadataHandler) GetSchemas(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	database := c.Query("database")
	schemas, err := session.Driver.GetSchemas(c.Request.Context(), session.DB, database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schemas)
}

func (h *MetadataHandler) GetTables(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	schema := c.Param("schema")
	tables, err := session.Driver.GetTables(c.Request.Context(), session.DB, schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tables)
}

func (h *MetadataHandler) GetColumns(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	schema := c.Param("schema")
	table := c.Param("table")
	columns, err := session.Driver.GetColumns(c.Request.Context(), session.DB, schema, table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, columns)
}

func (h *MetadataHandler) GetIndexes(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	schema := c.Param("schema")
	table := c.Param("table")
	indexes, err := session.Driver.GetIndexes(c.Request.Context(), session.DB, schema, table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, indexes)
}

func (h *MetadataHandler) GetForeignKeys(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	schema := c.Param("schema")
	table := c.Param("table")
	fks, err := session.Driver.GetForeignKeys(c.Request.Context(), session.DB, schema, table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fks)
}

func (h *MetadataHandler) GetDDL(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	schema := c.Param("schema")
	table := c.Param("table")
	ddl, err := session.Driver.GetDDL(c.Request.Context(), session.DB, schema, table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ddl": ddl})
}
