package generators

import (
	"errors"
	"fmt"
	"json2sql/types"
	"strings"

	"github.com/iancoleman/strcase"
)

type CreateTable struct {
	ThingName   string
	otherThings []types.ThingConfig
	thing       types.ThingConfig
}

func (ct *CreateTable) GetSql() ([]string, error) {
	thing, err := types.Get(ct.ThingName)
	if err != nil {
		return []string{}, err
	}
	ct.thing = thing

	results := []string{}
	var errs []error
	sql, err := ct.getTableSql(ct.thing)
	if err != nil {
		errs = append(errs, err)
	}
	results = append(results, sql)
	for _, otherThing := range ct.otherThings {
		sql, err := ct.getTableSql(otherThing)
		if err != nil {
			errs = append(errs, err)
		}
		results = append(results, sql)
	}
	return results, errors.Join(errs...)
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
	var errs []error

	for _, field := range thingConfig.GetFields() {
		fieldCreateString, err := GetTableFieldCreate(field)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if fieldCreateString != "" {
			formated := fmt.Sprintf("  %s", fieldCreateString)
			results = append(results, formated)
		}

		if field.Type == types.THING {
			otherThing, err := types.Get(field.TypeThingName)
			if err != nil {
				errs = append(errs, err)
			}
			ct.otherThings = append(ct.otherThings, otherThing)
		}

		if field.Type == types.RELATION {
			otherThing, err := types.Get(field.Relation.OtherThingName)
			if err != nil {
				errs = append(errs, err)
			}
			ct.otherThings = append(ct.otherThings, otherThing)
		}
	}

	return results, errors.Join(errs...)
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
