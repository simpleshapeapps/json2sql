package generators

import (
	"errors"
	"fmt"
	"json2sql/types"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

type InsertIntoTable struct {
	ThingName string
	Values    map[string]any
	thing     types.ThingConfig
}

func (iit *InsertIntoTable) GetValuesFieldNames() []string {
	fieldNames := maps.Keys(iit.Values)
	sort.Strings(fieldNames)
	return fieldNames
}

func (iit *InsertIntoTable) GetSql() (string, error) {
	var errs []error
	thing, err := types.Get(iit.ThingName)
	if err != nil {
		return "", err
	}
	intoString := ""
	valuesString := ""

	for _, fieldName := range iit.GetValuesFieldNames() {
		field, err := thing.GetField(fieldName)

		if err != nil {
			errs = append(errs, err)
			continue
		}

		if field.Type == types.PRIMARY_KEY {
			err := errors.New("cannot insert primary key")
			errs = append(errs, err)
			continue
		}

		valueFieldName := field.Name
		if field.Type == types.RELATION || field.Type == types.THING {
			valueFieldName = field.Name + "Id"
		}

		intoString += fmt.Sprintf(`"%s", `, field.GetColumnName())
		valuesString += fmt.Sprintf(":%s, ", valueFieldName)
	}

	intoString = strings.TrimSuffix(intoString, ", ")
	valuesString = strings.TrimSuffix(valuesString, ", ")

	query := fmt.Sprintf(`INSERT INTO "%s" (`+intoString+`)
VALUES (`+valuesString+`)`, thing.GetTableName())

	return query, errors.Join(errs...)
}
