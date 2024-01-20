package generators_test

import (
	"json2sql/generators"
	"json2sql/types"
	"testing"
)

var parentThing = types.ThingConfig{
	Name: "parentThing",
	Fields: map[string]types.FieldConfig{
		"primaryKey": {
			Name: "primaryKey",
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
		"primaryKey": {
			Name: "primaryKey",
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
		"primaryKey": {
			Name: "primary_key",
			Type: types.PRIMARY_KEY,
		},
		"string": {
			Name: "string",
			Type: types.STRING,
		},
	},
}

func TestCreateTableWithAllFieldTypes(t *testing.T) {
	types.Clear()
	types.Register(parentThing)
	types.Register(childThing)
	types.Register(otherThing)

	generator := generators.CreateTable{
		ThingName: parentThing.Name,
	}
	sqls, err := generator.GetSql()

	if err != nil {
		t.Fatal(err)
	}

	if len(sqls) != 3 {
		t.Fatalf("expected 3 queries got: %d", len(sqls))
	}

	sql := sqls[0]
	expected := `CREATE TABLE IF NOT EXISTS "parent_thing" (
  "boolean" BOOLEAN,
  "date" DATE,
  "number" NUMERIC(18, 4),
  "primary_key" SERIAL PRIMARY KEY,
  "string" TEXT,
  "thing_id" SERIAL
)`
	if sql != expected {
		t.Fatalf("expected: %s have: %s", expected, sql)
	}

	sql = sqls[1]
	expected = `CREATE TABLE IF NOT EXISTS "child_thing" (
  "boolean" BOOLEAN,
  "date" DATE,
  "many_to_one_id" SERIAL,
  "number" NUMERIC(18, 4),
  "primary_key" SERIAL PRIMARY KEY,
  "string" TEXT,
  "thing_id" SERIAL
)`
	if sql != expected {
		t.Fatalf("expected: %s have: %s", expected, sql)
	}

	sql = sqls[2]
	expected = `CREATE TABLE IF NOT EXISTS "other_thing" (
  "primary_key" SERIAL PRIMARY KEY,
  "string" TEXT
)`
	if sql != expected {
		t.Fatalf("expected: %s have: %s", expected, sql)
	}
}

func TestUnregisteredFieldTypeThing(t *testing.T) {
	types.Clear()
	types.Register(parentThing)

	generator := generators.CreateTable{
		ThingName: parentThing.Name,
	}
	_, err := generator.GetSql()

	expectedError := `thingConfig: childThing doesn't exists
thingConfig: otherThing doesn't exists`

	if err == nil {
		t.Fatal("error expected")
	} else if err.Error() != expectedError {
		t.Fatalf("expected error: %s got: %s", expectedError, err)
	}
}
