package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

////////////////////////////////////////////////////////////////
/////////                MAIN ; SERVER                 /////////
////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////
/////////                  DATABASE                    /////////
////////////////////////////////////////////////////////////////

func InitTables() {
	Register := `
	CREATE TABLE IF NOT EXISTS Register (
		ID INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		EMAIL          TEXT    NOT NULL UNIQUE,
		USERNAME       TEXT    NOT NULL UNIQUE,
		PASSWORD       TEXT    NOT NULL
	);
	`
	db.Exec(Register)

	Posts := `
	CREATE TABLE IF NOT EXISTS Posts (
		ID INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		USERNAME          TEXT    NOT NULL,
		TITLE             TEXT    NOT NULL,
		CONTENT           TEXT    NOT NULL,
		CATEGORIES        TEXT    NOT NULL
	);
	`
	db.Exec(Posts)

	Sessions := `
	CREATE TABLE IF NOT EXISTS Sessions (
		UUID TEXT PRIMARY KEY UNIQUE,
		USERNAME TEXT NOT NULL,
		EXPIRATION DATETIME NOT NULL
	);
	`
	db.Exec(Sessions)
}

////////////////////////////////////////////////////////////////
/////////                   INDEX                      /////////
////////////////////////////////////////////////////////////////

func Index(w http.ResponseWriter, r *http.Request) {
	posts := getPosts()
	index.ExecuteTemplate(w, "index.html", posts)
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

////////////////////////////////////////////////////////////////
/////////                   LOGIN                      /////////
////////////////////////////////////////////////////////////////

func Login(w http.ResponseWriter, r *http.Request) {
	username := CheckCookies(w, r)
	if username != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	if r.Method == "POST" {
		r.ParseForm()
		UsernameInput := r.FormValue("userName")
		PasswordInput := r.FormValue("userPassword")

		if loggingIn(UsernameInput, PasswordInput) {
			sessionID, err := uuid.NewV4()
			if err != nil {
				log.Fatal(err)
			}
			expiration := time.Now().Add(24 * time.Hour)
			_, err = db.Exec("INSERT INTO Sessions (UUID, USERNAME, EXPIRATION) VALUES (?, ?, ?)", sessionID.String(), UsernameInput, expiration)
			if err != nil {
				log.Fatal(err)
			}

			cookie := &http.Cookie{
				Name:    "session",
				Value:   sessionID.String(),
				Expires: expiration,
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	login.ExecuteTemplate(w, "login.html", nil)
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
		if PasswordIsGood(password, PasswordInput) {
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
		if PasswordIsGood(password, PasswordInput) {
			return true
		}
	}
	return false
}

////////////////////////////////////////////////////////////////
/////////                  REGISTER                    /////////
////////////////////////////////////////////////////////////////

type loginErrors struct {
	WrongEmail    bool
	WrongUsername bool
	WrongPassword bool
}

func Register(w http.ResponseWriter, r *http.Request) {
	username := CheckCookies(w, r)
	if username != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	var currentErrors = loginErrors{}

	if r.Method == "POST" {
		r.ParseForm()
		EmailInput := r.FormValue("newEmail")
		UsernameInput := r.FormValue("newUserName")
		PasswordInput := r.FormValue("newUserPassword")

		if !ValidEmail(EmailInput) || AlreadyTakenEmail(EmailInput) {
			currentErrors.WrongEmail = true
		} else if AlreadyTakenUsername(UsernameInput) {
			currentErrors.WrongUsername = true
		} else {
			Request := `INSERT INTO Register (EMAIL,USERNAME,PASSWORD) VALUES (?, ?, ?);`
			HachedPassword, errHash := bcrypt.GenerateFromPassword([]byte(PasswordInput), 10)
			if errHash != nil {
				log.Fatal(errHash)
			}
			_, err := db.Exec(Request, EmailInput, UsernameInput, string(HachedPassword))
			if err != nil {
				log.Fatal(err)
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	register.ExecuteTemplate(w, "register.html", currentErrors)
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

func PasswordIsGood(password string, PasswordInput string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(PasswordInput))

	if err == nil {
		return true
	}
	fmt.Println("Wrrong password: ", err)
	return false
}

////////////////////////////////////////////////////////////////
/////////                   LOGOUT                     /////////
////////////////////////////////////////////////////////////////

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		db.Exec("DELETE FROM Sessions WHERE UUID = ?", cookie.Value)
		cookie := &http.Cookie{
			Name:   "session",
			Value:  "",
			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

////////////////////////////////////////////////////////////////
/////////                  NEW POST                    /////////
////////////////////////////////////////////////////////////////

func NewPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var username string
	err = db.QueryRow("SELECT USERNAME FROM Sessions WHERE UUID = ?", cookie.Value).Scan(&username)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var Title, Content, Categories string

	if r.Method == "POST" {
		r.ParseForm()
		Title = r.FormValue("postTitle")
		Content = r.FormValue("postContent")
		Categories = r.FormValue("postCategories")

		if Title != "" {
			Request := `INSERT INTO Posts (USERNAME, TITLE, CONTENT, CATEGORIES) VALUES (?, ?, ?, ?);`
			_, err := db.Exec(Request, username, Title, Content, Categories)
			if err != nil {
				log.Fatal(err)
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	newpost.ExecuteTemplate(w, "newPost.html", nil)
}

////////////////////////////////////////////////////////////////
/////////                    POST                      /////////
////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////
/////////                   COOKIES                    /////////
////////////////////////////////////////////////////////////////

func CheckCookies(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}

	var username string
	var expiration time.Time
	err = db.QueryRow("SELECT USERNAME, EXPIRATION FROM Sessions WHERE UUID = ?", cookie.Value).Scan(&username, &expiration)
	if err != nil || expiration.Before(time.Now()) {
		cookie := &http.Cookie{
			Name:   "session",
			Value:  "",
			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
		return ""
	}

	return username
}
