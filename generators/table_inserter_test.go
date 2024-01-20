package generators_test

import (
	"json2sql/generators"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	generator := generators.InsertIntoTable{
		ThingName: parentThing.Name,
		Values: map[string]any{
			"string":  "test",
			"number":  1,
			"boolean": true,
			"date":    time.Now(),
		},
	}

	sql, err := generator.GetSql()
	if err != nil {
		t.Fatal(err)
	}

	expected := `INSERT INTO "parent_thing" ("boolean", "date", "number", "string")
VALUES (:boolean, :date, :number, :string)`

	if sql != expected {
		t.Fatalf("expected: %s got: %s", expected, sql)
	}
}
