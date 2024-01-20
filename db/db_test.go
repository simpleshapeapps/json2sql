package main

import (
	"fmt"
	"json2sql/generators"
	"json2sql/types"
	"os"
	"testing"
	"time"

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

func TestCreateSimpleTable(t *testing.T) {
	types.Register(parentThing)
	types.Register(childThing)
	types.Register(otherThing)
	defer types.Clear()

	generator := generators.CreateTable{
		ThingName: parentThing.Name,
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

func TestSimpleInsertIntoTable(t *testing.T) {
	types.Register(parentThing)
	types.Register(childThing)
	types.Register(otherThing)
	defer types.Clear()

	createTable := generators.CreateTable{
		ThingName: parentThing.Name,
	}
	createTableSql, err := createTable.GetSql()
	if err != nil {
		t.Fatal(err)
	}

	date := time.Now().Truncate(24 * time.Hour)

	insert := generators.InsertIntoTable{
		ThingName: parentThing.Name,
		Values: map[string]any{
			"string":  "test string",
			"number":  1.1,
			"boolean": true,
			"date":    date,
		},
	}

	insertSql, err := insert.GetSql()
	if err != nil {
		t.Fatal(err)
	}

	doAndRollback(func(tx *sqlx.Tx) {
		tx.MustExec(createTableSql[0])

		result, err := tx.NamedExec(insertSql, insert.Values)
		if err != nil {
			t.Fatal(err)
		}

		ra, err := result.RowsAffected()
		if err != nil {
			t.Fatal(err)
		}

		if ra != 1 {
			t.Fatalf("expected row affected 1 got: %d", ra)
		}

		rows, err := tx.Queryx("SELECT string, boolean, date, number FROM parent_thing")
		if err != nil {
			t.Fatal(err)
		}

		rows.Next()

		m := map[string]interface{}{}
		err = rows.MapScan(m)
		if err != nil {
			t.Fatal(err)
		}

		b, err := parentThing.Fields["boolean"].GetBool(m)
		if err != nil || b != true {
			t.Fatalf("expected: true got: %v, err: %v", b, err)
		}

		s, err := parentThing.Fields["string"].GetString(m)
		if err != nil || s != "test string" {
			t.Fatalf("expected: test string got: %v, err: %v", s, err)
		}

		n, err := parentThing.Fields["number"].GetFloat64(m)
		if err != nil || n != 1.1 {
			t.Fatalf("expected: 1.1 got: %v, err: %v", n, err)
		}

		d, err := parentThing.Fields["date"].GetDate(m)
		if err != nil || !d.Equal(date) {
			t.Fatalf("expected: %v got: %v, err: %v", date, n, err)
		}

		if rows.Next() != false {
			t.Fatal("expected only one row")
		}
	})
}

func TestSimpleSelect(t *testing.T) {
	types.Register(parentThing)
	types.Register(childThing)
	types.Register(otherThing)
	defer types.Clear()

	createTable := generators.CreateTable{
		ThingName: parentThing.Name,
	}

	date := time.Now().Truncate(24 * time.Hour)

	insert := generators.InsertIntoTable{
		ThingName: parentThing.Name,
		Values: map[string]any{
			"string":  "test string",
			"number":  1.1,
			"boolean": true,
			"date":    date,
		},
	}

	insert2 := generators.InsertIntoTable{
		ThingName: parentThing.Name,
		Values: map[string]any{
			"string":  "test string",
			"number":  1.1,
			"boolean": false,
			"date":    date,
		},
	}

	s := generators.SelectFromTable{
		ThingName: parentThing.Name,
		FieldsMap: map[string]any{
			"string":  "",
			"number":  "",
			"boolean": "",
			"date":    "",
			"_where":  "boolean = true AND number = 1.1",
		},
	}

	s2 := generators.SelectFromTable{
		ThingName: parentThing.Name,
		FieldsMap: map[string]any{
			"string":  "",
			"number":  "",
			"boolean": "",
			"date":    "",
			"_where":  "string = 'test string'",
		},
	}

	doAndRollback(func(tx *sqlx.Tx) {
		err := executeCreateTable(&createTable, tx)
		if err != nil {
			t.Fatal(err)
		}

		err = executeInsert(&insert, tx)
		if err != nil {
			t.Fatal(err)
		}

		err = executeInsert(&insert2, tx)
		if err != nil {
			t.Fatal(err)
		}

		result, err := executeSelect(&s, tx)
		if err != nil {
			t.Fatal(err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 row got: %d", len(result))
		}

		first := result[0]

		b, err := parentThing.Fields["boolean"].GetBool(first)
		if err != nil || b != true {
			t.Fatalf("expected: true got: %v, err: %v", b, err)
		}

		s, err := parentThing.Fields["string"].GetString(first)
		if err != nil || s != "test string" {
			t.Fatalf("expected: test string got: %v, err: %v", s, err)
		}

		n, err := parentThing.Fields["number"].GetFloat64(first)
		if err != nil || n != 1.1 {
			t.Fatalf("expected: 1.1 got: %v, err: %v", n, err)
		}

		d, err := parentThing.Fields["date"].GetDate(first)
		if err != nil || !d.Equal(date) {
			t.Fatalf("expected: %v got: %v, err: %v", date, n, err)
		}

		result, err = executeSelect(&s2, tx)
		if err != nil {
			t.Fatal(err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 1 row got: %d", len(result))
		}
	})
}

func executeCreateTable(ct *generators.CreateTable, tx *sqlx.Tx) error {
	createTableSql, err := ct.GetSql()
	if err != nil {
		return err
	}

	tx.MustExec(createTableSql[0])
	return nil
}

func executeInsert(iit *generators.InsertIntoTable, tx *sqlx.Tx) error {
	insertSql, err := iit.GetSql()
	if err != nil {
		return err
	}

	_, err = tx.NamedExec(insertSql, iit.Values)
	if err != nil {
		return err
	}

	return nil
}

func executeSelect(s *generators.SelectFromTable, tx *sqlx.Tx) ([]map[string]any, error) {
	result := []map[string]any{}
	query, err := s.GetSql()
	if err != nil {
		return result, err
	}

	whereValues := s.GetWhereValues()
	rows, err := tx.Queryx(query, whereValues...)
	if err != nil {
		return result, err
	}

	for rows.Next() {
		m := map[string]interface{}{}
		err = rows.MapScan(m)
		if err != nil {
			return result, err
		} else {
			result = append(result, m)
		}
	}

	return result, nil
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
