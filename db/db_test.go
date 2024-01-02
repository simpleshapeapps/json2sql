package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var host = getEnvOrDefault("POSTGRES_HOST", "localhost")
var port = getEnvOrDefault("POSTGRES_PORT", "5432")
var user = getEnvOrDefault("POSTGRES_USER", "postgres")
var password = getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
var databaseName = getEnvOrDefault("POSTGRES_DATABASE", "postgres")

func TestMain(m *testing.M) {
	if host != "" {
		code := m.Run()
		os.Exit(code)
	}

	os.Exit(0)
}

func TestDbConnection(t *testing.T) {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, databaseName)

	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Queryx("SELECT 1")
	if err != nil {
		t.Fatal(err)
	}

	var result int

	rows.Next()
	err = rows.Scan(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result != 1 {
		t.Fatalf("SELECT 1 returned: %d", result)
	}
}

func getEnvOrDefault(envName string, def string) string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return def
	}
	return envValue
}
