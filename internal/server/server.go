package server

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"sqlmg/internal/config"
	"sqlmg/internal/dbengine"
	"sqlmg/internal/handler"
	"sqlmg/internal/metastore"
	"sqlmg/internal/middleware"
)

type Server struct {
	engine  *gin.Engine
	config  *config.Config
	store   *metastore.Store
	manager *dbengine.Manager
}

func New(cfg *config.Config, store *metastore.Store, manager *dbengine.Manager) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())

	srv := &Server{
		engine:  engine,
		config:  cfg,
		store:   store,
		manager: manager,
	}

	srv.registerAPIRoutes()
	srv.serveFrontend()

	return srv, nil
}

func (s *Server) registerAPIRoutes() {
	api := s.engine.Group("/api")
	{
		connHandler := handler.NewConnectionHandler(s.store, s.manager)
		api.GET("/connections", connHandler.List)
		api.POST("/connections", connHandler.Create)
		api.GET("/connections/:id", connHandler.Get)
		api.PUT("/connections/:id", connHandler.Update)
		api.DELETE("/connections/:id", connHandler.Delete)
		api.POST("/connections/:id/test", connHandler.Test)
		api.POST("/connections/:id/connect", connHandler.Connect)

		metaHandler := handler.NewMetadataHandler(s.manager)
		api.GET("/connections/:id/databases", metaHandler.GetDatabases)
		api.GET("/connections/:id/schemas", metaHandler.GetSchemas)
		api.GET("/connections/:id/schemas/:schema/tables", metaHandler.GetTables)
		api.GET("/connections/:id/schemas/:schema/tables/:table/columns", metaHandler.GetColumns)
		api.GET("/connections/:id/schemas/:schema/tables/:table/indexes", metaHandler.GetIndexes)
		api.GET("/connections/:id/schemas/:schema/tables/:table/fks", metaHandler.GetForeignKeys)
		api.GET("/connections/:id/schemas/:schema/tables/:table/ddl", metaHandler.GetDDL)

		queryHandler := handler.NewQueryHandler(s.manager, s.store)
		api.POST("/connections/:id/query", queryHandler.Execute)
		api.POST("/connections/:id/explain", queryHandler.Explain)

		historyHandler := handler.NewHistoryHandler(s.store)
		api.GET("/connections/:id/history", historyHandler.List)
		api.DELETE("/connections/:id/history", historyHandler.Clear)
		api.POST("/connections/:id/favorites", historyHandler.SaveFavorite)

		ddlHandler := handler.NewDDLHandler(s.manager, s.store)
		api.POST("/connections/:id/ddl/create/preview", ddlHandler.PreviewCreate)
		api.POST("/connections/:id/ddl/create", ddlHandler.ExecuteCreate)
		api.POST("/connections/:id/ddl/alter/preview", ddlHandler.PreviewAlter)
		api.POST("/connections/:id/ddl/alter", ddlHandler.ExecuteAlter)
		api.DELETE("/connections/:id/schemas/:schema/tables/:table", ddlHandler.DropTable)
		api.POST("/connections/:id/schemas/:schema/tables/:table/truncate", ddlHandler.TruncateTable)
		api.POST("/connections/:id/schemas/:schema/tables/:table/rename", ddlHandler.RenameTable)

		exportHandler := handler.NewExportHandler(s.manager, s.store)
		api.GET("/connections/:id/schemas/:schema/tables/:table/export", exportHandler.ExportTable)
		api.POST("/connections/:id/export", exportHandler.ExportQuery)

		importHandler := handler.NewImportHandler(s.manager, s.store)
		api.POST("/connections/:id/import", importHandler.ImportFile)
		api.POST("/connections/:id/import/preview", importHandler.PreviewImport)

		erHandler := handler.NewERHandler(s.manager, s.store)
		api.GET("/connections/:id/schemas/:schema/er", erHandler.GetERGraph)
		api.GET("/connections/:id/schemas/:schema/tables/:table/relations", erHandler.GetTableRelations)
	}

	api.GET("/drivers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"drivers": dbengine.Drivers()})
	})
}

func (s *Server) serveFrontend() {
	distPath := filepath.Join("web", "dist")
	if info, err := os.Stat(distPath); err == nil && info.IsDir() {
		staticServer := http.FileServer(http.Dir(distPath))
		s.engine.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if path == "/" || path == "" {
				c.Request.URL.Path = "/"
				staticServer.ServeHTTP(c.Writer, c.Request)
				return
			}
			filePath := filepath.Join(distPath, path)
			if _, err := os.Stat(filePath); err == nil {
				staticServer.ServeHTTP(c.Writer, c.Request)
			} else {
				c.Request.URL.Path = "/"
				staticServer.ServeHTTP(c.Writer, c.Request)
			}
		})
	} else {
		s.engine.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "frontend not built - run 'npm run build' in web/"})
		})
	}
}

func (s *Server) RegisterFrontend(embedFS embed.FS, subDir string) {
	sub, err := fs.Sub(embedFS, subDir)
	if err != nil {
		return
	}
	staticServer := http.FileServer(http.FS(sub))
	s.engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		f, err := sub.Open(path[1:])
		if err == nil {
			f.Close()
			staticServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.Request.URL.Path = "/"
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
}

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

func (s *Server) Shutdown() error {
	return nil
}
