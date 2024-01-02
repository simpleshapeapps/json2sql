package types

import (
	"fmt"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/exp/maps"
)

const (
	PRIMARY_KEY FieldType         = "PRIMARY_KEY"
	STRING      FieldType         = "STRING"
	NUMBER      FieldType         = "NUMBER"
	BOOLEAN     FieldType         = "BOOLEAN"
	DATE        FieldType         = "DATE"
	THING       FieldType         = "THING"
	RELATION    FieldType         = "RELATION"
	ONE_TO_MANY ThingRelationType = "ONE_TO_MANY"
	MANY_TO_ONE ThingRelationType = "MANY_TO_ONE"
)

type FieldType string
type ThingRelationType string

type ThingConfig struct {
	Name        string                 `json:"name"`
	Constraints ThingConstraints       `json:"constraints"`
	Fields      map[string]FieldConfig `json:"fields"`
}

type FieldConfig struct {
	Name          string        `json:"name"`
	Type          FieldType     `json:"type"`
	TypeThingName string        `json:"typeThingName"`
	Relation      ThingRelation `json:"relation"`
}

type ThingRelation struct {
	Type           ThingRelationType `json:"type"`
	OtherThingName string            `json:"otherThingName"`
	OtherFieldName string            `json:"otherFieldName"`
}

type ThingConstraints struct {
	AssignedToUser bool
}

var thingConfigMap = map[string]ThingConfig{}

func Register(thing ThingConfig) {
	thingConfigMap[thing.Name] = thing
}

func Clear() {
	thingConfigMap = map[string]ThingConfig{}
}

func (fc *FieldConfig) GetColumnName() string {
	result := strcase.ToSnake(fc.Name)
	if fc.Type == THING || fc.Type == RELATION {
		result += "_id"
	}
	return result
}

func (tc *ThingConfig) GetField(name string) (FieldConfig, error) {
	fieldConfig, ok := tc.Fields[name]
	if !ok {
		return FieldConfig{}, fmt.Errorf("field: %s not in thing: %s", name, tc.Name)
	}
	return fieldConfig, nil
}

func (tc *ThingConfig) GetFields() []FieldConfig {
	fields := maps.Values(tc.Fields)
	slices.SortFunc(fields, func(a, b FieldConfig) int {
		return strings.Compare(a.Name, b.Name)
	})
	return fields
}

func (tc *ThingConfig) GetTableName() string {
	return strcase.ToSnake(tc.Name)
}

func (tc *ThingConfig) GetFieldsNames() []string {
	result := []string{}
	for _, field := range tc.Fields {
		result = append(result, field.Name)
	}
	return result
}

func Get(name string) (ThingConfig, error) {
	thing, ok := thingConfigMap[name]
	if !ok {
		return ThingConfig{}, fmt.Errorf("thingConfig: %s doesn't exists", name)
	}
	return thing, nil
}
