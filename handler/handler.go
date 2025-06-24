package handler

import (
	"mysqliteapp/render"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	const limit = 5
	page := 1
	pageStr := r.URL.Query().Get("page")
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	offset := (page - 1) * limit

	rows, err := DB.Query("SELECT id, username, email, mobile FROM New ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		render.RenderTemplateWithData(w, "Home.html", EditPageData{Error: "Error fetching users"})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Mobile); err != nil {
			render.RenderTemplateWithData(w, "Home.html", EditPageData{Error: "Error scanning user"})
			return
		}
		users = append(users, u)
	}

	var total int
	DB.QueryRow("SELECT COUNT(*) FROM New").Scan(&total)
	totalPages := (total + limit - 1) / limit

	flash := getFlashMessage(w, r)

	render.RenderTemplateWithData(w, "Home.html", EditPageData{
		Users:      users,
		Page:       page,
		TotalPages: totalPages,
		Error:      flash,
	})
}
