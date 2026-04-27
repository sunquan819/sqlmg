package app

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"sqlmg/internal/config"
	"sqlmg/internal/dbengine"
	"sqlmg/internal/metastore"
	"sqlmg/internal/server"
)

type App struct {
	config     *config.Config
	metaStore  *metastore.Store
	manager    *dbengine.Manager
	httpServer *server.Server
}

func New(cfg *config.Config) (*App, error) {
	store, err := metastore.New(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("init metastore: %w", err)
	}

	manager := dbengine.NewManager(store)

	srv, err := server.New(cfg, store, manager)
	if err != nil {
		return nil, fmt.Errorf("init server: %w", err)
	}

	return &App{
		config:     cfg,
		metaStore:  store,
		manager:    manager,
		httpServer: srv,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)

	go func() {
		if err := a.httpServer.Run(addr); err != nil {
			fmt.Printf("server error: %v\n", err)
		}
	}()

	url := fmt.Sprintf("http://%s:%d", a.config.Server.Host, a.config.Server.Port)
	fmt.Printf("SQLMG is running at %s\n", url)

	openBrowser(url)

	return nil
}

func (a *App) Shutdown() error {
	a.manager.CloseAll()
	if err := a.httpServer.Shutdown(); err != nil {
		return err
	}
	if err := a.metaStore.Close(); err != nil {
		return err
	}
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
