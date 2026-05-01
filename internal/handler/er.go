package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sqlmg/internal/dbengine"
	"sqlmg/internal/metastore"
)

type ERHandler struct {
	manager *dbengine.Manager
	store   *metastore.Store
}

func NewERHandler(manager *dbengine.Manager, store *metastore.Store) *ERHandler {
	return &ERHandler{manager: manager, store: store}
}

func (h *ERHandler) getSession(c *gin.Context) (*dbengine.Session, bool) {
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

type ERNode struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Schema   string   `json:"schema"`
	Columns  []ERCol  `json:"columns"`
	X        float64  `json:"x,omitempty"`
	Y        float64  `json:"y,omitempty"`
	Width    float64  `json:"width,omitempty"`
	Height   float64  `json:"height,omitempty"`
}

type ERCol struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsPrimary bool   `json:"isPrimary"`
}

type EREdge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	SourceCol string `json:"sourceCol"`
	Target    string `json:"target"`
	TargetCol string `json:"targetCol"`
	Label     string `json:"label,omitempty"`
}

type ERGraph struct {
	Nodes []ERNode `json:"nodes"`
	Edges []EREdge `json:"edges"`
}

func (h *ERHandler) GetERGraph(c *gin.Context) {
	session, ok := h.getSession(c)
	if !ok {
		return
	}

	schema := c.Param("schema")
	if schema == "" {
		schema = "main"
	}

	tables, err := session.Driver.GetTables(c.Request.Context(), session.DB, schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	graph := ERGraph{Nodes: []ERNode{}, Edges: []EREdge{}}

	nodeMap := make(map[string]*ERNode)

	for _, tbl := range tables {
		if tbl.Type != "TABLE" {
			continue
		}

		columns, err := session.Driver.GetColumns(c.Request.Context(), session.DB, schema, tbl.Name)
		if err != nil {
			continue
		}

		node := ERNode{
			ID:      tbl.Name,
			Name:    tbl.Name,
			Type:    "table",
			Schema:  schema,
			Columns: []ERCol{},
		}

		for _, col := range columns {
			node.Columns = append(node.Columns, ERCol{
				Name:     col.Name,
				Type:     col.Type,
				IsPrimary: col.IsPrimary,
			})
		}

		graph.Nodes = append(graph.Nodes, node)
		nodeMap[tbl.Name] = &graph.Nodes[len(graph.Nodes)-1]
	}

	for _, tbl := range tables {
		if tbl.Type != "TABLE" {
			continue
		}

		fks, err := session.Driver.GetForeignKeys(c.Request.Context(), session.DB, schema, tbl.Name)
		if err != nil {
			continue
		}

		for _, fk := range fks {
			for i, srcCol := range fk.Columns {
				targetCol := ""
				if i < len(fk.RefColumns) {
					targetCol = fk.RefColumns[i]
				}

				edge := EREdge{
					ID:        fk.Name + "_" + srcCol,
					Source:    tbl.Name,
					SourceCol: srcCol,
					Target:    fk.RefTable,
					TargetCol: targetCol,
					Label:     fk.Name,
				}
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}

	c.JSON(http.StatusOK, graph)
}

func (h *ERHandler) GetTableRelations(c *gin.Context) {
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

	type Relation struct {
		Table      string   `json:"table"`
		Type       string   `json:"type"` // "outgoing" or "incoming"
		Columns    []string `json:"columns"`
		RefTable   string   `json:"refTable"`
		RefColumns []string `json:"refColumns"`
		OnDelete   string   `json:"onDelete"`
		OnUpdate   string   `json:"onUpdate"`
	}

relations := []Relation{}

	for _, fk := range fks {
		relations = append(relations, Relation{
			Table:      table,
			Type:       "outgoing",
			Columns:    fk.Columns,
			RefTable:   fk.RefTable,
			RefColumns: fk.RefColumns,
			OnDelete:   fk.OnDelete,
			OnUpdate:   fk.OnUpdate,
		})
	}

	allTables, err := session.Driver.GetTables(c.Request.Context(), session.DB, schema)
	if err == nil {
		for _, tbl := range allTables {
			if tbl.Name == table || tbl.Type != "TABLE" {
				continue
			}
			otherFKs, err := session.Driver.GetForeignKeys(c.Request.Context(), session.DB, schema, tbl.Name)
			if err != nil {
				continue
			}
			for _, fk := range otherFKs {
				if fk.RefTable == table {
					relations = append(relations, Relation{
						Table:      tbl.Name,
						Type:       "incoming",
						Columns:    fk.Columns,
						RefTable:   table,
						RefColumns: fk.RefColumns,
						OnDelete:   fk.OnDelete,
						OnUpdate:   fk.OnUpdate,
					})
				}
			}
		}
	}

	c.JSON(http.StatusOK, relations)
}