# WikiWeb Web Application

Created for Code2College End of year Project in May, 2026

## Getting started
This repo works side by side with a MySQL Server that contains a table with the following attributes


### Schema of database.pages

| Field       | Type         | Null | Key | Default           | Extra             |
|-------------|--------------|------|-----|-------------------|-------------------|
| idpages     | int          | NO   | PRI |                   | auto_increment    |
| slug        | varchar(45)  | NO   | UNI |                   |                   |
| title       | varchar(45)  | NO   |     | Untitled          |                   |
| author      | varchar(45)  | NO   |     | Anonymous         |                   |
| views       | int          | NO   |     | 0                 |                   |
| content     | text         | NO   |     |                   |                   |
| pageType    | varchar(45)  | NO   |     | user              |                   |
| dateCreated | timestamp    | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
| LastUpdated | timestamp    | YES  |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |



### Populating .env

Make sure you set up environment variables in a file called .env
Follows the same template as .env.example

``` 
DB_PASSWORD=yourpassword
DB_USER=root
DB_NAME=yourdbname 
```
^ Fill these in with your values!!



### Running Program

1. run "go run wikiweb.go" in the terminal to start the program
2. If popup appears, allow access to network if you want to access website from other devices
3. Message will appear in terminal indicating web address where website is available 
Ex: ("...Server running at http://localhost:8080/home")
4. Go to http://localhost:8080/home to interact with web application



### Ending the program

To terminate the program,
1. Go to terminal
2. Press `Ctrl` + `C` while focused in terminal

This will terminate the program. You can start it again by following the steps above.