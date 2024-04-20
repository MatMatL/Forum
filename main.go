package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

const port = ":8080"

var index = template.Must(template.ParseFiles("index.html"))
var login = template.Must(template.ParseFiles("login.html"))
var register = template.Must(template.ParseFiles("register.html"))

func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)

	fmt.Println("//localhost:8080")
	http.ListenAndServe(port, nil)
}

type User struct {
	name     string
	password string
}

type Session struct {
	user        User
	isConnected bool
}

var Users []User
var CurrentSession = Session{}

func Index(w http.ResponseWriter, r *http.Request) {
	if !CurrentSession.isConnected {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	index.ExecuteTemplate(w, "index.html", nil)
}

type loginErrors struct {
	WrongUsername bool
	WrongPassword bool
}

func Login(w http.ResponseWriter, r *http.Request) {
	var moi = User{"Mathieu", "1234"}
	Users = make([]User, 1)
	Users[0] = moi
	var currentErrors = loginErrors{}

	var UsernameInput string
	var PasswordInput string

	if r.Method == "POST" {
		time.Sleep(69 * time.Millisecond)

		r.ParseForm()
		UsernameInput = r.FormValue("userName")
		fmt.Println("User name : ", CurrentSession.user.name)

		PasswordInput = r.FormValue("userPassword")
		fmt.Println("User password : ", CurrentSession.user.password)
	}

	time.Sleep(200 * time.Millisecond)

	for _, currentUser := range Users {
		if currentUser.name == UsernameInput {
			currentErrors.WrongUsername = false
			fmt.Println("found username")
			if currentUser.password == PasswordInput {
				fmt.Println("found password")
				CurrentSession.isConnected = true
				CurrentSession.user = currentUser
			} else {
				currentErrors.WrongPassword = true
			}
			break
		}
	}

	if !CurrentSession.isConnected && !currentErrors.WrongPassword && UsernameInput != "" {
		fmt.Println("user not found")
		currentErrors.WrongUsername = true
	}

	if CurrentSession.isConnected {
		fmt.Println("redirection")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	login.ExecuteTemplate(w, "login.html", currentErrors)

}

func Register(w http.ResponseWriter, r *http.Request) {
	if CurrentSession.isConnected {
		fmt.Println("redirection")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}
