package render

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

func RenderTemplate(w http.ResponseWriter, temp string) {
	t, err := template.ParseFiles(filepath.Join("templates", temp))
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		fmt.Println("RenderTemplate error: ", err)
		return
	}
	t.Execute(w, nil)
}

func RenderTemplateWithData(w http.ResponseWriter, temp string, data any) {
	t, err := template.ParseFiles(filepath.Join("templates", temp))
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	t.Execute(w, data)
}
