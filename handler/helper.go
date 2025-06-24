package handler

import (
	"net/http"
)

func setFlashMessage(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "flash",
		Value: message,
		Path:  "/",
	})
}

func getFlashMessage(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("flash")
	if err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "flash",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	return cookie.Value
}

func getCountriesFromDB() ([]string, error) {
	rows, err := DB.Query("SELECT name FROM Countries")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		countries = append(countries, name)
	}
	return countries, nil
}
