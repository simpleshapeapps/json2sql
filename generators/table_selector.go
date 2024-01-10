package generators

import (
	"errors"
	"fmt"
	"json2sql/parsers"
	"json2sql/types"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type SelectFromTable struct {
	Thing         types.ThingConfig
	FieldsMap     map[string]any
	Page          uint
	Count         uint
	columnsString string
	whereString   string
	whereValues   []any
}

type SelectColumn struct {
	columnName     string
	aliasName      string
	tableAliasName string
}

const (
	mainTableAlias = "t"
)

func (s *SelectFromTable) GetSql() (string, error) {
	err := s.prepareSelect()
	if err != nil {
		return "", nil
	}

	mainTableName := strcase.ToSnake(s.Thing.Name)

	query := fmt.Sprintf("SELECT %s\n"+
		"FROM \"%s\" %s", s.columnsString, mainTableName, mainTableAlias)

	if s.whereString != "" {
		query += fmt.Sprintf("\nWHERE %s", s.whereString)
	}

	if s.Count > 0 {
		limit := strconv.FormatUint(uint64(s.Count), 10)
		query += fmt.Sprintf("\nLIMIT %s", limit)

		offset := (uint64(s.Page) - uint64(1)) * uint64(s.Count)
		offsetString := strconv.FormatUint(offset, 10)
		query += fmt.Sprintf("\nOFFSET %s", offsetString)
	}

	return query, nil
}

func (s *SelectFromTable) GetWhereValues() []any {
	return s.whereValues
}

func (s *SelectFromTable) prepareSelect() error {
	var errs []error
	thingConfig := s.Thing

	keys := maps.Keys(s.FieldsMap)
	slices.SortFunc(keys, func(a, b string) int {
		return strings.Compare(a, b)
	})

	for _, fieldName := range keys {
		if strings.HasPrefix(fieldName, "_") {
			continue
		}

		fieldConfig, err := thingConfig.GetField(fieldName)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		selectColumn := SelectColumn{
			columnName:     fieldConfig.GetColumnName(),
			aliasName:      fieldName,
			tableAliasName: mainTableAlias,
		}

		s.columnsString += GetColumnString(selectColumn)
	}

	s.columnsString = strings.TrimSuffix(s.columnsString, ", ")
	whereString, err := s.GetWhereString()
	if err != nil {
		errs = append(errs, err)
	} else {
		s.whereString = whereString
	}

	return errors.Join(errs...)
}

func GetColumnString(column SelectColumn) string {
	alias := column.columnName
	if column.aliasName != "" {
		alias = column.aliasName
	}
	return fmt.Sprintf(`%s."%s" as "%s", `, column.tableAliasName, column.columnName, alias)

}

func (s *SelectFromTable) GetWhereString() (string, error) {
	result := ""
	w, ok := s.FieldsMap["_where"]

	if !ok {
		return "", nil
	}

	whereValue, ok := w.(string)
	if !ok {
		return "", fmt.Errorf("_where must be string")
	}

	parser := parsers.Parser{}
	tokens := parser.Parse(whereValue)
	if len(tokens) <= 0 {
		return "", fmt.Errorf("_where is empty")
	}

	thing := s.Thing
	for _, token := range tokens {
		whereField, isField := thing.Fields[token]
		if isField {
			if whereField.Type == types.NUMBER {
				result += fmt.Sprintf(`COALESCE(%s."%s", 0) `, mainTableAlias, whereField.GetColumnName())
			} else {
				result += fmt.Sprintf(`%s."%s" `, mainTableAlias, whereField.GetColumnName())
			}
			continue
		}

		compareToken := getCompareToken(token)
		if compareToken != "" {
			result += compareToken + " "
			continue
		}

		logicalToken := getLogicalToken(token)
		if logicalToken != "" {
			result += " " + logicalToken + " "
			continue
		}

		result += fmt.Sprintf("$%d", len(s.whereValues)+1)
		s.whereValues = append(s.whereValues, token)
	}

	return result, nil
}

func getCompareToken(token string) string {
	compareTokens := []string{"<", ">", "<=", ">=", "="}
	isMathOperator := slices.Contains(compareTokens, token)
	if isMathOperator {
		return token
	}
	return ""
}

func getLogicalToken(token string) string {
	logicalOperators := []string{"and", "AND", "or", "OR"}
	isLogicalOrerator := slices.Contains(logicalOperators, token)
	if isLogicalOrerator {
		return token
	}
	return ""
}
