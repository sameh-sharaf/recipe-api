package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func CreateRecipe(name string, prepTime int, difficulty int8, vegeterian bool) error {
	createdAt := time.Now().UTC()
	err := db.ExecInsertRecipeQuery(name, prepTime, difficulty, vegeterian, createdAt)
	return err
}

func DeleteRecipe(recipeID int64) error {
	idVal := strconv.FormatInt(recipeID, 10)

	// Delete all recipe rates first
	err := db.ExecDeleteQuery("app.rates", "recipeid", idVal)
	if err != nil {
		return err
	}

	// Then, delete it
	return db.ExecDeleteQuery("app.recipes", "id", idVal)
}

func UpdateRecipe(recipeID int64, params map[string]string) error {
	updateClauses := []string{}
	params["name"] = fmt.Sprintf("'%s'", db.EscapeQuotes(params["name"]))
	params["updatedat"] = fmt.Sprintf("'%s'", time.Now().UTC().Format(time.RFC3339))
	for col, value := range params {
		updateClauses = append(updateClauses, fmt.Sprintf("%s = %s", col, value))
	}

	query := fmt.Sprintf("UPDATE app.recipes SET %s WHERE id = %v;", strings.Join(updateClauses, ", "), recipeID)
	_, err = db.ExecQuery(query)
	return err
}

func RateRecipe(recipeID int64, rate int8) error {
	createdAt := time.Now().UTC()
	return db.ExecInsertRateQuery(recipeID, rate, createdAt)
}

func IsUserExists(username string) bool {
	_, err := db.GetUser(username)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return false
	}

	return true
}
