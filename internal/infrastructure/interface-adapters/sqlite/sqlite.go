package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"go-redis/config"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteConnectionPool struct {
	connectionPool *sql.DB
	config         config.IConfig
	mu             sync.RWMutex
}

type ISQLiteDb interface {
	GetConnectionPool(ctx context.Context) *sql.DB
	Close() error
}

func NewSQLitePool(config config.IConfig) ISQLiteDb {
	sqliteDb := &sqliteConnectionPool{config: config}
	sqliteDb.GetConnectionPool(context.Background())
	return sqliteDb
}

func (c *sqliteConnectionPool) GetConnectionPool(ctx context.Context) *sql.DB {
	c.mu.RLock()
	if c.connectionPool != nil {
		defer c.mu.RUnlock()
		return c.connectionPool
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.connectionPool != nil {
		return c.connectionPool
	}

	log.Println("Getting SQLite Connection pool")

	dbPath := getSQLiteDBPath(c.config)
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=on")
	if err != nil {
		log.Panic("Cannot open SQLite database", err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(getConnectionPoolSize())
	db.SetMaxIdleConns(getIdleConnectionPoolSize())

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		log.Panic("Failed to ping SQLite database", err)
	}

	// Initialize schema
	if err := c.initializeSchema(db); err != nil {
		log.Panic("Failed to initialize schema", err)
	}

	c.connectionPool = db
	log.Println("SQLite connection pool initialized successfully")
	return c.connectionPool
}

func (c *sqliteConnectionPool) initializeSchema(db *sql.DB) error {
	schemaPath := filepath.Join("internal", "infrastructure", "interface-adapters", "sqlite", "schema.sql")
	schemaSQL, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err := db.Exec(string(schemaSQL)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("SQLite schema initialized successfully")
	return nil
}

func (c *sqliteConnectionPool) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connectionPool != nil {
		return c.connectionPool.Close()
	}
	return nil
}

func getConnectionPoolSize() int {
	connectionSize := os.Getenv("SQLITE_MAX_CONNECTION")
	if connectionSize == "" {
		return 25 // SQLite default for concurrent connections
	}
	connSize, err := strconv.Atoi(connectionSize)
	if err != nil {
		log.Fatal("Error converting SQLite connection size", err)
	}
	return connSize
}

func getIdleConnectionPoolSize() int {
	idleSize := os.Getenv("SQLITE_MAX_IDLE_CONNECTION")
	if idleSize == "" {
		return 10
	}
	connSize, err := strconv.Atoi(idleSize)
	if err != nil {
		log.Fatal("Error converting SQLite idle connection size", err)
	}
	return connSize
}

func getSQLiteDBPath(config config.IConfig) string {
	dbPath := config.GetConfig().SQLiteDBPath
	if dbPath == "" {
		log.Panic("No SQLite database path provided")
	}
	return dbPath
}

type SQLiteConnection struct {
	pool ISQLiteDb
}

func NewSQLiteConnection(config config.IConfig) (*SQLiteConnection, error) {
	// Create a minimal config for backward compatibility
	pool := NewSQLitePool(config.GetConfig())

	return &SQLiteConnection{pool: pool}, nil
}

func (c *SQLiteConnection) GetDB() *sql.DB {
	return c.pool.GetConnectionPool(context.Background())
}

func (c *SQLiteConnection) Close() error {
	return c.pool.Close()
}
