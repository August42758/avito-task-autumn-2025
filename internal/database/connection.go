package database

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func ConnectDB(addr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func CloseDD(db *sql.DB) error {
	if err := db.Close(); err != nil {
		return err
	}
	return nil
}

func GetDbAddres(host, port, name, user, password string) string {
	addr := "postgres://"
	addr += user + ":"
	addr += password + "@"
	addr += host + ":"
	addr += port + "/"
	addr += name + "?sslmode=disable"

	return addr
}

func RunMigrations(addr string) error {
	m, err := migrate.New(
		"file://migrations",
		addr,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
