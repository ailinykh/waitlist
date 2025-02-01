package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func WithURL(url string) func(*Connection) {
	return func(c *Connection) {
		c.url = url
	}
}

func WithMigrations(migrations fs.FS) func(*Connection) {
	return func(c *Connection) {
		c.migrations = migrations
	}
}

func New(logger *slog.Logger, opts ...func(*Connection)) (*sql.DB, error) {
	conn := &Connection{
		url:        "mysql://root:password@/waitlist?parseTime=true",
		logger:     logger,
		migrations: nil,
	}

	for _, opt := range opts {
		opt(conn)
	}

	return conn.build()
}

type Connection struct {
	url        string
	logger     *slog.Logger
	migrations fs.FS
}

func (conn *Connection) build() (*sql.DB, error) {
	url := strings.TrimPrefix(conn.url, "mysql://")
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}
	// TODO: move to opts
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if conn.migrations != nil {
		conn.logger.Debug("running migrations")

		source, err := iofs.New(conn.migrations, "migrations")
		if err != nil {
			return nil, fmt.Errorf("failed to create source: %w", err)
		}

		migrator, err := migrate.NewWithSourceInstance("iofs", source, conn.url)
		if err != nil {
			return nil, fmt.Errorf("failed to create migration source: %s", err)
		}

		if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return db, nil
}
