package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var host = getEnvOrDefault("POSTGRES_HOST", "localhost")
var port = getEnvOrDefault("POSTGRES_PORT", "postgres")
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
}

func getEnvOrDefault(envName string, def string) string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return def
	}
	return envValue
}
