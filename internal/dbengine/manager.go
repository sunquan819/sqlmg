package dbengine

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"sqlmg/internal/metastore"
)

type Manager struct {
	store    *metastore.Store
	sessions map[string]*Session
	mu       sync.RWMutex
}

type Session struct {
	ID           string
	ConnectionID string
	Driver       IDriver
	DB           *sql.DB
	DriverName   string
}

func NewManager(store *metastore.Store) *Manager {
	return &Manager{
		store:    store,
		sessions: make(map[string]*Session),
	}
}

func (m *Manager) Connect(ctx context.Context, connectionID string) (*Session, error) {
	m.mu.RLock()
	for _, s := range m.sessions {
		if s.ConnectionID == connectionID {
			m.mu.RUnlock()
			return s, nil
		}
	}
	m.mu.RUnlock()

	var name, driver, host, username, password, database, options string
	var port int
	err := m.store.DB().QueryRow(
		"SELECT name, driver, host, port, username, password, database, options FROM connections WHERE id = ?",
		connectionID,
	).Scan(&name, &driver, &host, &port, &username, &password, &database, &options)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	drv, ok := Get(driver)
	if !ok {
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	var dsn string
	switch driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
		if !strings.Contains(options, "parseTime") {
			dsn += "?parseTime=true&charset=utf8mb4&loc=Local"
		}
	case "postgres":
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=prefer", username, password, host, port, database)
	case "sqlite":
		dsn = database
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	db, err := drv.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}

	session := &Session{
		ID:           connectionID,
		ConnectionID: connectionID,
		Driver:       drv,
		DB:           db,
		DriverName:   driver,
	}

	m.mu.Lock()
	m.sessions[connectionID] = session
	m.mu.Unlock()

	return session, nil
}

func (m *Manager) GetSession(connectionID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[connectionID]
	return s, ok
}

func (m *Manager) Disconnect(connectionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[connectionID]
	if !ok {
		return nil
	}

	err := session.Driver.Close(session.DB)
	delete(m.sessions, connectionID)
	return err
}

func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, session := range m.sessions {
		_ = session.Driver.Close(session.DB)
		delete(m.sessions, id)
	}
}
