package main

import (
	"fmt"
	"strconv"
	"strings"
)

type SearchQuery struct {
	FilterGroups []FilterGroup `json:"groups"`
}

type FilterGroup struct {
	Filters []Filter `json:"filters"`
}

type Filter struct {
	Type          string `json:"type"`
	Operation     string `json:"operation"`
	Value         string `json:"value"`
	CaseSensitive bool   `json:"case_sensitive"`
}

var cols = map[string]string{
	"name":       "name",
	"difficulty": "difficulty",
	"prep_time":  "prep_time",
	"rate":       "rating",
	"vegeterian": "vegeterian",
}

func Search(searchQuery SearchQuery) ([]Recipe, error) {
	results := []Recipe{}

	parsedFilters, err := parseFilters(searchQuery)
	if err != nil || len(parsedFilters) == 0 {
		return results, err
	}

	return db.GetRecipesByFilters(parsedFilters)
}

func parseFilters(query SearchQuery) (string, error) {
	parsedFilters := []string{}

	for _, group := range query.FilterGroups {
		conditions := []string{}
		for _, filter := range group.Filters {
			switch {
			case filter.Type == "name":
				condition, err := parseStringFilter(filter)
				if err != nil {
					return "", err
				}
				conditions = append(conditions, condition)
			case filter.Type == "difficulty" || filter.Type == "prep_time" || filter.Type == "rate":
				condition, err := parseNumericFilter(filter)
				if err != nil {
					return "", err
				}
				conditions = append(conditions, condition)
			case filter.Type == "vegeterian":
				condition, err := parseBoolFilter(filter)
				if err != nil {
					return "", err
				}
				conditions = append(conditions, condition)
			default:
				return "", fmt.Errorf("filter type %s is not supported.", filter.Type)
			}
		}

		parsedFilters = append(parsedFilters, fmt.Sprintf("(%s)", strings.Join(conditions, " AND ")))
	}

	return strings.Join(parsedFilters, " OR "), nil
}

func parseNumericFilter(filter Filter) (string, error) {
	condition := ""
	if filter.Operation != "=" && filter.Operation != ">=" && filter.Operation != "<=" && filter.Operation != ">" && filter.Operation != "<" && filter.Operation != "!=" {
		return condition, fmt.Errorf("filter operation '%s' for %s is not supported.", filter.Operation, filter.Type)
	}

	if _, err := strconv.ParseInt(filter.Value, 10, 8); err != nil {
		return condition, fmt.Errorf("filter value '%s' for %s is invalid.", filter.Value, filter.Type)
	}
	condition = fmt.Sprintf("a.%s %s %s", cols[filter.Type], filter.Operation, filter.Value)
	return condition, nil
}

func parseStringFilter(filter Filter) (string, error) {
	condition := ""
	op := "ILIKE"
	if filter.CaseSensitive {
		op = "LIKE"
	}
	switch filter.Operation {
	case "match":
		condition = fmt.Sprintf("%s %s '%s'", cols[filter.Type], op, db.EscapeQuotes(filter.Value))
	case "=":
		condition = fmt.Sprintf("%s %s '%s'", cols[filter.Type], op, db.EscapeQuotes(filter.Value))
	case "start":
		condition = fmt.Sprintf("%s %s '%s%%'", cols[filter.Type], op, db.EscapeQuotes(filter.Value))
	case "end":
		condition = fmt.Sprintf("%s %s '%%%s'", cols[filter.Type], op, db.EscapeQuotes(filter.Value))
	case "contain":
		condition = fmt.Sprintf("%s %s '%%%s%%'", cols[filter.Type], op, db.EscapeQuotes(filter.Value))
	default:
		return condition, fmt.Errorf("filter operation %s is not supported.", filter.Operation)
	}

	return condition, nil
}

func parseBoolFilter(filter Filter) (string, error) {
	val, err := strconv.ParseBool(filter.Value)
	if err != nil {
		return "", err
	}

	condition := fmt.Sprintf("%s = %v", cols[filter.Type], val)
	return condition, nil
}
