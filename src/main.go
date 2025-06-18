package main

import (
	"fmt"
	"mysqliteapp/handler"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler.RegisterHandler)
	http.HandleFunc("/home", handler.HomeHandler)
	http.HandleFunc("/edit", handler.EditHandler)
	http.HandleFunc("/update", handler.UpdateHandler)
	http.HandleFunc("/delete", handler.DeleteHandler)

	fmt.Println("Application running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
