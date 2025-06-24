package handler

import (
	"encoding/base64"
	"io"
	"mysqliteapp/render"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

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

		//image
		file, _, err := r.FormFile("image")
		if err != nil {
			render.RenderTemplateWithData(w, "Registration.html", EditPageData{
				Error: "Error in image uploading",
			})
			return
		}
		defer file.Close()

		imageData, err := io.ReadAll(file)
		if err != nil {
			render.RenderTemplateWithData(w, "Registration.html", EditPageData{
				Error: "Error in image uploading",
			})
			return
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

		_, err = DB.Exec("INSERT INTO New(username, password, email, mobile, address, gender, sports, dob, country, image) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", username, hashed, email, mobile, address, gender, joinedSports, dob, country, imageData)

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

func EditHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		render.RenderTemplateWithData(w, "Home.html", EditPageData{Error: "Missing user ID"})
		return
	}

	//Fetching and displaying image
	var imageBytes []byte
	var user User
	err := DB.QueryRow("SELECT id, username, email, mobile, address, gender, sports, dob, country, image FROM New WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Email, &user.Mobile, &user.Address, &user.Gender, &user.Sports, &user.DOB, &user.Country, &imageBytes)

	user.image = imageBytes

	if err != nil {
		render.RenderTemplateWithData(w, "Home.html", EditPageData{Error: "User not found"})
		return
	}

	// Convert image bytes to base64
	if len(imageBytes) > 0 {
		user.ImageBase64 = base64.StdEncoding.EncodeToString(imageBytes)
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

		file, _, err := r.FormFile("image")
		var imageData []byte
		if err == nil {
			defer file.Close()
			imageData, _ = io.ReadAll(file)
		}

		setFlashMessage(w, "Updated Successfully")
		if len(imageData) > 0 {
			_, err = DB.Exec(`UPDATE New SET username=?, mobile=?, address=?, gender=?, sports=?, dob=?, country=?, image=? WHERE id=?`, username, mobile, address, gender, sports, dob, country, imageData, id)
		} else {
			_, err = DB.Exec(`UPDATE New SET username=?, mobile=?, address=?, gender=?, sports=?, dob=?, country=? WHERE id=?`,
				username, mobile, address, gender, sports, dob, country, id)
		}

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
		_, err := DB.Exec("DELETE FROM New WHERE id = ?", id)
		if err != nil {
			setFlashMessage(w, "Error deleting user")
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}
		setFlashMessage(w, "User deleted!")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}
