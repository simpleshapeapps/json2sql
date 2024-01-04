package types

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

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

func (fc FieldConfig) GetBool(valuesMap map[string]any) (bool, error) {
	fieldName := fc.Name
	value, ok := valuesMap[fieldName]
	if !ok {
		return false, fmt.Errorf("key: %s is not present in valuesMap", fieldName)
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	default:
		return false, fmt.Errorf("value: %v is not boolean", v)
	}
}

func (fc FieldConfig) GetString(valuesMap map[string]any) (string, error) {
	fieldName := fc.Name
	value, ok := valuesMap[fieldName]
	if !ok {
		return "", fmt.Errorf("key: %s is not present in valuesMap", fieldName)
	}

	switch v := value.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("value: %v is not string", v)
	}
}

func (fc FieldConfig) GetFloat64(valuesMap map[string]any) (float64, error) {
	fieldName := fc.Name
	value, ok := valuesMap[fieldName]
	if !ok {
		return 0, fmt.Errorf("key: %s is not present in valuesMap", fieldName)
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case []uint8:
		str := string(v)
		float, err := strconv.ParseFloat(str, 64)
		return float, err
	default:
		return 0, fmt.Errorf("value: %v is not float64", v)
	}
}

func (fc FieldConfig) GetDate(valuesMap map[string]any) (time.Time, error) {
	fieldName := fc.Name
	value, ok := valuesMap[fieldName]
	if !ok {
		return time.Time{}, fmt.Errorf("key: %s is not present in valuesMap", fieldName)
	}

	switch v := value.(type) {
	case time.Time:
		return v, nil
	default:
		return time.Time{}, fmt.Errorf("value: %v is not date", v)
	}
}
