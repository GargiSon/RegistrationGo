package handler

import (
	"database/sql"
	"mysqliteapp/render"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var db *sql.DB

type User struct {
	ID       int
	Username string
	Email    string
	Mobile   string
	Address  string
	Gender   string
	Sports   string
	DOB      string
	Country  string
}

type EditPageData struct {
	User      User
	Countries []string
	SportsMap map[string]bool
}

func init() {
	var err error
	db, err = sql.Open("sqlite", "./New.db") //driver name, datasource name
	if err != nil {
		panic(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS New (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		mobile TEXT NOT NULL,
		address TEXT NOT NULL,
		gender TEXT NOT NULL,
		sports TEXT NOT NULL,
		dob TEXT NOT NULL,
		country TEXT NOT NULL
	);`

	createCountryTable := `
	CREATE TABLE IF NOT EXISTS Countries(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	)`

	if _, err = db.Exec(createTable); err != nil { //returns res, err but res not used here
		panic(err) //this stops the program immediately, when an error occurs
	}

	if _, err = db.Exec(createCountryTable); err != nil {
		panic(err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM Countries").Scan(&count); err != nil {
		panic(err)
	}

	if count == 0 {
		_, err = db.Exec(`INSERT INTO Countries (name) VALUES ('INDIA'),('AFGHANISTHAN'),('FRANCE')`)
		if err != nil {
			panic(err)
		}
	}
}

func getCountriesFromDB() ([]string, error) {
	rows, err := db.Query("SELECT name FROM Countries")
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirm := r.FormValue("confirm")
		email := r.FormValue("email")
		mobile := r.FormValue("mobile")
		address := r.FormValue("address")
		gender := r.FormValue("gender")
		sports := r.Form["sports"] //Slice
		dob := r.FormValue("dob")
		country := r.FormValue("country")

		// sports := "" //string
		// if len(sportslice) > 0 {
		// 	sports = strings.Join(sportslice, ",") //Join differet values using a comma
		// }
		joinedSports := strings.Join(sports, ",")

		if password != confirm {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //securely hash users password, default cost value is 10, but values like12, 14 are more secure but slow, and hashed password is in slice byte format
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO New(username, password, email, mobile, address, gender, sports, dob, country) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", username, hashed, email, mobile, address, gender, joinedSports, dob, country)

		if err != nil {
			http.Error(w, "Database error"+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	countries, err := getCountriesFromDB()
	if err != nil {
		http.Error(w, "Error fetching countries: "+err.Error(), http.StatusInternalServerError)
		return
	}
	render.RenderTemplateWithData(w, "Registration.html", map[string]any{"Countries": countries})
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, username, email, mobile FROM New")
	if err != nil {
		http.Error(w, "Error fetching users: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Mobile); err != nil {
			http.Error(w, "Error scanning user: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}
	render.RenderTemplateWithData(w, "Home.html", map[string]any{"Users": users})
}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT id, username, email, mobile, address, gender, sports, dob, country FROM New WHERE id = ?", id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Mobile, &user.Address, &user.Gender, &user.Sports, &user.DOB, &user.Country)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
	countries, err := getCountriesFromDB()
	if err != nil {
		http.Error(w, "Error fetching countries: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert sports (comma-separated string) to map for checkbox logic
	sportsMap := make(map[string]bool)
	for _, sport := range strings.Split(user.Sports, ",") {
		sport = strings.TrimSpace(sport)
		if sport != "" {
			sportsMap[sport] = true
		}
	}

	render.RenderTemplateWithData(w, "Edit.html", map[string]any{
		"User":      user,
		"Countries": countries,
		"SportsMap": sportsMap,
	})
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		username := r.FormValue("username")
		mobile := r.FormValue("mobile")
		address := r.FormValue("address")
		gender := r.FormValue("gender")
		dob := r.FormValue("dob")
		country := r.FormValue("country")
		sportsSlice := r.Form["sports"]

		sports := strings.Join(sportsSlice, ",")

		_, err := db.Exec(`UPDATE New SET username=?, mobile=?, address=?, gender=?, sports=?, dob=?, country=? WHERE id=?`,
			username, mobile, address, gender, sports, dob, country, id)
		if err != nil {
			http.Error(w, "Updation Failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		_, err := db.Exec("DELETE FROM New WHERE id = ?", id)
		if err != nil {
			http.Error(w, "Error deleting user: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}
