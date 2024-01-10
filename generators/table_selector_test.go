package generators_test

import (
	"json2sql/generators"
	"json2sql/types"
	"testing"
	"time"
)

func TestSimpleSelect(t *testing.T) {
	types.Clear()
	types.Register(parentThing)

	s := generators.SelectFromTable{
		Thing: parentThing,
		FieldsMap: map[string]any{
			"string":  "",
			"boolean": 0,
			"number":  nil,
			"date":    time.Time{},
		},
		Page:  1,
		Count: 10,
	}

	query, err := s.GetSql()
	if err != nil {
		t.Fatal(err)
	}

	expected := `SELECT t."boolean" as "boolean", t."date" as "date", t."number" as "number", t."string" as "string"
FROM "parent_thing" t
LIMIT 10
OFFSET 0`

	if query != expected {
		t.Fatalf("expected: %s, got: %s", expected, query)
	}
}

func TestSimpleSelectWithWhere(t *testing.T) {
	types.Clear()
	types.Register(parentThing)

	s := generators.SelectFromTable{
		Thing: parentThing,
		FieldsMap: map[string]any{
			"string":  "",
			"boolean": 0,
			"number":  nil,
			"date":    time.Time{},
			"_where":  "string = 'test test' AND boolean = true",
		},
		Page:  1,
		Count: 10,
	}

	query, err := s.GetSql()
	if err != nil {
		t.Fatal(err)
	}

	expected := `SELECT t."boolean" as "boolean", t."date" as "date", t."number" as "number", t."string" as "string"
FROM "parent_thing" t
WHERE t."string" = $1 AND t."boolean" = $2
LIMIT 10
OFFSET 0`

	if query != expected {
		t.Fatalf("expected: %s, got: %s", expected, query)
	}

	whereValues := s.GetWhereValues()

	if len(whereValues) != 2 {
		t.Fatalf("expected 2 where values got: %d", len(whereValues))
	}

	if whereValues[0] != "test test" {
		t.Fatalf("expected: %s got: %s", "test test", whereValues[0])
	}

	if whereValues[1] != "true" {
		t.Fatalf("expected: %s got: %s", "true", whereValues[1])
	}
}
