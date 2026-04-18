package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// db declaration so it can be accessed by handler
var db *sql.DB
var dbname string

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// The home page

	// later I'll do logic to show the most recent/most popular pages, but for now the generic home page
	tmpl, err := template.ParseFiles("templates/home.html")

	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl.Execute(w, nil)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	// The slug of the webpage that the user visited on our site
	slug := r.URL.Path[1:]

	// Query database for that page
	// Stores the content of a webpage from MySQL Database
	var content string // Declaring a variable w/o a value so explicity state the type
	row := db.QueryRow("Select content From "+dbname+".pages Where slug = ?", slug)
	err := row.Scan(&content)

	// if no page is found, 404 error
	if err != nil {
		http.NotFound(w, r)
		return // don't do anything after this if slug not found
	}

	// Send content from db to browser
	fmt.Fprint(w, content)
}

func main() {
	var err error

	// Loading the .env file
	godotenv.Load()

	// Getting the env variables (secrets o.o")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbname = os.Getenv("DB_NAME")

	db, err = sql.Open("mysql", user+":"+pass+"@tcp(localhost:3306)/"+dbname)
	if err != nil { // If the error wasn't nothing (if there was an error)
		log.Fatal(err) // panic(err) ??
	}

	perr := db.Ping()
	if perr != nil {
		log.Fatal("Could not connect to database:", perr)
	}

	defer db.Close()

	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/", pageHandler)

	// fmt.Println("Server running at http://localhost:8080/home")
	log.Println("Server running at http://localhost:8080/home")
	// Opening the http server
	http.ListenAndServe(":8080", nil)
}
