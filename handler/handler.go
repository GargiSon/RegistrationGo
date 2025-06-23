package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"mysqliteapp/render"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	User       User
	Countries  []string
	SportsMap  map[string]bool
	Error      string
	Title      string
	Users      []User
	Page       int
	TotalPages int
	Info       string
	Email      string
	Ts         string
	Token      string
}

func init() {
	var err error
	db, err = sql.Open("sqlite", "./New.db") //driver name, datasource name
	if err != nil {
		log.Println("Error")
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

	createAdminTable := `
	CREATE TABLE IF NOT EXISTS Admin(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	)`

	if _, err = db.Exec(createTable); err != nil { //returns res, err but res not used here
		log.Println("Error") //this stops the program immediately, when an error occurs
	}

	if _, err = db.Exec(createCountryTable); err != nil {
		log.Println("Error")
	}

	if _, err = db.Exec(createAdminTable); err != nil {
		log.Println("Error")
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM Countries").Scan(&count); err != nil {
		log.Println("Error")
	}

	if count == 0 {
		_, err = db.Exec(`INSERT INTO Countries (name) VALUES ('INDIA'),('AFGHANISTHAN'),('FRANCE')`)
		if err != nil {
			log.Println("Error")
		}
	}

	var adminCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM Admin").Scan(&adminCount)
	if adminCount == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("admin1001"), bcrypt.DefaultCost)
		_, _ = db.Exec("INSERT INTO Admin(email, password) VALUES (?, ?)", "farziemail@yopmail.com", hashed)
	}
}

const resetsecret = "hubjinkom"

const (
	smtpHost     = "smtp.yopmail.com"
	smtpPort     = "587"
	smtpEmail    = "farziemail@yopmail.com"
	smtpPassword = "admin1001"
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

func sendResetEmail(toEmail, resetLink string) error {
	auth := smtp.PlainAuth("", smtpEmail, smtpPassword, smtpHost)

	subject := "Subject: Reset Your Admin Password\n"
	body := fmt.Sprintf("To reset your password, click the link below:\n\n%s", resetLink)

	msg := []byte(subject + "\n" + body)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpEmail, []string{toEmail}, msg)
	return err
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		render.RenderTemplateWithData(w, "Login.html", EditPageData{
			Error: "",
		})
		return
	}
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var storedHash string
		err := db.QueryRow("SELECT password FROM Admin WHERE email = ?", email).Scan(&storedHash)
		if err != nil {
			setFlashMessage(w, "Invalid email or password")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) != nil {
			render.RenderTemplateWithData(w, "Login.html", EditPageData{
				Error: "Invalid email or password",
			})
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "admin_logged_in",
			Value: "true",
			Path:  "/",
		})
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "admin_logged_in",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		render.RenderTemplateWithData(w, "Forgot.html", EditPageData{
			Error: "",
			Info:  getFlashMessage(w, r),
		})
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM Admin WHERE email = ?)", email).Scan(&exists)
		if err != nil || !exists {
			setFlashMessage(w, "Email not Found")
			http.Redirect(w, r, "/forgot", http.StatusSeeOther)
			return
		}

		ts := fmt.Sprint(time.Now().Unix())
		hash := sha256.Sum256([]byte(email + ts + resetsecret))
		token := hex.EncodeToString(hash[:])

		link := fmt.Sprintf("http://localhost:8080/reset?email=%s&ts=%s&token=%s", url.QueryEscape(email), ts, token)
		if err := sendResetEmail(email, link); err != nil {
			log.Println("Failed to send email:", err)
			setFlashMessage(w, "Failed to send reset link. Try again.")
		} else {
			setFlashMessage(w, "Reset link sent! Check your email.")
		}
		http.Redirect(w, r, "/forgot", http.StatusSeeOther)
	}
	render.RenderTemplateWithData(w, "forgot.html", nil)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	countries, err := getCountriesFromDB()
	if err != nil {
		render.RenderTemplateWithData(w, "Registration.html", EditPageData{
			Error: "Error fetching countries: " + err.Error(),
		})
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		confirm := r.FormValue("confirm")
		email := r.FormValue("email")
		mobile := r.FormValue("mobile")
		address := r.FormValue("address")
		gender := r.FormValue("gender")
		sports := r.Form["sports"] //Slice
		dobStr := r.FormValue("dob")
		country := r.FormValue("country")

		joinedSports := strings.Join(sports, ",")

		user := User{
			Username: username,
			Email:    email,
			Mobile:   mobile,
			Address:  address,
			Gender:   gender,
			Sports:   joinedSports,
			DOB:      dobStr,
			Country:  country,
		}

		sportsMap := make(map[string]bool)
		for _, s := range sports {
			sportsMap[s] = true
		}

		//Same password
		if password != confirm {
			render.RenderTemplateWithData(w, "Registration.html", EditPageData{
				Error:     "Passwords do not match",
				Countries: countries,
				User:      user,
				SportsMap: sportsMap,
			})
			return
		}

		//First changing dob from string to time format, then checking
		dob, err := time.Parse("2006-01-02", dobStr)
		if err != nil || dob.After(time.Now()) {
			render.RenderTemplateWithData(w, "Registration.html", EditPageData{
				Error:     "Invalid or future DOB",
				Countries: countries,
				User:      user,
				SportsMap: sportsMap,
			})
			return
		}
		//Mobile number constraint
		match, err := regexp.MatchString(`^(\+\d{1,3})?\d{10}$`, mobile)
		if err != nil || !match {
			render.RenderTemplateWithData(w, "Registration.html", EditPageData{
				Error:     "Invalid mobile number format",
				Countries: countries,
				User:      user,
				SportsMap: sportsMap,
			})
			return
		}

		//Hash password
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //securely hash users password, default cost value is 10, but values like12, 14 are more secure but slow, and hashed password is in slice byte format
		if err != nil {
			render.RenderTemplateWithData(w, "Registration.html", EditPageData{
				Error:     "Password hashing failed",
				Countries: countries,
				User:      user,
				SportsMap: sportsMap,
			})
			return
		}

		_, err = db.Exec("INSERT INTO New(username, password, email, mobile, address, gender, sports, dob, country) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", username, hashed, email, mobile, address, gender, joinedSports, dob, country)

		if err != nil {
			errMsg := err.Error()
			userMessage := "Registration failed. Please try again."

			// Customize user-friendly messages
			if strings.Contains(errMsg, "UNIQUE constraint failed: New.email") {
				userMessage = "Email already used, try a different one."
			} else if strings.Contains(errMsg, "UNIQUE constraint failed: New.mobile") {
				userMessage = "Mobile number already registered."
			}
			render.RenderTemplateWithData(w, "Registration.html", EditPageData{
				Error:     userMessage,
				Countries: countries,
				User:      user,
				SportsMap: sportsMap,
			})
			return
		}
		setFlashMessage(w, "User successfully registered!")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	render.RenderTemplateWithData(w, "Registration.html", EditPageData{
		Countries: countries,
		Title:     "Add User",
	})
}

func ResetHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	ts := r.URL.Query().Get("ts")
	token := r.URL.Query().Get("token")

	expectedHash := sha256.Sum256([]byte(email + ts + resetsecret))
	expectedToken := hex.EncodeToString(expectedHash[:])

	if token != expectedToken {
		render.RenderTemplateWithData(w, "Reset.html", EditPageData{
			Error: "Invalid or tampered reset link",
		})
		return
	}

	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || time.Now().Unix()-tsInt > 15*60 {
		render.RenderTemplateWithData(w, "Reset.html", EditPageData{
			Error: "Reset link has expired.",
		})
		return
	}

	if r.Method == http.MethodPost {
		newPass := r.FormValue("password")
		confirm := r.FormValue("confirm")

		if newPass != confirm {
			render.RenderTemplateWithData(w, "Reset.html", EditPageData{
				Error: "Passwords donot match.",
			})
			return
		}

		hashed, _ := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
		_, err := db.Exec("Update Admin SET password=? WHERE email = ?", hashed, email)

		if err != nil {
			render.RenderTemplateWithData(w, "Reset.html", EditPageData{
				Error: "Failed to update password",
			})
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	render.RenderTemplateWithData(w, "Reset.html", EditPageData{
		Email: email,
		Ts:    ts,
		Token: token,
	})
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	const limit = 5
	page := 1
	pageStr := r.URL.Query().Get("page")
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	offset := (page - 1) * limit

	rows, err := db.Query("SELECT id, username, email, mobile FROM New ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
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
	db.QueryRow("SELECT COUNT(*) FROM New").Scan(&total)
	totalPages := (total + limit - 1) / limit

	flash := getFlashMessage(w, r)

	render.RenderTemplateWithData(w, "Home.html", EditPageData{
		Users:      users,
		Page:       page,
		TotalPages: totalPages,
		Error:      flash,
	})
}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		render.RenderTemplateWithData(w, "Home.html", EditPageData{Error: "Missing user ID"})
		return
	}

	var user User
	err := db.QueryRow("SELECT id, username, email, mobile, address, gender, sports, dob, country FROM New WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Mobile, &user.Address, &user.Gender, &user.Sports, &user.DOB, &user.Country)

	if err != nil {
		render.RenderTemplateWithData(w, "Home.html", EditPageData{Error: "User not found"})
		return
	}

	countries, _ := getCountriesFromDB()

	// Convert sports (comma-separated string) to map for checkbox logic
	sportsMap := make(map[string]bool)
	for _, sport := range strings.Split(user.Sports, ",") {
		sport = strings.TrimSpace(sport)
		if sport != "" {
			sportsMap[sport] = true
		}
	}

	// Fix date
	if len(user.DOB) > 10 {
		user.DOB = user.DOB[:10]
	}

	render.RenderTemplateWithData(w, "Edit.html", EditPageData{
		Title:     "Edit User",
		User:      user,
		Countries: countries,
		SportsMap: sportsMap,
	})
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		username := r.FormValue("username")
		mobile := r.FormValue("mobile")
		address := r.FormValue("address")
		gender := r.FormValue("gender")
		dobStr := r.FormValue("dob")
		country := r.FormValue("country")
		sportsSlice := r.Form["sports"]

		sports := strings.Join(sportsSlice, ",")

		// Validate mobile format
		match, err := regexp.MatchString(`^(\+\d{1,3})?\d{10}$`, mobile)
		if err != nil || !match {
			setFlashMessage(w, "Invalid mobile format")
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}

		//First changing dob from string to time format
		dob, err := time.Parse("2006-01-02", dobStr)
		if err != nil || dob.After(time.Now()) {
			setFlashMessage(w, "Invalid DOB")
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}

		setFlashMessage(w, "Updated Successfully")

		_, err = db.Exec(`UPDATE New SET username=?, mobile=?, address=?, gender=?, sports=?, dob=?, country=? WHERE id=?`,
			username, mobile, address, gender, sports, dob, country, id)
		if err != nil {
			setFlashMessage(w, "Update failed: "+err.Error())
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		setFlashMessage(w, "User successfully updated!")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		_, err := db.Exec("DELETE FROM New WHERE id = ?", id)
		if err != nil {
			setFlashMessage(w, "Error deleting user")
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		setFlashMessage(w, "User deleted!")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}
