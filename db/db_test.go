package main

import (
	"fmt"
	"json2sql/generators"
	"json2sql/types"
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

var db *sqlx.DB

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

func TestCreateTable(t *testing.T) {
	types.Register(parentThing)
	types.Register(childThing)
	types.Register(otherThing)
	defer types.Clear()

	generator := generators.CreateTable{
		Thing: parentThing,
	}

	sqls, err := generator.GetSql()
	if err != nil {
		t.Fatal(err)
	}

	doAndRollback(func(tx *sqlx.Tx) {
		for _, sql := range sqls {
			tx.MustExec(sql)
		}

		things := []types.ThingConfig{parentThing, childThing, otherThing}
		for _, thing := range things {
			tableName := thing.GetTableName()
			rows, err := tx.Queryx("SELECT * FROM " + tableName)
			if err != nil {
				t.Fatal(err)
			}

			if rows.Next() {
				t.Fatalf("No rows in %s expected", tableName)
			}
		}
	})
}

func getDb() *sqlx.DB {
	if db == nil {
		connectionString := fmt.Sprintf("host=%s port=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, databaseName)

		db = sqlx.MustConnect("postgres", connectionString)
	}

	return db
}

func doAndRollback(callback func(tx *sqlx.Tx)) {
	db := getDb()
	tx := db.MustBegin()
	callback(tx)
	tx.Rollback()
}

func getEnvOrDefault(envName string, def string) string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return def
	}
	return envValue
}

var parentThing = types.ThingConfig{
	Name: "parentThing",
	Fields: map[string]types.FieldConfig{
		"primary_key": {
			Name: "primary_key",
			Type: types.PRIMARY_KEY,
		},
		"string": {
			Name: "string",
			Type: types.STRING,
		},
		"number": {
			Name: "number",
			Type: types.NUMBER,
		},
		"boolean": {
			Name: "boolean",
			Type: types.BOOLEAN,
		},
		"date": {
			Name: "date",
			Type: types.DATE,
		},
		"thing": {
			Name:          "thing",
			Type:          types.THING,
			TypeThingName: "otherThing",
		},
		"oneToMany": {
			Name: "oneToMany",
			Type: types.RELATION,
			Relation: types.ThingRelation{
				Type:           types.ONE_TO_MANY,
				OtherThingName: "childThing",
				OtherFieldName: "manyToOne",
			},
		},
	},
}

var childThing = types.ThingConfig{
	Name: "childThing",
	Fields: map[string]types.FieldConfig{
		"primary_key": {
			Name: "primary_key",
			Type: types.PRIMARY_KEY,
		},
		"string": {
			Name: "string",
			Type: types.STRING,
		},
		"number": {
			Name: "number",
			Type: types.NUMBER,
		},
		"boolean": {
			Name: "boolean",
			Type: types.BOOLEAN,
		},
		"date": {
			Name: "date",
			Type: types.DATE,
		},
		"thing": {
			Name:          "thing",
			Type:          types.THING,
			TypeThingName: "otherThing",
		},
		"manyToOne": {
			Name: "manyToOne",
			Type: types.RELATION,
			Relation: types.ThingRelation{
				Type:           types.MANY_TO_ONE,
				OtherThingName: "parentThing",
				OtherFieldName: "oneToMany",
			},
		},
	},
}

var otherThing = types.ThingConfig{
	Name: "otherThing",
	Fields: map[string]types.FieldConfig{
		"primary_key": {
			Name: "primary_key",
			Type: types.PRIMARY_KEY,
		},
		"string": {
			Name: "string",
			Type: types.STRING,
		},
	},
}
