package generators

import (
	"fmt"
	"json2sql/types"
	"strings"
)

type InsertIntoTable struct {
	Thing  types.ThingConfig
	Values map[string]any
}

func (iit *InsertIntoTable) GetSql() (string, error) {
	thing := iit.Thing
	valuesMap := iit.Values

	intoString := ""
	valuesString := ""

	for _, field := range thing.GetFields() {
		valueFieldName := field.Name
		if field.Type == types.RELATION || field.Type == types.THING {
			valueFieldName = field.Name + "Id"
		}

		if valueFieldName == "id" || valuesMap[valueFieldName] == nil {
			continue
		}

		intoString += fmt.Sprintf(`"%s", `, field.GetColumnName())
		valuesString += fmt.Sprintf(":%s, ", valueFieldName)
	}

	intoString = strings.TrimSuffix(intoString, ", ")
	valuesString = strings.TrimSuffix(valuesString, ", ")

	query := fmt.Sprintf(`INSERT INTO "%s" (`+intoString+`)
VALUES (`+valuesString+`)`, thing.GetTableName())

	return query, nil
}
