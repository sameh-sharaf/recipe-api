package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	fullName := strings.TrimSpace(r.FormValue("fullname"))

	if len(username) == 0 || len(password) == 0 || len(fullName) == 0 {
		fmt.Fprint(w, "one or more credentials are missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if IsUserExists(username) {
		fmt.Fprint(w, "user already exists")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	passwordHash, err := sessionManager.EncryptPassword(password)
	if err != nil {
		log.Println("could not encrypt password", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.InsertUser(username, fullName, passwordHash)
	if err != nil {
		log.Println("could not add new user:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	if len(username) == 0 || len(password) == 0 {
		fmt.Fprint(w, "credentials are not provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := db.GetUser(username)
	if err != nil {
		log.Println("could not get user:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Println("could not compare hashed password:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = sessionManager.SessionStart(w, r, user)
	if err != nil {
		log.Println("could not start session:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sid, err := r.Cookie(os.Getenv("COOKIE_SID"))
	if err != nil {
		log.Println("could not get session key from cookie", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sessionKey, _ := url.QueryUnescape(sid.Value)
	sessionManager.DestroySession(sessionKey)
}

func RecipesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ListHandler(w, r)
	case "POST":
		CreateHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func RecipeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		GetHandler(w, r)
	case "PUT":
		UpdateHandler(w, r)
	case "PATCH":
		UpdateHandler(w, r)
	case "DELETE":
		DeleteHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	var items int64
	var page int64
	var err error

	itemsVal := strings.TrimSpace(r.FormValue("items"))
	if len(itemsVal) != 0 {
		items, err = strconv.ParseInt(itemsVal, 10, 32)
		if err != nil {
			log.Println("items number is not valid", err)
			items = 0
		}
	}

	pageVal := strings.TrimSpace(r.FormValue("page"))
	if len(pageVal) != 0 {
		page, err = strconv.ParseInt(pageVal, 10, 32)
		if err != nil {
			log.Println("page number is not valid", err)
			page = 0
		}
	}

	recipes, err := db.GetRecipes(0, int(items), int(page))
	if err != nil {
		log.Println("could not list recipes:", err)
		fmt.Fprint(w, "could not list recipes")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(recipes)
	if err != nil {
		log.Println("could not parse JSON:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/json")
	w.Write(b)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(w, r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if len(name) == 0 {
		fmt.Fprint(w, "name is not set")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	prepTime, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("prep_time")), 10, 32)
	if err != nil || prepTime < 0 {
		log.Println("could not parse prep_time:", err)
		fmt.Fprint(w, "prep_time is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	difficulty, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("difficulty")), 10, 8)
	if err != nil || difficulty < 1 || difficulty > 3 {
		log.Println("could not parse difficulty:", err)
		fmt.Fprint(w, "difficulty is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vegeterian, err := strconv.ParseBool(strings.TrimSpace(r.FormValue("vegeterian")))
	if err != nil {
		log.Println("could not parse vegeterian:", err)
		fmt.Fprint(w, "vegeterian is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := CreateRecipe(name, int(prepTime), int8(difficulty), vegeterian); err != nil {
		log.Println("cannot add new recipe:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) == 0 {
		fmt.Fprint(w, "id is not set")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recipeID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		log.Println("could not parse recipeID", err)
		fmt.Fprintf(w, "ERROR: INVALID ID; %s", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recipes, err := db.GetRecipes(recipeID, 0, 0)
	if err != nil {
		log.Println("could not get recipe(s):", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(recipes) == 0 {
		log.Println("no recipe found")
		return
	}

	b, err := json.Marshal(recipes)
	if err != nil {
		log.Printf("could not convert to JSON: %s\r\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write(b)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(w, r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) == 0 {
		fmt.Fprint(w, "id is not set")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recipeID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Println("could not parse recipeID:", err)
		fmt.Fprintf(w, "ERROR: INVALID ID; %s", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if recipe ID provided exists
	recipes, err := db.GetRecipes(recipeID, 0, 0)
	if err != nil {
		log.Println("could not update recipe", err)
		fmt.Fprintf(w, "could not update recipe")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(recipes) == 0 {
		fmt.Fprintf(w, "recipe id provided does not exist")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	updateMap := make(map[string]string)
	name := strings.TrimSpace(r.FormValue("name"))
	if len(name) != 0 {
		updateMap["name"] = name
	}

	prepTime := strings.TrimSpace(r.FormValue("prep_time"))
	if len(prepTime) != 0 {
		_, err := strconv.ParseInt(prepTime, 10, 64)
		if err != nil {
			fmt.Fprint(w, "prep_time is not valid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		updateMap["prep_time"] = prepTime
	}

	difficulty := strings.TrimSpace(r.FormValue("difficulty"))
	if len(difficulty) != 0 {
		_, err := strconv.ParseInt(difficulty, 10, 64)
		if err != nil {
			fmt.Fprint(w, "difficulty is not valid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		updateMap["difficulty"] = difficulty
	}

	vegeterian := strings.TrimSpace(r.FormValue("vegeterian"))
	var vegBool bool
	if len(vegeterian) != 0 {
		vegBool, err = strconv.ParseBool(vegeterian)
		if err != nil {
			fmt.Fprint(w, "vegeterian is not valid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		updateMap["vegeterian"] = strconv.FormatBool(vegBool)
	}

	if len(updateMap) == 0 {
		fmt.Fprint(w, "no update params was provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = UpdateRecipe(recipeID, updateMap)
	if err != nil {
		log.Println("could not update recipe:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(w, r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) == 0 {
		fmt.Fprint(w, "id is not set")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recipeID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		fmt.Fprintf(w, "ERROR: INVALID ID; %s", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = DeleteRecipe(recipeID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func RateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "PATCH" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	recipeID, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		fmt.Fprintf(w, "ERROR: INVALID ID %s", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ratingVal := strings.TrimSpace(r.FormValue("rating"))
	if len(ratingVal) == 0 {
		fmt.Fprint(w, "rating is not set")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rating, err := strconv.ParseInt(ratingVal, 10, 8)
	if err != nil || !(rating > 0 && rating <= 5) {
		fmt.Fprintf(w, "invalid rating score: %s", r.FormValue("rating"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = RateRecipe(recipeID, int8(rating))
	if err != nil {
		log.Println("could not rate recipe:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	query := strings.TrimSpace(r.FormValue("query"))
	if len(query) == 0 {
		fmt.Fprint(w, "No search query was provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	searchQuery := SearchQuery{}
	err := json.Unmarshal([]byte(query), &searchQuery)
	if err != nil {
		log.Println("could not unmarshal search query", err)
		fmt.Fprint(w, "Invalid search query")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	results, err := Search(searchQuery)
	if err != nil {
		log.Println("could not search db:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(results)
	if err != nil {
		log.Println("could not convert results to JSON", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/json")
	w.Write(b)
}
