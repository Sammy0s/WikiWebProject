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
	fmt.Println("Go to localhost:8080/home in your browser")

	http.HandleFunc("/home", homeHandler)

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "About page!")
	})

	// Opening the http server
	http.ListenAndServe(":8080", nil)
}
