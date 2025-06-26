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
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte("super-secret-session-key")) //This means session is stored in client side browser cookies

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
	<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 40px 0;">
		<div style="max-width: 600px; margin: auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
			<p style="font-size: 18px;">Hello,</p>
			<p style="font-size: 16px;">Click the button below to reset your password:</p>
			<p style="text-align: center;">
				<a href="%s" style="display: inline-block; background-color: #007BFF; color: white; padding: 12px 20px; text-decoration: none; border-radius: 5px; font-size: 16px;">Reset Password</a>
			</p>
			<p style="font-size: 14px;">Or copy and paste this URL into your browser:</p>
			<p style="word-break: break-all; font-size: 14px; color: #333;">%s</p>
			<br>
			<p style="font-size: 14px;">If you didn’t request this, please ignore this email.</p>
			<p style="font-size: 14px;">Thanks,<br><strong>Your Team</strong></p>
		</div>
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
		render.RenderTemplateWithData(w, "Login.html", EditPageData{})
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var storedHash string
	err := DB.QueryRow("SELECT password FROM AdminNew WHERE email = ?", email).Scan(&storedHash)
	if err != nil {
		render.RenderTemplateWithData(w, "Login.html", EditPageData{
			Error: "Invalid email or password",
		})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) != nil {
		render.RenderTemplateWithData(w, "Login.html", EditPageData{
			Error: "Invalid email or password",
		})
		return
	}

	//Session is created
	session, _ := store.Get(r, "session")
	session.Values["authenticated"] = true
	session.Values["email"] = email
	// Get name before '@'
	parts := strings.Split(email, "@")
	username := parts[0]
	session.Values["admin_name"] = username
	err = session.Save(r, w)

	if err != nil {
		render.RenderTemplateWithData(w, "Login.html", EditPageData{
			Error: "Failed to start session",
		})
		return
	}
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1 //Expire or delete cookie
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		render.RenderTemplateWithData(w, "Forgot.html", EditPageData{Info: getFlashMessage(w, r)})
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
	email := r.FormValue("email")
	ts := r.FormValue("ts")
	token := r.FormValue("token")

	expectedHash := sha256.Sum256([]byte(email + ts + resetsecret))
	expectedToken := hex.EncodeToString(expectedHash[:])

	if token != expectedToken {
		http.Error(w, "Invalid or tampered reset link.", http.StatusUnauthorized)
		return
	}

	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || time.Now().Unix()-tsInt > 900 {
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
