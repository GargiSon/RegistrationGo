package render

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"seq": func(start, end int) []int {
		s := make([]int, 0, end-start+1)
		for i := start; i <= end; i++ {
			s = append(s, i)
		}
		return s
	},
}

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

	t, err := template.New("base.html").Funcs(funcMap).ParseFiles(tmplFiles...)
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
