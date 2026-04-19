package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Struct Declarations
type PageData struct {
	Title       string
	Author      string
	CreatedDate string
	LastUpdated string
	Content     string
}

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
	// PageData struct that holds all the info for the page from SQL
	var data PageData
	row := db.QueryRow("Select title, author, dateCreated, LastUpdated, content From "+dbname+".pages Where slug = ?", slug)
	err := row.Scan(&data.Title, &data.Author, &data.CreatedDate, &data.LastUpdated, &data.Content)

	// if no page is found, 404 error
	if err != nil { // the error from the sql database query
		http.NotFound(w, r)
		return // don't do anything after this if slug not found
	}

	// Use the template and fill in this page's info
	// The Pages template
	tmpl, err := template.ParseFiles("templates/userPage.html")

	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	// Fill the UserPage template with the url's data
	tmpl.Execute(w, data)
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
