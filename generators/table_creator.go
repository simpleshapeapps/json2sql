package generators

import (
	"errors"
	"fmt"
	"json2sql/types"
	"slices"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/exp/maps"
)

type CreateTable struct {
	Thing       types.ThingConfig
	otherThings []types.ThingConfig
}

func (ct *CreateTable) GetSql() ([]string, error) {
	results := []string{}
	var errs error
	sql, err := ct.getTableSql(ct.Thing)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	results = append(results, sql)
	for _, otherThing := range ct.otherThings {
		sql, err := ct.getTableSql(otherThing)
		if err != nil {
			errs = errors.Join(errs, err)
		}
		results = append(results, sql)
	}
	return results, errs
}

func (ct *CreateTable) getTableSql(thingConfig types.ThingConfig) (string, error) {
	tableName := thingConfig.GetTableName()
	fields, err := ct.getFieldCreateStrings(thingConfig)
	fieldsString := strings.Join(fields, ",\n")

	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
%s
)`, tableName, fieldsString), err

}

func (ct *CreateTable) getFieldCreateStrings(thingConfig types.ThingConfig) ([]string, error) {
	results := []string{}
	var errs error

	fields := maps.Values(thingConfig.Fields)
	slices.SortFunc(fields, func(a, b types.FieldConfig) int {
		return strings.Compare(a.Name, b.Name)
	})

	for _, field := range fields {
		fieldCreateString, err := GetTableFieldCreate(field)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if fieldCreateString != "" {
			formated := fmt.Sprintf("  %s", fieldCreateString)
			results = append(results, formated)
		}

		if field.Type == types.THING {
			otherThing, err := types.Get(field.TypeThingName)
			if err != nil {
				errs = errors.Join(errs, err)
			}
			ct.otherThings = append(ct.otherThings, otherThing)
		}

		if field.Type == types.RELATION {
			otherThing, err := types.Get(field.Relation.OtherThingName)
			if err != nil {
				errs = errors.Join(errs, err)
			}
			ct.otherThings = append(ct.otherThings, otherThing)
		}
	}

	return results, errs
}

func GetTableFieldCreate(field types.FieldConfig) (string, error) {
	fieldName := strcase.ToSnake(field.Name)
	switch field.Type {
	case types.PRIMARY_KEY:
		return fmt.Sprintf(`"%s" SERIAL PRIMARY KEY`, fieldName), nil
	case types.STRING:
		return fmt.Sprintf(`"%s" TEXT`, fieldName), nil
	case types.NUMBER:
		return fmt.Sprintf(`"%s" NUMERIC(18, 4)`, fieldName), nil
	case types.BOOLEAN:
		return fmt.Sprintf(`"%s" BOOLEAN`, fieldName), nil
	case types.DATE:
		return fmt.Sprintf(`"%s" DATE`, fieldName), nil
	case types.THING:
		return fmt.Sprintf(`"%s_id" SERIAL`, fieldName), nil
	case types.RELATION:
		if field.Relation.Type == types.MANY_TO_ONE {
			return fmt.Sprintf(`"%s_id" SERIAL`, fieldName), nil
		} else if field.Relation.Type == types.ONE_TO_MANY {
			return "", nil
		}
	}

	return "", fmt.Errorf("field type: %s is not supported", field.Type)
}
