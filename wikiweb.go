package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/microcosm-cc/bluemonday"
)

// Struct Declarations
type PageData struct {
	Slug        string
	Title       string
	Author      string
	CreatedDate string
	LastUpdated string
	Content     string
}

type SearchPage struct {
	Query   string
	Count   int
	Results []PageData
}

type CreatePage struct {
	ErrorMessage string
	Title        string
	Author       string
	Content      string
	Slug         string
}

// db declaration so it can be accessed by handler
var db *sql.DB
var dbname string
var p *bluemonday.Policy
var s *bluemonday.Policy

func slugify(text string) string {
	text = strings.ToLower(text)
	alph := "abcdefghijklmnopqrstuvwxyz-_"
	slug := ""
	for i := 0; i < len(text); i++ {
		if strings.ContainsRune(alph, rune(text[i])) {
			slug += strings.ToLower(string(text[i]))
		} else if strings.EqualFold(string(text[i]), " ") {
			slug += "-"
		}
	}
	return slug
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// The home page
	// okay I want it to query the database to get all of the user gen pages
	data := []PageData{}

	// SQL command
	rows, err := db.Query("Select slug, title, author, dateCreated, LastUpdated, content From "+dbname+".pages where pageType = ?", "user")

	// Making sure we got results from the DB
	if err != nil {
		// throw an error
		log.Println("DB error:", err)
		return
	}
	defer rows.Close()

	// Fill in info for every row
	for rows.Next() {
		var p PageData
		rows.Scan(&p.Slug, &p.Title, &p.Author, &p.CreatedDate, &p.LastUpdated, &p.Content)
		data = append(data, p)
	}

	// later I'll do logic to show the most recent/most popular pages, but for now the generic home page
	tmpl, err := template.ParseFiles("templates/home.html")

	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl.Execute(w, data)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	// The slug of the webpage that the user visited on our site
	slug := r.URL.Path[1:]

	if slug == "" {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	// Query database for that page
	// PageData struct that holds all the info for the page from SQL
	var data PageData
	row := db.QueryRow("Select slug, title, author, dateCreated, LastUpdated, content From "+dbname+".pages Where slug = ?", slug)
	err := row.Scan(&data.Slug, &data.Title, &data.Author, &data.CreatedDate, &data.LastUpdated, &data.Content)

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

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// okay so the search query is in the url
	// step 1: get the search query

	query := r.URL.Query().Get("q")

	// step 2: search the database for results

	// SQL command
	rows, err := db.Query("Select slug, title, author, dateCreated, LastUpdated, content From pages where title like ? or content like ? order by LastUpdated DESC", "%"+query+"%", "%"+query+"%")

	// Making sure we got results from the DB
	if err != nil {
		// throw an error
		log.Println("DB error:", err)
		return
	}
	defer rows.Close()

	// step 3: fill in a PageData slice (auto grows) with any/all results

	data := []PageData{}

	// Fill in info for every row
	for rows.Next() {
		var p PageData
		rows.Scan(&p.Slug, &p.Title, &p.Author, &p.CreatedDate, &p.LastUpdated, &p.Content)
		data = append(data, p)
	}

	// step 4: submit the PageData slice into the searchpage html template

	// later I'll do logic to show the most recent/most popular pages, but for now the generic home page
	tmpl, err := template.ParseFiles("templates/search.html")

	page := SearchPage{Query: query, Count: len(data), Results: data}

	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl.Execute(w, page)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var info CreatePage
	// type CreatePage struct {
	// 	ErrorMessage string
	// 	Title		string
	// 	Author		string
	// 	Content		string
	if r.Method == "POST" {
		// okay so this is a html sent info with the completed data form

		// Follows same rules as edit
		info.Title = sanitizeUserInput(r.FormValue("title"), true)      // Strict Sanitize
		info.Author = sanitizeUserInput(r.FormValue("author"), true)    // Strict Sanitize
		info.Content = sanitizeUserInput(r.FormValue("content"), false) // Lose Sanitize
		info.ErrorMessage = isValidPageInfo(info)

		if len(info.ErrorMessage) > 0 {
			tmpl, err := template.ParseFiles("templates/createPage.html")

			if err != nil {
				log.Println("Template Error:", err)
				http.Error(w, "Internal Server Error", 500)
				return
			}
			tmpl.Execute(w, info)
			return
		}

		log.Println("User created a new webpage. Title:" + info.Title + ", Author:" + info.Author + ", Content:" + info.Content)
		// Valid-ish input- sanitize & submit to database

		// Submit new page to database
		info.Slug = slugify(info.Title)

		// SQL command
		_, err := db.Exec("INSERT INTO "+dbname+".pages (slug, title, author, content, pageType) VALUES (?, ?, ?, ?, ?)",
			info.Slug, info.Title, info.Author, info.Content, "user")

		if err != nil {
			log.Println("DB error:", err)
			http.Error(w, "Internal Server Error", 500)
			return
		}

		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/createPage.html")

	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl.Execute(w, info)
}

// Utilizes bluemonday package to sanitize either strictly (s) or normal (p)
func sanitizeUserInput(text string, strict bool) string {
	// sanitize user input so that the user doesn't put anything crazy in it
	if strict {
		return s.Sanitize(text)
	}
	return p.Sanitize(text)
}

func isValidPageInfo(info CreatePage) string {
	// okay so this is a html sent info with the completed data form
	// Goes through the same validation process- This can be a separate function.

	// Need to sanitize user fields. This is what info looks like
	// info.Title = r.FormValue("title")
	// info.Author = r.FormValue("author")
	// info.Content = r.FormValue("content")
	// info.ErrorMessage = ""

	errmsg := ""

	if info.Title == "" || info.Author == "" || info.Content == "" {
		errmsg += "All fields are required!! "
	}

	if len(info.Title) > 20 {
		errmsg += "Title too long! Title must be less than 20 characters! "
	}

	if len(info.Author) > 20 {
		errmsg += "Author too long! Author must be less than 20 characters! "
	}

	if len(info.Content) > 3000 {
		errmsg += "Content too long! Content must be less than 3000 characters! "
	}

	return errmsg
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	var info CreatePage
	// type CreatePage struct {
	// 	ErrorMessage string
	// 	Title		string
	// 	Author		string
	// 	Content		string

	if r.Method == "POST" {
		log.Println("Browser sent a POST request")
		// okay so this is a html sent info with the completed data form
		// Goes through the same validation process- This can be a separate function.

		// Need to sanitize everything from the browser cuz can be edited
		info.Title = sanitizeUserInput(r.FormValue("title"), true)      // Strict Sanitize
		info.Author = sanitizeUserInput(r.FormValue("author"), true)    // Strict Sanitize
		info.Content = sanitizeUserInput(r.FormValue("content"), false) // Lose Sanitize
		info.Slug = slugify(sanitizeUserInput(r.FormValue("p"), true))  // Strict Sanitize
		info.ErrorMessage = isValidPageInfo(info)

		// if there's an error,
		if len(info.ErrorMessage) > 0 {
			// Reload the page but with an error message to go with it (error message != "" so will be shown in template)
			tmpl, err := template.ParseFiles("templates/createPage.html")

			if err != nil {
				log.Println("Template Error:", err)
				http.Error(w, "Internal Server Error", 500)
				return
			}
			tmpl.Execute(w, info)
			return
		}
		// if there's not an error, push change(s)
		log.Println("User edited an existing webpage. Slug: " + info.Slug + "Title:" + info.Title + ", Author:" + info.Author + ", Content:" + info.Content)
		// Valid-ish input- sanitize & submit to database

		// Update a page in the database

		// // SQL command
		_, err := db.Exec("UPDATE "+dbname+".pages set title = ?, author = ?, content = ? where slug = ?",
			info.Title, info.Author, info.Content, info.Slug)

		if err != nil {
			log.Println("DB error:", err)
			http.Error(w, "Internal Server Error", 500)
			return
		}

		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	// NOT a POST request- Display page & prefil with info in the database
	slug := r.URL.Query().Get("p")
	log.Println("The user is trying to edit the page:" + slug)

	// 	type CreatePage struct { info struct
	// 	ErrorMessage string
	// 	Title        string
	// 	Author       string
	// 	Content      string
	// 	Slug         string

	// var info CreatePage
	row := db.QueryRow("Select slug, title, author, content From "+dbname+".pages Where slug = ?", slug)
	err := row.Scan(&info.Slug, &info.Title, &info.Author, &info.Content)

	// if no page is found, 404 error
	if err != nil { // the error from the sql database query
		http.NotFound(w, r)
		return // don't do anything after this if slug not found
	}

	tmpl, err := template.ParseFiles("templates/edit.html")

	if err != nil {
		log.Println("Template Error:", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl.Execute(w, info)
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

	// Defining sanitizing policies
	p = bluemonday.UGCPolicy()
	s = bluemonday.StrictPolicy()

	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/create-page", createHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/", pageHandler)

	// fmt.Println("Server running at http://localhost:8080/home")
	log.Println("Server running at http://localhost:8080/home")
	// Opening the http server
	http.ListenAndServe(":8080", nil)
}
