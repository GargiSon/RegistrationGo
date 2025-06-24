package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"mysqliteapp/render"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const resetsecret = "hubjinkom"

func sendResetEmail(toEmail, resetLink string) error {
	email := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	auth := smtp.PlainAuth("", email, password, "smtp.gmail.com")

	subject := "Subject: Password Reset Link\n"
	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<body style="font-family: Arial, sans-serif; line-height: 1.6;">
		<p>Hello,</p>
		<p>Click the link below to reset your password:</p>
		<p><a href="%s" style="background-color: #007BFF; color: white; padding: 10px 15px; text-decoration: none; border-radius: 5px;">Reset Password</a></p>
		<p>Or copy and paste this URL into your browser:</p>
		<p>%s</p>
		<br>
		<p>If you didn’t request this, please ignore this email.</p>
		<p>Thanks,<br>Your Team</p>
	</body>
	</html>
	`, resetLink, resetLink)

	msg := []byte(subject + headers + body)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		email,
		[]string{toEmail},
		msg,
	)
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
		err := DB.QueryRow("SELECT password FROM AdminNew WHERE email = ?", email).Scan(&storedHash)
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
		err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM AdminNew WHERE email = ?)", email).Scan(&exists)
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

func ResetHandler(w http.ResponseWriter, r *http.Request) {
	var email, ts, token string

	if r.Method == http.MethodGet {
		email = r.URL.Query().Get("email")
		ts = r.URL.Query().Get("ts")
		token = r.URL.Query().Get("token")
	} else if r.Method == http.MethodPost {
		email = r.FormValue("email")
		ts = r.FormValue("ts")
		token = r.FormValue("token")
	}

	expectedHash := sha256.Sum256([]byte(email + ts + resetsecret))
	expectedToken := hex.EncodeToString(expectedHash[:])

	if token != expectedToken {
		http.Error(w, "Invalid or tampered reset link.", http.StatusUnauthorized)
		return
	}

	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || time.Now().Unix()-tsInt > 15*60 {
		http.Error(w, "Reset link has expired.", http.StatusUnauthorized)
		return
	}
	if r.Method == http.MethodPost {
		newPass := r.FormValue("password")
		confirm := r.FormValue("confirm")
		if newPass != confirm {
			render.RenderTemplateWithData(w, "Reset.html", EditPageData{
				Error: "Passwords do not match.",
				Email: email,
				Ts:    ts,
				Token: token,
			})
			return
		}
		hashed, _ := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
		_, err := DB.Exec("UPDATE AdminNew SET password = ? WHERE email = ?", hashed, email)
		if err != nil {
			render.RenderTemplateWithData(w, "Reset.html", EditPageData{
				Error: "Failed to update password",
				Email: email,
				Ts:    ts,
				Token: token,
			})
			return
		}
		setFlashMessage(w, "Password updated successfully.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	render.RenderTemplateWithData(w, "Reset.html", EditPageData{
		Email: email,
		Ts:    ts,
		Token: token,
	})
}
