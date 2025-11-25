package testutils

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func NewTestDB(t *testing.T) *sql.DB {
	// подключаемся к базе
	db, err := sql.Open("postgres", "postgres://postgres:12345@localhost:5432/pr-service-test?sslmode=disable")
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
