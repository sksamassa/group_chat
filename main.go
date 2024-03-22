package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "404 - Page not found")
	})

	if err := http.ListenAndServe(":4004", r); err != nil {
		panic(err)
	}
}
