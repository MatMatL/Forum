package main

import (
	"fmt"
	"html/template"
	"net/http"
)

const port = ":8080"

var index = template.Must(template.ParseFiles("Index.html"))

func main() {
	http.HandleFunc("/", Index)

	fmt.Println("//localhost:8080")
	http.ListenAndServe(port, nil)
}

func Index(w http.ResponseWriter, r *http.Request) {
	index.ExecuteTemplate(w, "Index.html", nil)
}
