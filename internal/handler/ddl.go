package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sqlmg/internal/dbengine"
	"sqlmg/internal/metastore"
)

type DDLHandler struct {
	manager *dbengine.Manager
	store   *metastore.Store
}

func NewDDLHandler(manager *dbengine.Manager, store *metastore.Store) *DDLHandler {
	return &DDLHandler{manager: manager, store: store}
}

func (h *DDLHandler) getDriverName(c *gin.Context) string {
	connID := c.Param("id")
	if session, ok := h.manager.GetSession(connID); ok {
		return session.DriverName
	}
	var driver string
	_ = h.store.DB().QueryRow("SELECT driver FROM connections WHERE id = ?", connID).Scan(&driver)
	if driver != "" {
		return driver
	}
	return "mysql"
}

func (h *DDLHandler) getSession(c *gin.Context) (*dbengine.Session, bool) {
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

type CreateTableRequest struct {
	Database string         `json:"database"`
	Schema   string         `json:"schema"`
	Table    string         `json:"table"`
	Columns  []ColumnDef    `json:"columns"`
	Indexes  []IndexDef     `json:"indexes,omitempty"`
	ForeignKeys []FKDef     `json:"foreignKeys,omitempty"`
	Comment  string         `json:"comment,omitempty"`
}

type ColumnDef struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Length        int    `json:"length,omitempty"`
	Nullable      bool   `json:"nullable"`
	DefaultValue  string `json:"defaultValue,omitempty"`
	IsPrimary     bool   `json:"isPrimary"`
	AutoIncrement bool   `json:"autoIncrement"`
	Comment       string `json:"comment,omitempty"`
}

type IndexDef struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type,omitempty"`
}

type FKDef struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	RefTable   string   `json:"refTable"`
	RefColumns []string `json:"refColumns"`
	OnDelete   string   `json:"onDelete,omitempty"`
	OnUpdate   string   `json:"onUpdate,omitempty"`
}

type AlterTableRequest struct {
	Database   string      `json:"database"`
	Schema     string      `json:"schema"`
	Table      string      `json:"table"`
	Operations []AlterOp   `json:"operations"`
}

type AlterOp struct {
	Type     string     `json:"type"`
	Column   *ColumnDef `json:"column,omitempty"`
	OldName  string     `json:"oldName,omitempty"`
	Index    *IndexDef  `json:"index,omitempty"`
	FK       *FKDef     `json:"fk,omitempty"`
}

func (h *DDLHandler) PreviewCreate(c *gin.Context) {
	var req CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sql := buildCreateTableSQL(req, h.getDriverName(c))
	c.JSON(http.StatusOK, gin.H{"sql": sql})
}

func (h *DDLHandler) ExecuteCreate(c *gin.Context) {
	var req CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sql := buildCreateTableSQL(req, h.getDriverName(c))

	session, ok := h.getSession(c)
	if !ok {
		return
	}

	result, err := session.Driver.Exec(c.Request.Context(), session.DB, sql)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "sql": sql})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "table created",
		"sql": sql,
		"rowsAffected": result.RowsAffected,
	})
}

func (h *DDLHandler) PreviewAlter(c *gin.Context) {
	var req AlterTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statements := buildAlterTableSQL(req, h.getDriverName(c))
	c.JSON(http.StatusOK, gin.H{"sql": strings.Join(statements, ";\n")})
}

func (h *DDLHandler) ExecuteAlter(c *gin.Context) {
	var req AlterTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statements := buildAlterTableSQL(req, h.getDriverName(c))

	session, ok := h.getSession(c)
	if !ok {
		return
	}

	var results []map[string]any
	for _, stmt := range statements {
		if stmt == "" {
			continue
		}
		result, err := session.Driver.Exec(c.Request.Context(), session.DB, stmt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
				"failedSql": stmt,
				"executed": results,
			})
			return
		}
		results = append(results, map[string]any{
			"sql": stmt,
			"rowsAffected": result.RowsAffected,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "table altered", "results": results})
}

func (h *DDLHandler) DropTable(c *gin.Context) {
	_ = c.Param("id")
	schema := c.Param("schema")
	table := c.Param("table")

	var req struct {
		Cascade bool `json:"cascade"`
	}
	c.ShouldBindJSON(&req)

	session, ok := h.getSession(c)
	if !ok {
		return
	}

	sql := fmt.Sprintf("DROP TABLE `%s`.`%s`", schema, table)
	if req.Cascade {
		sql += " CASCADE"
	}

	_, err := session.Driver.Exec(c.Request.Context(), session.DB, sql)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "sql": sql})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "table dropped", "sql": sql})
}

func (h *DDLHandler) TruncateTable(c *gin.Context) {
	schema := c.Param("schema")
	table := c.Param("table")

	session, ok := h.getSession(c)
	if !ok {
		return
	}

	sql := fmt.Sprintf("TRUNCATE TABLE `%s`.`%s`", schema, table)

	_, err := session.Driver.Exec(c.Request.Context(), session.DB, sql)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "sql": sql})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "table truncated", "sql": sql})
}

func (h *DDLHandler) RenameTable(c *gin.Context) {
	var req struct {
		NewName string `json:"newName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schema := c.Param("schema")
	table := c.Param("table")

	session, ok := h.getSession(c)
	if !ok {
		return
	}

	var sql string
	switch session.DriverName {
	case "postgres":
		sql = fmt.Sprintf("ALTER TABLE \"%s\".\"%s\" RENAME TO \"%s\"", schema, table, req.NewName)
	default:
		sql = fmt.Sprintf("RENAME TABLE `%s`.`%s` TO `%s`.`%s`", schema, table, schema, req.NewName)
	}

	_, err := session.Driver.Exec(c.Request.Context(), session.DB, sql)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "sql": sql})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "table renamed", "sql": sql})
}

func buildCreateTableSQL(req CreateTableRequest, driverName string) string {
	var sb strings.Builder

	isSQLite := driverName == "sqlite"
	isPG := driverName == "postgres"

	schema := req.Schema
	if schema == "" {
		schema = req.Database
	}

	if isSQLite || isPG {
		sb.WriteString(fmt.Sprintf("CREATE TABLE \"%s\" (\n", req.Table))
	} else {
		sb.WriteString(fmt.Sprintf("CREATE TABLE `%s`.`%s` (\n", schema, req.Table))
	}

	var primaryKeys []string
	var columnDefs []string

	for i, col := range req.Columns {
		def := fmt.Sprintf("  `%s` %s", col.Name, col.Type)

		if col.Length > 0 && !strings.Contains(strings.ToUpper(col.Type), "INT") && !strings.Contains(strings.ToUpper(col.Type), "TEXT") && !strings.Contains(strings.ToUpper(col.Type), "DATE") && !strings.Contains(strings.ToUpper(col.Type), "BOOL") {
			def += fmt.Sprintf("(%d)", col.Length)
		}

		if !col.Nullable {
			def += " NOT NULL"
		}

		if col.AutoIncrement {
			if isSQLite {
				def += " PRIMARY KEY AUTOINCREMENT"
			} else {
				def += " AUTO_INCREMENT"
			}
		}

		if col.DefaultValue != "" {
			if strings.ToUpper(col.DefaultValue) == "NULL" {
				def += " DEFAULT NULL"
			} else if strings.ToUpper(col.DefaultValue) == "CURRENT_TIMESTAMP" {
				def += " DEFAULT CURRENT_TIMESTAMP"
			} else {
				def += fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue)
			}
		}

		if col.Comment != "" {
			def += fmt.Sprintf(" COMMENT '%s'", col.Comment)
		}

		if col.IsPrimary {
			if !(isSQLite && col.AutoIncrement) {
				primaryKeys = append(primaryKeys, col.Name)
			}
		}

		columnDefs = append(columnDefs, def)
		if i < len(req.Columns)-1 || len(primaryKeys) > 0 || len(req.Indexes) > 0 || len(req.ForeignKeys) > 0 {
			def += ","
		}
		columnDefs[len(columnDefs)-1] = def
	}

	if len(primaryKeys) > 0 {
		pkCols := make([]string, len(primaryKeys))
		for i, pk := range primaryKeys {
			pkCols[i] = fmt.Sprintf("`%s`", pk)
		}
		pkDef := fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(pkCols, ", "))
		if len(req.Indexes) > 0 || len(req.ForeignKeys) > 0 {
			pkDef += ","
		}
		columnDefs = append(columnDefs, pkDef)
	}

	for i, idx := range req.Indexes {
		idxCols := make([]string, len(idx.Columns))
		for j, col := range idx.Columns {
			idxCols[j] = fmt.Sprintf("`%s`", col)
		}
		idxDef := fmt.Sprintf("  ")
		if idx.Unique {
			idxDef += "UNIQUE "
		}
		idxDef += fmt.Sprintf("INDEX `%s` (%s)", idx.Name, strings.Join(idxCols, ", "))
		if i < len(req.Indexes)-1 || len(req.ForeignKeys) > 0 {
			idxDef += ","
		}
		columnDefs = append(columnDefs, idxDef)
	}

	for i, fk := range req.ForeignKeys {
		fkCols := make([]string, len(fk.Columns))
		for j, col := range fk.Columns {
			fkCols[j] = fmt.Sprintf("`%s`", col)
		}
		refCols := make([]string, len(fk.RefColumns))
		for j, col := range fk.RefColumns {
			refCols[j] = fmt.Sprintf("`%s`", col)
		}
		fkDef := fmt.Sprintf("  CONSTRAINT `%s` FOREIGN KEY (%s) REFERENCES `%s` (%s)",
			fk.Name, strings.Join(fkCols, ", "), fk.RefTable, strings.Join(refCols, ", "))
		if fk.OnDelete != "" {
			fkDef += fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
		}
		if fk.OnUpdate != "" {
			fkDef += fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
		}
		if i < len(req.ForeignKeys)-1 {
			fkDef += ","
		}
		columnDefs = append(columnDefs, fkDef)
	}

	sb.WriteString(strings.Join(columnDefs, "\n"))
	sb.WriteString("\n)")

	if req.Comment != "" {
		sb.WriteString(fmt.Sprintf(" COMMENT='%s'", req.Comment))
	}

	return sb.String()
}

func buildAlterTableSQL(req AlterTableRequest, driverName string) []string {
	var statements []string
	schema := req.Schema
	if schema == "" {
		schema = req.Database
	}

	var tableName string
	if driverName == "sqlite" || driverName == "postgres" {
		tableName = fmt.Sprintf("\"%s\"", req.Table)
	} else {
		tableName = fmt.Sprintf("`%s`.`%s`", schema, req.Table)
	}

	for _, op := range req.Operations {
		switch op.Type {
		case "add_column":
			if op.Column != nil {
				def := fmt.Sprintf("`%s` %s", op.Column.Name, op.Column.Type)
				if !op.Column.Nullable {
					def += " NOT NULL"
				}
				if op.Column.DefaultValue != "" {
					def += fmt.Sprintf(" DEFAULT '%s'", op.Column.DefaultValue)
				}
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, def))
			}

		case "drop_column":
			if op.OldName != "" {
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s DROP COLUMN `%s`", tableName, op.OldName))
			}

		case "rename_column":
			if op.Column != nil && op.OldName != "" {
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN `%s` TO `%s`", tableName, op.OldName, op.Column.Name))
			}

		case "modify_column":
			if op.Column != nil {
				def := fmt.Sprintf("`%s` %s", op.Column.Name, op.Column.Type)
				if !op.Column.Nullable {
					def += " NOT NULL"
				}
				if op.Column.DefaultValue != "" {
					def += fmt.Sprintf(" DEFAULT '%s'", op.Column.DefaultValue)
				}
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", tableName, def))
			}

		case "add_index":
			if op.Index != nil {
				idxCols := make([]string, len(op.Index.Columns))
				for i, col := range op.Index.Columns {
					idxCols[i] = fmt.Sprintf("`%s`", col)
				}
				idxType := "INDEX"
				if op.Index.Unique {
					idxType = "UNIQUE INDEX"
				}
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s ADD %s `%s` (%s)", tableName, idxType, op.Index.Name, strings.Join(idxCols, ", ")))
			}

		case "drop_index":
			if op.OldName != "" {
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s DROP INDEX `%s`", tableName, op.OldName))
			}

		case "add_fk":
			if op.FK != nil {
				fkCols := make([]string, len(op.FK.Columns))
				for i, col := range op.FK.Columns {
					fkCols[i] = fmt.Sprintf("`%s`", col)
				}
				refCols := make([]string, len(op.FK.RefColumns))
				for i, col := range op.FK.RefColumns {
					refCols[i] = fmt.Sprintf("`%s`", col)
				}
				sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT `%s` FOREIGN KEY (%s) REFERENCES `%s` (%s)",
					tableName, op.FK.Name, strings.Join(fkCols, ", "), op.FK.RefTable, strings.Join(refCols, ", "))
				if op.FK.OnDelete != "" {
					sql += fmt.Sprintf(" ON DELETE %s", op.FK.OnDelete)
				}
				if op.FK.OnUpdate != "" {
					sql += fmt.Sprintf(" ON UPDATE %s", op.FK.OnUpdate)
				}
				statements = append(statements, sql)
			}

		case "drop_fk":
			if op.OldName != "" {
				statements = append(statements, fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY `%s`", tableName, op.OldName))
			}
		}
	}

	return statements
}
