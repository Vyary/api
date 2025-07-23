package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Service interface{}

type service struct {
	db *sql.DB
}

func New() Service {
	dbName := os.Getenv("DB_NAME")
	token := os.Getenv("TOKEN")

	if dbName == "" {
		slog.Error("DB_NAME env variable is required")
		os.Exit(1)
	}

	if token == "" {
		slog.Error("TOKEN env variable is required")
		os.Exit(1)
	}

	url := fmt.Sprintf("libsql://%s.turso.io?authToken=%s", dbName, token)

	db, err := sql.Open("libsql", url)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return &service{db: db}
}
