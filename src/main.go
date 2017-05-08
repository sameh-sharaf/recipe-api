package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/register", RegisterHandler)
	router.HandleFunc("/login", LoginHandler)
	router.HandleFunc("/logout", LogoutHandler)

	router.HandleFunc("/recipes", RecipesHandler)
	router.HandleFunc("/recipes/{id}", RecipeHandler)
	router.HandleFunc("/recipes/{id}/rate", RateHandler)
	router.HandleFunc("/search", SearchHandler)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "This is the main page")
	})

	http.Handle("/", router)
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil)
}
