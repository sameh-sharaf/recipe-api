package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const n = 8
const recipePrefix = "Recipe_Test"

func TestInsertRecipe(t *testing.T) {
	validRecipe := getRandomRecipe()
	// Check # of test recipes before insert
	query := getStringSearchQuery("name", "match", validRecipe.Name, true)
	resultsBefore, _ := Search(query)

	// Insert new test recipe
	CreateRecipe(validRecipe.Name, validRecipe.PrepTime, validRecipe.Difficulty, validRecipe.Vegeterian)

	// Check search results after insert
	resultsAfter, _ := Search(query)

	if len(resultsBefore)+1 != len(resultsAfter) {
		t.Error(
			"For", "insert "+validRecipe.Name,
			"expected", 1,
			"got", len(resultsAfter)-len(resultsBefore),
		)
	}
}

func TestList(t *testing.T) {
	recipesCount := CountRecipes("")
	recipes, err := db.GetRecipes(0, 0, 0)
	if err != nil || len(recipes) != recipesCount {
		t.Error(
			"For", "list",
			"expected", recipesCount,
			"got", len(recipes),
		)
	}
}

func TestDelete(t *testing.T) {
	validRecipe := getRandomRecipe()
	// insert new test recipe
	CreateRecipe(validRecipe.Name, validRecipe.PrepTime, validRecipe.Difficulty, validRecipe.Vegeterian)

	// Look for inserted recipe
	query := getStringSearchQuery("name", "match", validRecipe.Name, true)
	results, _ := Search(query)
	if len(results) == 0 {
		t.Error(
			"For", "delete",
			"expected", "found inserted test recipe to delete",
			"got", "no recipe",
		)
	}

	// delete a test recipe
	deleteID := results[0].ID
	DeleteRecipe(int64(deleteID))
	recipes, _ := db.GetRecipes(deleteID, 0, 0)

	if len(recipes) != 0 {
		t.Error(
			"For", "delete",
			"expected", "no results",
			"got", len(recipes),
		)
	}
}

func TestGet(t *testing.T) {
	validRecipe := getRandomRecipe()
	// insert new test recipe
	CreateRecipe(validRecipe.Name, validRecipe.PrepTime, validRecipe.Difficulty, validRecipe.Vegeterian)

	// Look for inserted recipe
	query := getStringSearchQuery("name", "match", validRecipe.Name, true)
	results, _ := Search(query)
	if len(results) == 0 {
		t.Error(
			"For", "Get",
			"expected", "search results",
			"got", len(results),
		)
	}

	// Get test recipe by ID
	recipes, _ := db.GetRecipes(results[0].ID, 0, 0)
	if len(recipes) == 0 || !isMatched(recipes[0], validRecipe) {
		t.Error(
			"For", "Get",
			"expected", true,
			"got", false,
		)
	}
}

func TestRate(t *testing.T) {
	validRecipe := getRandomRecipe()

	// Insert new test recipe
	CreateRecipe(validRecipe.Name, validRecipe.PrepTime, validRecipe.Difficulty, validRecipe.Vegeterian)

	// Look for it
	query := getStringSearchQuery("name", "match", validRecipe.Name, true)
	results, _ := Search(query)
	if len(results) == 0 {
		t.Error(
			"For", "Rate",
			"expected", "one record",
			"got", "no records",
		)
	}

	// Count number of rates before
	rateBefore := CountRate(results[0].ID)

	// Rate it
	RateRecipe(results[0].ID, 5)

	// Check count of rates after
	rateAfter := CountRate(results[0].ID)

	if !(rateBefore < rateAfter) {
		t.Error(
			"For", "Rate",
			"expected", "new rate record",
			"got", "no new record",
		)
	}
}

func TestUpdate(t *testing.T) {
	// insert new recipe
	validRecipe := getRandomRecipe()
	CreateRecipe(validRecipe.Name, validRecipe.PrepTime, validRecipe.Difficulty, validRecipe.Vegeterian)

	// search for this recipe
	query := getStringSearchQuery("name", "match", validRecipe.Name, true)
	results, _ := Search(query)
	if len(results) == 0 {
		t.Error(
			"For", "Rate",
			"expected", "one record",
			"got", "no records",
		)
	}

	// update the recipe record
	validRecipe = getRandomRecipe()
	difficulty := strconv.FormatInt(int64(validRecipe.Difficulty), 10)
	prepTime := strconv.FormatInt(int64(validRecipe.PrepTime), 10)
	params := map[string]string{
		"name":       validRecipe.Name,
		"difficulty": difficulty,
		"prep_time":  prepTime,
	}
	UpdateRecipe(results[0].ID, params)

	// get the recipe after update
	recipes, _ := db.GetRecipes(results[0].ID, 0, 0)

	// should match
	if !isMatched(validRecipe, recipes[0]) {
		t.Error(
			"For", "Update",
			"expected", "match",
			"got", "no match",
		)
	}
}

func TestCleanUp(t *testing.T) {
	log.Println("Cleaning up previous test recipes..")
	query := getStringSearchQuery("name", "start", recipePrefix, false)
	results, _ := Search(query)

	log.Println("test recipes found:", len(results))
	for _, result := range results {
		DeleteRecipe(result.ID)
	}
}

func CountRecipes(recipeName string) int {
	whereClause := ""
	if len(recipeName) > 0 {
		whereClause = fmt.Sprintf("WHERE name = '%s'", recipeName)
	}

	count := 0
	query := fmt.Sprintf("SELECT COUNT(*) FROM app.recipes %s;", whereClause)
	res, err := db.ExecQuery(query)
	if err != nil {
		return count
	}

	if res.Next() {
		res.Scan(&count)
	}

	return count
}

func CountRate(recipeID int64) int {
	whereClause := fmt.Sprintf("WHERE recipeid = %v", recipeID)

	count := 0
	query := fmt.Sprintf("SELECT COUNT(*) FROM app.rates %s;", whereClause)
	res, err := db.ExecQuery(query)
	if err != nil {
		return count
	}

	if res.Next() {
		res.Scan(&count)
	}

	return count
}

func isMatched(recipe1, recipe2 Recipe) bool {
	return recipe1.Name == recipe2.Name && recipe1.PrepTime == recipe2.PrepTime && recipe1.Difficulty == recipe2.Difficulty
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getRandomRecipe() Recipe {
	return Recipe{0,
		"Recipe_Test" + RandStringRunes(n),
		rand.Intn(18000) + 1,
		int8(rand.Intn(2) + 1),
		rand.Intn(1) == 1,
		0.0,
		time.Now(),
		time.Now(),
	}
}

func getStringSearchQuery(filterType, operation, value string, caseSensitive bool) SearchQuery {
	filter := Filter{
		Type:          filterType,
		Operation:     operation,
		Value:         value,
		CaseSensitive: caseSensitive,
	}
	filterGroup := []FilterGroup{FilterGroup{[]Filter{filter}}}

	return SearchQuery{filterGroup}
}
