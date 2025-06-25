package handler

import (
	"fmt"
	"mysqliteapp/render"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	const limit = 5
	page := 1
	sortOrder := "DESC"

	pageStr := r.URL.Query().Get("page")
	sort := r.URL.Query().Get("sort")

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	if sort == "asc" {
		sortOrder = "ASC"
	} else {
		sortOrder = "DESC"
	}

	offset := (page - 1) * limit

	query := fmt.Sprintf("SELECT id, username, email, mobile FROM New ORDER BY id %s LIMIT ? OFFSET ?", sortOrder)

	rows, err := DB.Query(query, limit, offset)
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
		Title:      "User Listing",
		Sort:       sort,
	})
}
