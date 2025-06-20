package main

import (
	"fmt"
	"mysqliteapp/handler"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler.LoginHandler)
	http.HandleFunc("/login", handler.LoginHandler)
	http.HandleFunc("/logout", handler.LogoutHandler)
	http.HandleFunc("/forgot-password", handler.ForgotPasswordHandler)

	http.HandleFunc("/register", handler.RegisterHandler)
	http.HandleFunc("/home", handler.HomeHandler)
	http.HandleFunc("/edit", handler.EditHandler)
	http.HandleFunc("/update", handler.UpdateHandler)
	http.HandleFunc("/delete", handler.DeleteHandler)

	fmt.Println("Application running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
