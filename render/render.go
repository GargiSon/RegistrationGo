package render

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

func RenderTemplate(w http.ResponseWriter, temp string) {
	RenderTemplateWithData(w, temp, nil)
}

func RenderTemplateWithData(w http.ResponseWriter, temp string, data any) {
	tmplFiles := []string{
		filepath.Join("templates", "base.html"),
		filepath.Join("templates", "header.html"),
		filepath.Join("templates", "footer.html"),
		filepath.Join("templates", temp), // e.g. home.html or about.html
	}

	t, err := template.ParseFiles(tmplFiles...)
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		fmt.Println("RenderTemplateWithData error:", err)
		return
	}

	// Always execute the base layout
	err = t.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		fmt.Println("Template execution error:", err)
	}
}
