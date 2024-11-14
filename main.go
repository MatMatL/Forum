package main

//désolé tous dans le main pas le temps de split mais c'est structuré avec les gros bandeaux
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

// port definition
const port = ":8080"

// templates definition
var index = template.Must(template.ParseFiles("index.html"))
var login = template.Must(template.ParseFiles("login.html"))
var register = template.Must(template.ParseFiles("register.html"))
var newpost = template.Must(template.ParseFiles("newPost.html"))
var post = template.Must(template.ParseFiles("post.html"))
var newCategorie = template.Must(template.ParseFiles("newCategorie.html"))
var categorie = template.Must(template.ParseFiles("categorie.html"))
var user = template.Must(template.ParseFiles("user.html"))
var profil = template.Must(template.ParseFiles("profil.html"))
var categories = template.Must(template.ParseFiles("categories.html"))

// db link
var db *sql.DB

func main() {
	//db oppening
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//tables initialisations
	InitTables()

	//handlers declaration
	http.HandleFunc("/", Index)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/logout", Logout)
	http.HandleFunc("/newPost", NewPost)
	http.HandleFunc("/post", Post)
	http.HandleFunc("/profil", Profil)
	http.HandleFunc("/newCategorie", NewCategorie)
	http.HandleFunc("/categorie", Categorie)
	http.HandleFunc("/user", User)
	http.HandleFunc("/categories", Categories)
	http.HandleFunc("/deletePost", DeletePost)
	http.HandleFunc("/deleteCategorie", DeleteCategorie)

	//handle wokspace files (css and pictures)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	fmt.Println("//localhost:8080")
	http.ListenAndServe(port, nil)
}

////////////////////////////////////////////////////////////////
/////////                  DATABASE                    /////////
////////////////////////////////////////////////////////////////

// function that creates tables if they do not already exist
func InitTables() {
	Register := `
	CREATE TABLE IF NOT EXISTS Register (
		ID INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		EMAIL          TEXT    NOT NULL UNIQUE,
		USERNAME       TEXT    NOT NULL UNIQUE,
		PASSWORD       TEXT    NOT NULL,
		IMAGEPATH      TEXT
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
		UUID       TEXT PRIMARY KEY UNIQUE,
		USERNAME   TEXT     NOT NULL,
		EXPIRATION DATETIME NOT NULL
	);
	`
	db.Exec(Sessions)

	Comments := `
	CREATE TABLE IF NOT EXISTS Comments (
		ID INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		POSTID     INT      NOT NULL,
		USERNAME   TEXT     NOT NULL,
		CONTENT    TEXT     NOT NULL UNIQUE
	);
	`
	db.Exec(Comments)
}

////////////////////////////////////////////////////////////////
/////////                   INDEX                      /////////
////////////////////////////////////////////////////////////////

// struct with data to be send in the index
type IndexData struct {
	Posts       []PostData
	Categories  []CategorieData
	Users       []UserData
	IsNotLogged bool
}

// index handler
func Index(w http.ResponseWriter, r *http.Request) {
	var data IndexData
	data.Posts = FormatingPosts()
	data.Categories = FormatingCategories()
	data.Users = FormatingUsers()

	_, err := r.Cookie("session")
	if err != nil {
		data.IsNotLogged = true
	} else {
		data.IsNotLogged = false
	}

	index.ExecuteTemplate(w, "index.html", data)
}

// function to limit the number of post in index post section (max 7 posts)
func FormatingPosts() []PostData {
	data := GetPosts()
	var formatedPosts []PostData
	if len(data) >= 7 {
		for i := 0; i < 7; i++ {
			tempoFormated := FormatingPost(data[i])
			formatedPosts = append(formatedPosts, tempoFormated)
		}
	} else {
		for i := 0; i < len(data); i++ {
			tempoFormated := FormatingPost(data[i])
			formatedPosts = append(formatedPosts, tempoFormated)
		}
	}

	return formatedPosts
}

// function that limit number of characters in index post section (max 25 char for title and max 100 char for content)
func FormatingPost(posts PostData) PostData {
	var tempoFormated PostData
	if len(posts.Title) > 25 {
		tempoFormated.Title = posts.Title[:25] + "..."
	} else {
		tempoFormated.Title = posts.Title
	}
	if len(posts.Content) > 100 {
		tempoFormated.Content = posts.Content[:100] + "..."
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

// function to limit the number of categories in index categories section (max 4 categories)
func FormatingCategories() []CategorieData {
	data := GetCategories()
	var formatedCategories []CategorieData
	if len(data) >= 4 {
		for i := 0; i < 4; i++ {
			formatedCategories = append(formatedCategories, data[i])
		}
	} else {
		for i := 0; i < len(data); i++ {
			formatedCategories = append(formatedCategories, data[i])
		}
	}

	return formatedCategories
}

// function to limit the number of ursers in index ursers section (max 4 ursers)
func FormatingUsers() []UserData {
	Users := GetUser()
	var formatedUsers []UserData
	if len(Users) >= 4 {
		for i := 0; i < 4; i++ {
			formatedUsers = append(formatedUsers, Users[i])
		}
	} else {
		for i := 0; i < len(Users); i++ {
			formatedUsers = append(formatedUsers, Users[i])
		}
	}

	return formatedUsers
}

////////////////////////////////////////////////////////////////
/////////                   LOGIN                      /////////
////////////////////////////////////////////////////////////////

// login handler
func Login(w http.ResponseWriter, r *http.Request) {
	//code to check if the user is connected or not (if yes it is redirect to main menu)
	username := CheckCookies(w, r)
	if username != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	//check if page send data
	if r.Method == "POST" {
		//parse data
		r.ParseForm()
		UsernameInput := r.FormValue("userName")
		PasswordInput := r.FormValue("userPassword")

		//check if input data is correct
		username := loggingIn(UsernameInput, PasswordInput)
		if username != "" { //if it is correct then ->
			//create session uuid using package uuid ("github.com/gofrs/uuid")
			sessionID, err := uuid.NewV4()
			if err != nil {
				log.Fatal(err)
			}
			//define cookie living time
			expiration := time.Now().Add(24 * time.Hour)
			//insert into data base
			_, err = db.Exec("INSERT INTO Sessions (UUID, USERNAME, EXPIRATION) VALUES (?, ?, ?)", sessionID.String(), username, expiration)
			if err != nil {
				log.Fatal(err)
			}

			//create cookie
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

// function that check if the user inputs are correct with the 2 ways of connection (with email or username)
func loggingIn(UsernameInput string, PasswordInput string) string {
	//first check using username
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
			//if yes then early return
			return UsernameInput
		}
	}

	//check with email
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
	//if not found or incorrect then nothing
	return ""
}

////////////////////////////////////////////////////////////////
/////////                  REGISTER                    /////////
////////////////////////////////////////////////////////////////

// structure to be send to transfer errors
type loginErrors struct {
	WrongEmail    bool
	WrongUsername bool
	WrongPassword bool
}

// register handler
func Register(w http.ResponseWriter, r *http.Request) {
	//check if user is already logged in
	username := CheckCookies(w, r)
	if username != "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	//error container
	var currentErrors = loginErrors{}

	if r.Method == "POST" {
		//data collection
		r.ParseForm()
		EmailInput := r.FormValue("newEmail")
		UsernameInput := r.FormValue("newUserName")
		PasswordInput := r.FormValue("newUserPassword")

		//check if email can be taken
		if !ValidEmail(EmailInput) || AlreadyTakenEmail(EmailInput) {
			currentErrors.WrongEmail = true
		} else if AlreadyTakenUsername(UsernameInput) {
			currentErrors.WrongUsername = true
		} else {
			//if all ok then isert into db
			Request := `INSERT INTO Register (EMAIL,USERNAME,PASSWORD) VALUES (?, ?, ?);`
			//hach password
			HachedPassword, errHash := bcrypt.GenerateFromPassword([]byte(PasswordInput), 10)
			if errHash != nil {
				log.Fatal(errHash)
			}
			_, err := db.Exec(Request, EmailInput, UsernameInput, string(HachedPassword))
			if err != nil {
				log.Fatal(err)
			}
			//redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	register.ExecuteTemplate(w, "register.html", currentErrors)
}

// check if an email is valid (only based on the presence of a '@' xD)
func ValidEmail(email string) bool {
	for _, i := range email {
		if i == '@' {
			return true
		}
	}
	return false
}

// check with a query if the email is already used
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

// check with a query if the username is already used
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

// compare the hach of the user input with the hach in the db matching the username
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

// small hadler to log out of the session
func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		//delet from the db
		db.Exec("DELETE FROM Sessions WHERE UUID = ?", cookie.Value)
		cookie := &http.Cookie{
			Name:  "session",
			Value: "",
			//kill the cookie
			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

////////////////////////////////////////////////////////////////
/////////                  NEW POST                    /////////
////////////////////////////////////////////////////////////////

// struct to send categories to html
type CategoriesStruct struct {
	Categorie string
}

// new post handler
func NewPost(w http.ResponseWriter, r *http.Request) {
	//check cookie
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
		err := r.ParseMultipartForm(10 << 20) // 10 Mo limit
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		//parse form to get data
		Title = r.FormValue("postTitle")
		Content = r.FormValue("postContent")
		Categories = r.FormValue("postCategories")
		file, handler, _ := r.FormFile("postImage")

		//because picture is not an obligation
		if file != nil {
			defer file.Close()
		}

		//only if there is a title and content to post
		if Title != "" && Content != "" {
			if file != nil {

				//create the picture file
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
				//insert into db
				request := `INSERT INTO Posts (USERNAME, TITLE, CONTENT, CATEGORIES, IMAGEPATH) VALUES (?, ?, ?, ?, ?);`
				_, err = db.Exec(request, username, Title, Content, Categories, filePath)
				if err != nil {
					fmt.Println(err)
				} else {
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			} else {
				//if there is no picture (way easier)
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

	//get categories for the categorie selection
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

// handler to create categories
func NewCategorie(w http.ResponseWriter, r *http.Request) {
	//check cookie
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
		err := r.ParseMultipartForm(10 << 20) // 10 Mo limit
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		//get data
		newCategorie := r.FormValue("newCategorie")
		file, handler, _ := r.FormFile("categoryImage")

		if file != nil {
			defer file.Close()
		}

		//same logic as for new posts
		if newCategorie != "" {
			if file != nil {
				//create picture file
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

				//insert into db
				request := `INSERT INTO Categories (TITLE, IMAGEPATH) VALUES (?, ?);`
				_, err = db.Exec(request, newCategorie, filePath)
				if err != nil {
					fmt.Println(err)
				} else {
					http.Redirect(w, r, "/newPost", http.StatusSeeOther)
					return
				}
			} else {
				//if no picture
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

// a struct to represent a post in a single struct
type PostData struct {
	ID          int
	Username    string
	Title       string
	Content     string
	ImagePath   string
	WithPicture bool
	Categories  string
	Comments    []CommentData
	IsAdmin     bool
}

type CommentData struct {
	ID       int
	PostID   int
	Username string
	Content  string
}

// post handler
func Post(w http.ResponseWriter, r *http.Request) {
	//to get the ID in the url
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

	//get the post matching the id in the url
	postData, err := getPostByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	username := CheckCookies(w, r)

	if r.Method == "POST" && username != "" {
		//data collection
		r.ParseForm()
		comment := r.FormValue("comment")
		if comment != "" {
			//insert into data base
			db.Exec("INSERT INTO Comments (POSTID, USERNAME, CONTENT) VALUES (?, ?, ?)", id, username, comment)
		}
		comment = ""
	}

	postData.Comments = GetCommentsPostsByID(id)

	if username == "admin" {
		postData.IsAdmin = true
	} else {
		postData.IsAdmin = false
	}

	post.ExecuteTemplate(w, "post.html", postData)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	username := CheckCookies(w, r)

	//to get the ID in the url
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

	if username != "admin" {
		path := "/post?id=" + string(id)
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}

	DeleteAPost(id)

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func DeleteAPost(id int) {
	DeleteAComment(id)
	db.Exec("DELETE FROM Posts WHERE ID = ?", id)
}

func DeleteAComment(id int) {
	db.Exec("DELETE FROM Comments WHERE POSTID = ?", id)
}

// get all the post from the db and return an array of them (using postdata structure)
func GetPosts() []PostData {
	//query to get them all
	rows, err := db.Query("SELECT ID, USERNAME, TITLE, CONTENT, CATEGORIES, IMAGEPATH FROM Posts")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	//define the container
	var posts []PostData
	//fill the array row by row
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

	//return the filled array
	return posts
}

// get a post matching it's id
func getPostByID(id int) (PostData, error) {
	var post PostData
	row := db.QueryRow("SELECT ID, USERNAME, TITLE, CONTENT, CATEGORIES FROM Posts WHERE ID = ?", id)
	err := row.Scan(&post.ID, &post.Username, &post.Title, &post.Content, &post.Categories)
	if err != nil {
		return post, err
	}
	return post, nil
}

// get all ths posts made by a user (from his username)
func GetPostsByUsername(username string) []PostData {
	rows, err := db.Query("SELECT ID, USERNAME, TITLE, CONTENT, CATEGORIES, IMAGEPATH FROM Posts WHERE username = ?", username)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	//create the array and fill it row by row
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

	//return filled array
	return posts
}

// get all the posts of a given categorie and return them in an PostData array
func GetPostsByCategory(category string) []PostData {
	rows, err := db.Query("SELECT ID, TITLE, CONTENT, IMAGEPATH FROM Posts WHERE CATEGORIES = ?", category)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	//create the array and fill it row by row
	var posts []PostData
	for rows.Next() {
		var post PostData
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImagePath); err != nil {
			fmt.Println(err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	//return filled array
	return posts
}

func GetCommentsPostsByID(id int) []CommentData {
	//query to get all comments matching the post id
	rows, err := db.Query("SELECT ID, POSTID, USERNAME, CONTENT FROM Comments WHERE POSTID = ?", id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	//define the container
	var comments []CommentData
	//fill the array row by row
	for rows.Next() {
		var comment CommentData
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.Username, &comment.Content); err != nil {
			fmt.Println(err)
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	//return the filled array
	return comments
}

////////////////////////////////////////////////////////////////
/////////                 Categorie                    /////////
////////////////////////////////////////////////////////////////

// struct that contains all the data of a categorie
type CategorieData struct {
	Name      string
	ImagePath string
	ID        int
	Posts     []PostData
	IsAdmin   bool
}

// categories page handler
func Categories(w http.ResponseWriter, r *http.Request) {
	categorieData := GetCategories()

	categories.ExecuteTemplate(w, "categories.html", categorieData)
}

// single category page handler
func Categorie(w http.ResponseWriter, r *http.Request) {
	//get the id of the category in the url
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

	//get the data of this category
	categorieData := getCategorieByID(id)

	//and get the posts made on this categorie
	categorieData.Posts = GetPostsByCategory(categorieData.Name)

	username := CheckCookies(w, r)

	if username == "admin" {
		categorieData.IsAdmin = true
	} else {
		categorieData.IsAdmin = false
	}

	categorie.ExecuteTemplate(w, "categorie.html", categorieData)
}

// function that get all the categories and return them in a 'CategorieData' array
func GetCategories() []CategorieData {
	rows, err := db.Query("SELECT ID, TITLE, IMAGEPATH FROM Categories")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	var categories []CategorieData
	for rows.Next() {
		var categorie CategorieData
		rows.Scan(&categorie.ID, &categorie.Name, &categorie.ImagePath)
		categories = append(categories, categorie)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return categories
}

// function that return category data from it's ID
func getCategorieByID(id int) CategorieData {
	var categorie CategorieData
	row := db.QueryRow("SELECT ID, TITLE, IMAGEPATH FROM Categories WHERE ID = ?", id)
	row.Scan(&categorie.ID, &categorie.Name, &categorie.ImagePath)

	return categorie
}

// get category name of the category and call delete posts
func DeleteACategory(id int) {
	row := db.QueryRow("SELECT TITLE FROM Categories WHERE ID = ?", id)

	var title string
	row.Scan(&title)

	DeleteAPostsFromCategory(title)
	db.Exec("DELETE FROM Categories WHERE ID = ?", id)
}

func DeleteAPostsFromCategory(category string) {
	rows, _ := db.Query("SELECT ID FROM Posts WHERE category = ?", category)
	var categories []string
	for rows.Next() {
		var category string
		rows.Scan(&category)
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	for i := 0; i < len(categories); i++ {
		id, _ := strconv.Atoi(categories[i])
		DeleteAPost(id)
	}
}

func DeleteCategorie(w http.ResponseWriter, r *http.Request) {
	username := CheckCookies(w, r)

	//to get the ID in the url
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

	if username != "admin" {
		path := "/post?id=" + string(id)
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}

	DeleteACategory(id)

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

////////////////////////////////////////////////////////////////
/////////                    PROFIL                    /////////
////////////////////////////////////////////////////////////////

// structure that contains all user data (including it's posts)
type UserData struct {
	ID        int
	Username  string
	ImagePath string
	Posts     []PostData
}

// function that get and return all the users in a 'UserData' array
func GetUser() []UserData {
	rows, err := db.Query("SELECT ID, USERNAME, IMAGEPATH FROM Register")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var users []UserData
	for rows.Next() {
		var user UserData
		rows.Scan(&user.ID, &user.Username, &user.ImagePath)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}

	return users
}

// get user data from it's ID
func GetUserByID(id int) UserData {
	var user UserData
	row := db.QueryRow("SELECT ID, USERNAME, IMAGEPATH FROM Register WHERE ID = ?", id)
	row.Scan(&user.ID, &user.Username, &user.ImagePath)

	return user
}

// get user data from it's username
func GetUserByUsername(username string) UserData {
	var user UserData
	row := db.QueryRow("SELECT ID, USERNAME, IMAGEPATH FROM Register WHERE USERNAME = ?", username)
	row.Scan(&user.ID, &user.Username, &user.ImagePath)

	return user
}

// profil handler
func Profil(w http.ResponseWriter, r *http.Request) {
	//check if the user is connected
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

	//get user data from it's name (that we got in the cookie)
	userData := GetUserByUsername(username)

	//get it's posts
	userData.Posts = GetPostsByUsername(userData.Username)

	//part for updating data
	if r.Method == "POST" {
		//parse forms
		err := r.ParseMultipartForm(10 << 20) // Limite à 10 Mo
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		//get new email if there is one
		newEmail := r.FormValue("newEmail")
		if ValidEmail(newEmail) && !AlreadyTakenEmail(newEmail) && newEmail != "" {
			query := `UPDATE Register SET EMAIL = ? WHERE ID = ?`
			db.Exec(query, newEmail, userData.ID)
		}

		//get new username if there is one
		newUserName := r.FormValue("newUserName")
		if !AlreadyTakenUsername(newUserName) && newUserName != "" {
			query := `UPDATE Register SET USERNAME = ? WHERE ID = ?`
			db.Exec(query, newUserName, userData.ID)
		}

		//get new password if there is one
		newUserPassword := r.FormValue("newUserPassword")
		if newUserPassword != "" {
			query := `UPDATE Register SET PASSWORD = ? WHERE ID = ?`
			db.Exec(query, newUserPassword, userData.ID)
		}

		//get new profil picture if there is one
		file, handler, _ := r.FormFile("postImage")
		if file != nil {
			//create the file
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
			//update the db
			query := `UPDATE Register SET IMAGEPATH = ? WHERE ID = ?`
			db.Exec(query, filePath, userData.ID)
		}
	}

	profil.ExecuteTemplate(w, "profil.html", userData)
}

// user handler
func User(w http.ResponseWriter, r *http.Request) {
	//get the id in the url
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

	//get data of the user
	userData := GetUserByID(id)

	//get posts of this user
	userData.Posts = GetPostsByUsername(userData.Username)

	user.ExecuteTemplate(w, "user.html", userData)
}

////////////////////////////////////////////////////////////////
/////////                   COOKIES                    /////////
////////////////////////////////////////////////////////////////

// check if the user has a cookie and if his session is still valid, if yes it returns it's username
func CheckCookies(w http.ResponseWriter, r *http.Request) string {
	//get cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}

	//check db data
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
