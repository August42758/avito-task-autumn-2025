package testutils

import (
	"database/sql"
	"os"
	"testing"

	"pr-service/internal/config"
	"pr-service/internal/database"

	_ "github.com/lib/pq"
)

func NewTestDB(t *testing.T) *sql.DB {
	// загрузка .env
	cfg, err := config.Load("../.env")
	if err != nil {
		t.Fatal(err)
		return nil
	}

	dbAddr := database.GetDbAddres(cfg.TestDBHost, cfg.TestDBPort, cfg.TestDBName, cfg.TestDBUser, cfg.TestDBPassword)

	// подключаемся к базе
	db, err := database.ConnectDB(dbAddr)
	if err != nil {
		t.Fatal(err)
	}

	// создаем таблицы
	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	return db
}

func DeleteDb(t *testing.T, db *sql.DB) {
	defer db.Close()

	script, err := os.ReadFile("./testdata/teardown.sql")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}
}

// Вспомогательная функция для создания пользоватлея
func RunQuery(t *testing.T, db *sql.DB, path string) {
	script, err := os.ReadFile(path)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
}
