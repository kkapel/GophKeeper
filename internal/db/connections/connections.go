// Package connections устанавливает подключение к PostgreSQL
// и применяет миграции.
package connections

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB — обёртка над пулом подключений к базе данных.
type DB struct {
	db *sql.DB
}

// InitDB открывает подключение к базе, проверяет его и применяет миграции.
func InitDB(dbConnect string, migrationsPath string) (*DB, error) {
	if dbConnect == "" {
		return nil, errors.New("database DSN is empty")
	}
	db, err := sql.Open("pgx", dbConnect)
	if err != nil {
		return nil, err
	}

	databaseInstance := &DB{db: db}

	if err := databaseInstance.CheckConnect(); err != nil {
		return nil, err
	}

	if err = migrateDB(dbConnect, migrationsPath); err != nil {
		return nil, err
	}

	return databaseInstance, nil

}

// CheckConnect проверяет доступность базы данных.
func (db *DB) CheckConnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

// Close закрывает подключение к базе данных.
func (db *DB) Close() error {
	err := db.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func migrateDB(dbConnect string, migrationsPath string) error {
	m, err := migrate.New("file://"+migrationsPath, dbConnect)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		if err.Error() == "Dirty database version 1. Fix and force version." ||
			strings.Contains(err.Error(), "dirty") {
			_ = m.Force(1) // Снимаем флаг dirty
			return m.Up()  // Пробуем накатить снова
		}
		return err
	}
	return nil
}

// GetSqlDb возвращает базовый пул *sql.DB.
func (db *DB) GetSqlDb() *sql.DB {
	return db.db
}
