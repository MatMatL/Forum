package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const port = ":8080"

var index = template.Must(template.ParseFiles("index.html"))
var login = template.Must(template.ParseFiles("login.html"))
var register = template.Must(template.ParseFiles("register.html"))
var newpost = template.Must(template.ParseFiles("newPost.html"))
var post = template.Must(template.ParseFiles("post.html"))

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	InitTables()

	http.HandleFunc("/", Index)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/logout", Logout)
	http.HandleFunc("/newPost", NewPost)
	http.HandleFunc("/post", Post)

	fmt.Println("//localhost:8080")
	http.ListenAndServe(port, nil)
}

type Session struct {
	username    string
	isConnected bool
}

var CurrentSession = Session{}

func Index(w http.ResponseWriter, r *http.Request) {
	posts := getPosts()
	index.ExecuteTemplate(w, "index.html", posts)
}

type loginErrors struct {
	WrongEmail    bool
	WrongUsername bool
	WrongPassword bool
}

func Login(w http.ResponseWriter, r *http.Request) {
	if CurrentSession.isConnected {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	var currentErrors = loginErrors{}

	var UsernameInput string
	var PasswordInput string

	var readyToGo bool = false

	if r.Method == "POST" {
		time.Sleep(69 * time.Millisecond)

		r.ParseForm()
		UsernameInput = r.FormValue("userName")
		fmt.Println("User name : ", UsernameInput)

		PasswordInput = r.FormValue("userPassword")
		fmt.Println("User password : ", PasswordInput)

		readyToGo = true
	}

	time.Sleep(200 * time.Millisecond)

	if readyToGo {
		if loggingIn(UsernameInput, PasswordInput) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	login.ExecuteTemplate(w, "login.html", currentErrors)

}

func Register(w http.ResponseWriter, r *http.Request) {
	if CurrentSession.isConnected {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var currentErrors = loginErrors{}

	var EmailInput string
	var UsernameInput string
	var PasswordInput string

	if r.Method == "POST" {
		time.Sleep(69 * time.Millisecond)

		r.ParseForm()

		EmailInput = r.FormValue("newEmail")
		fmt.Println("User email : ", EmailInput)

		UsernameInput = r.FormValue("newUserName")
		fmt.Println("User name : ", UsernameInput)

		PasswordInput = r.FormValue("newUserPassword")
		fmt.Println("User password : ", PasswordInput)
	}

	if !ValidEmail(EmailInput) || AlreadyTakenEmail(EmailInput) {
		currentErrors.WrongEmail = true
		fmt.Println("invalide email or email already taken")
	} else if AlreadyTakenUsername(UsernameInput) {
		currentErrors.WrongUsername = true
		fmt.Println("Username already taken")
	} else {
		Request := `INSERT INTO Register (EMAIL,USERNAME,PASSWORD) VALUES (?, ?, ?);`
		fmt.Println("Send : ", Request)
		_, err := db.Exec(Request, EmailInput, UsernameInput, PasswordInput)
		if err != nil {
			log.Fatal(err)
		}

		PrintTable("Register")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	register.ExecuteTemplate(w, "register.html", currentErrors)
}

func InitTables() {
	Register := `
	CREATE TABLE Register (
		ID INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		EMAIL          TEXT    NOT NULL UNIQUE,
		USERNAME       TEXT    NOT NULL UNIQUE,
		PASSWORD       TEXT    NOT NULL
	);
	`
	db.Exec(Register)

	Posts := `
	CREATE TABLE Posts (
		ID INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		USERNAME          TEXT    NOT NULL,
		TITLE             TEXT    NOT NULL,
		CONTENT           TEXT    NOT NULL,
		CATEGORIES        TEXT    NOT NULL
	);
	`
	result, err := db.Exec(Posts)
	fmt.Println(result, err)
}

func PrintTable(tableName string) {
	rows, err := db.Query("SELECT * FROM " + tableName)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var titre string
		var auteur string
		var date string
		err = rows.Scan(&id, &titre, &auteur, &date)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, titre, auteur, date)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func ValidEmail(email string) bool {
	for _, i := range email {
		if i == '@' {
			return true
		}
	}
	return false
}

func AlreadyTakenEmail(EmailInput string) bool {
	row := db.QueryRow("SELECT EMAIL FROM Register WHERE EMAIL = ?", EmailInput)

	var email string
	err := row.Scan(&email)

	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			log.Fatal(err)
		}
	}

	return true
}

func AlreadyTakenUsername(UsernameInput string) bool {
	row := db.QueryRow("SELECT USERNAME FROM Register WHERE USERNAME = ?", UsernameInput)

	var username string
	err := row.Scan(&username)

	if err != nil {
		if err == sql.ErrNoRows {
			return false
		} else {
			log.Fatal(err)
		}
	}

	return true
}

func loggingIn(UsernameInput string, PasswordInput string) bool {
	row := db.QueryRow("SELECT PASSWORD FROM Register WHERE USERNAME = ?", UsernameInput)

	var password string
	err := row.Scan(&password)

	if err != nil {
		if err == sql.ErrNoRows {
		} else {
			log.Fatal(err)
		}
	} else {
		if password == PasswordInput {
			CurrentSession.isConnected = true
			CurrentSession.username = UsernameInput
			return true
		}
	}

	row = db.QueryRow("SELECT USERNAME, PASSWORD FROM Register WHERE EMAIL = ?", UsernameInput)

	var username string
	err = row.Scan(&username, &password)

	if err != nil {
		if err == sql.ErrNoRows {
		} else {
			log.Fatal(err)
		}
	} else {
		if password == PasswordInput {
			CurrentSession.isConnected = true
			CurrentSession.username = username
			return true
		}
	}
	return false
}

func Logout(w http.ResponseWriter, r *http.Request) {
	CurrentSession.isConnected = false
	CurrentSession.username = ""
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func NewPost(w http.ResponseWriter, r *http.Request) {
	if !CurrentSession.isConnected {
		fmt.Println("redirection")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var currentErrors = loginErrors{}

	var Title string
	var Content string
	var Categories string

	if r.Method == "POST" {
		time.Sleep(69 * time.Millisecond)

		r.ParseForm()

		Title = r.FormValue("postTitle")
		fmt.Println("Title : ", Title)

		Content = r.FormValue("postContent")
		fmt.Println("Content : ", Content)

		Categories = r.FormValue("postCategories")
		fmt.Println("categories : ", Categories)
	}

	if Title != "" {
		Request := `INSERT INTO Posts (USERNAME,TITLE,CONTENT,CATEGORIES) VALUES (?, ?, ?, ?);`
		fmt.Println("Send : ", Request)
		_, err := db.Exec(Request, CurrentSession.username, Title, Content, Categories)
		if err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	newpost.ExecuteTemplate(w, "newPost.html", currentErrors)
}

type PostData struct {
	ID         int
	Username   string
	Title      string
	Content    string
	Categories string
}

func getPosts() []PostData {
	rows, err := db.Query("SELECT ID, USERNAME, TITLE, CONTENT, CATEGORIES FROM Posts")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var posts []PostData
	for rows.Next() {
		var post PostData
		if err := rows.Scan(&post.ID, &post.Username, &post.Title, &post.Content, &post.Categories); err != nil {
			fmt.Println(err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return posts
}

func Post(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	postData, err := getPostByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	post.ExecuteTemplate(w, "post.html", postData)
}

func getPostByID(id int) (PostData, error) {
	var post PostData
	row := db.QueryRow("SELECT ID, USERNAME, TITLE, CONTENT, CATEGORIES FROM Posts WHERE ID = ?", id)
	err := row.Scan(&post.ID, &post.Username, &post.Title, &post.Content, &post.Categories)
	if err != nil {
		return post, err
	}
	return post, nil
}
