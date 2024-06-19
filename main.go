package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
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
var newCategorie = template.Must(template.ParseFiles("newCategorie.html"))

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
	http.HandleFunc("/profil", Profil)
	http.HandleFunc("/newCategorie", NewCategorie)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

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
		CATEGORIES        TEXT    NOT NULL,
		IMAGEPATH         TEXT
	);
	`
	db.Exec(Posts)

	Categories := `
	CREATE TABLE IF NOT EXISTS Categories (
		ID INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		TITLE     TEXT NOT NULL UNIQUE,
		IMAGEPATH TEXT
	);
	`
	db.Exec(Categories)

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
	posts = FormatingPosts(posts)
	index.ExecuteTemplate(w, "index.html", posts)
}

type PostData struct {
	ID          int
	Username    string
	Title       string
	Content     string
	Categories  string
	ImagePath   string
	WithPicture bool
}

func getPosts() []PostData {
	rows, err := db.Query("SELECT ID, USERNAME, TITLE, CONTENT, CATEGORIES, IMAGEPATH FROM Posts")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var posts []PostData
	for rows.Next() {
		var post PostData
		if err := rows.Scan(&post.ID, &post.Username, &post.Title, &post.Content, &post.Categories, &post.ImagePath); err != nil {
			fmt.Println(err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return posts
}

func FormatingPosts(posts []PostData) []PostData {
	var formatedPosts []PostData
	if len(posts) >= 6 {
		for i := 0; i < 6; i++ {
			tempoFormated := FormatingPost(posts[i])
			formatedPosts = append(formatedPosts, tempoFormated)
		}
	} else {
		for i := 0; i < len(posts); i++ {
			tempoFormated := FormatingPost(posts[i])
			formatedPosts = append(formatedPosts, tempoFormated)
		}
	}

	return formatedPosts
}

func FormatingPost(posts PostData) PostData {
	var tempoFormated PostData
	if len(posts.Title) > 25 {
		tempoFormated.Title = posts.Title[:25] + "..."
	} else {
		tempoFormated.Title = posts.Title
	}
	if len(posts.Content) > 70 {
		tempoFormated.Content = posts.Content[:70] + "..."
	} else {
		tempoFormated.Content = posts.Content
	}
	if posts.ImagePath != "" {
		tempoFormated.WithPicture = true
		tempoFormated.ImagePath = posts.ImagePath
	} else {
		tempoFormated.WithPicture = false
	}
	tempoFormated.ID = posts.ID
	return tempoFormated
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

		username := loggingIn(UsernameInput, PasswordInput)

		if username != "" {
			sessionID, err := uuid.NewV4()
			if err != nil {
				log.Fatal(err)
			}
			expiration := time.Now().Add(24 * time.Hour)
			_, err = db.Exec("INSERT INTO Sessions (UUID, USERNAME, EXPIRATION) VALUES (?, ?, ?)", sessionID.String(), username, expiration)
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

func loggingIn(UsernameInput string, PasswordInput string) string {
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
			return UsernameInput
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
			return username
		}
	}
	return ""
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

type CategoriesStruct struct {
	Categorie string
}

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
		err := r.ParseMultipartForm(10 << 20) // Limite à 10 Mo
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		Title = r.FormValue("postTitle")
		Content = r.FormValue("postContent")
		Categories = r.FormValue("postCategories")
		file, handler, _ := r.FormFile("postImage")

		if file != nil {
			defer file.Close()
		}

		if Title != "" && Content != "" {
			if file != nil {

				filePath := "./uploads/" + handler.Filename
				out, err := os.Create(filePath)
				if err != nil {
					http.Error(w, "Unable to create the file for writing. Check your write access privilege", http.StatusInternalServerError)
					return
				}
				defer out.Close()

				_, err = io.Copy(out, file)
				if err != nil {
					http.Error(w, "Error occurred while saving the file", http.StatusInternalServerError)
					return
				}

				request := `INSERT INTO Posts (USERNAME, TITLE, CONTENT, CATEGORIES, IMAGEPATH) VALUES (?, ?, ?, ?, ?);`
				_, err = db.Exec(request, username, Title, Content, Categories, filePath)
				if err != nil {
					fmt.Println(err)
				} else {
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			} else {
				Request := `INSERT INTO Posts (USERNAME, TITLE, CONTENT, CATEGORIES, IMAGEPATH) VALUES (?, ?, ?, ?, ?);`
				_, err := db.Exec(Request, username, Title, Content, Categories, "")
				if err != nil {
					log.Fatal(err)
				}
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}
	}

	rows, err := db.Query("SELECT TITLE FROM Categories")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	var categories []CategoriesStruct
	for rows.Next() {
		var catego CategoriesStruct
		if err := rows.Scan(&catego.Categorie); err != nil {
			fmt.Println(err)
		}
		categories = append(categories, catego)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	if len(categories) == 0 {
		fmt.Println("No categories")
		http.Redirect(w, r, "/newCategorie", http.StatusSeeOther)
		return
	}

	newpost.ExecuteTemplate(w, "newPost.html", categories)
}

////////////////////////////////////////////////////////////////
/////////               NEWCATEGORIE                   /////////
////////////////////////////////////////////////////////////////

func NewCategorie(w http.ResponseWriter, r *http.Request) {
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

	if r.Method == "POST" {
		err := r.ParseMultipartForm(10 << 20) // Limite à 10 Mo
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		newCategorie := r.FormValue("newCategorie")
		file, handler, _ := r.FormFile("categoryImage")

		if file != nil {
			defer file.Close()
		}

		if newCategorie != "" {
			if file != nil {
				filePath := "./uploads/" + handler.Filename
				out, err := os.Create(filePath)
				if err != nil {
					http.Error(w, "Unable to create the file for writing. Check your write access privilege", http.StatusInternalServerError)
					return
				}
				defer out.Close()

				_, err = io.Copy(out, file)
				if err != nil {
					http.Error(w, "Error occurred while saving the file", http.StatusInternalServerError)
					return
				}

				request := `INSERT INTO Categories (TITLE, IMAGEPATH) VALUES (?, ?);`
				_, err = db.Exec(request, newCategorie, filePath)
				if err != nil {
					fmt.Println(err)
				} else {
					http.Redirect(w, r, "/newPost", http.StatusSeeOther)
					return
				}
			} else {
				request := `INSERT INTO Categories (TITLE) VALUES (?);`
				_, err = db.Exec(request, newCategorie)
				if err != nil {
					fmt.Println(err)
				} else {
					http.Redirect(w, r, "/newPost", http.StatusSeeOther)
					return
				}
			}
		}
	}

	newCategorie.ExecuteTemplate(w, "newCategorie.html", nil)
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
/////////                    PROFIL                    /////////
////////////////////////////////////////////////////////////////

func Profil(w http.ResponseWriter, r *http.Request) {
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
