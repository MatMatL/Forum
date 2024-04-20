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

func main() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/login", Login)

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

func Index(w http.ResponseWriter, r *http.Request) {
	index.ExecuteTemplate(w, "index.html", nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
	login.ExecuteTemplate(w, "login.html", nil)

	//manage request
	if r.Method == "POST" {
		//sleep to let the time to the programe to run
		time.Sleep(69 * time.Millisecond)

		r.ParseForm()
		userName := r.FormValue("userName")
		fmt.Println("User name : ", userName)

		userPassword := r.FormValue("userPassword")
		fmt.Println("User password : ", userPassword)
	}
}
