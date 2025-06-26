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

	session, _ := store.Get(r, "session")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	adminName := ""
	if name, ok := session.Values["admin_name"].(string); ok {
		adminName = name
	}

	//Getting query parameters
	pageStr := r.URL.Query().Get("page")
	sortField := r.URL.Query().Get("field")
	sortOrder := r.URL.Query().Get("order")

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	switch sortField {
	case "username", "email", "id":
	default:
		sortField = "id"
	}

	switch sortOrder {
	case "asc", "desc":
	default:
		sortOrder = "desc"
	}

	offset := (page - 1) * limit

	query := fmt.Sprintf("SELECT id, username, email, mobile FROM New ORDER BY %s %s LIMIT ? OFFSET ?", sortField, sortOrder)

	rows, err := DB.Query(query, limit, offset)
	if err != nil {
		render.RenderTemplateWithData(w, "Home.html", EditPageData{
			Error: "Error fetching users",
		})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Mobile); err != nil {
			render.RenderTemplateWithData(w, "Home.html", EditPageData{
				Error:      "Error scanning user",
				Page:       page,
				SortField:  sortField,
				SortOrder:  sortOrder,
				TotalPages: 0,
			})
			return
		}
		users = append(users, u)
	}

	var total int
	if err := DB.QueryRow("SELECT COUNT(*) FROM New").Scan(&total); err != nil {
		render.RenderTemplateWithData(w, "Home.html", EditPageData{
			Error:      "Error counting users",
			Page:       page,
			SortField:  sortField,
			SortOrder:  sortOrder,
			TotalPages: 0,
		})
	}
	totalPages := (total + limit - 1) / limit

	flash := getFlashMessage(w, r)

	render.RenderTemplateWithData(w, "Home.html", EditPageData{
		Users:      users,
		Page:       page,
		TotalPages: totalPages,
		Error:      flash,
		Title:      "User Listing",
		SortField:  sortField,
		SortOrder:  sortOrder,
		AdminName:  adminName,
	})
}
