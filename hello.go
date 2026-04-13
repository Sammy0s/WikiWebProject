package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/home.html")
	tmpl.Execute(w, nil)
}

func main() {
	fmt.Println("type localhost:8080/home into your browser")

	http.HandleFunc("/home", homeHandler)

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "About page!")
	})

	// This MUST be last — it starts the server and blocks everything after it
	http.ListenAndServe(":8080", nil)
}
