package handler

import (
	"database/sql"
	"log"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	_ = godotenv.Load()
	var err error
	DB, err = sql.Open("sqlite", "./New.db") //driver name, datasource name
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
		country TEXT NOT NULL,
		image BLOB
	);`

	createCountryTable := `
	CREATE TABLE IF NOT EXISTS Countries(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`

	createAdminTable := `
	CREATE TABLE IF NOT EXISTS AdminNew(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);`

	if _, err = DB.Exec(createTable); err != nil { //returns res, err but res not used here
		log.Println("Error") //this stops the program immediately, when an error occurs
	}

	if _, err = DB.Exec(createCountryTable); err != nil {
		log.Println("Error")
	}

	if _, err = DB.Exec(createAdminTable); err != nil {
		log.Println("Error")
	}

	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM Countries").Scan(&count); err != nil {
		log.Println("Error")
	}

	if count == 0 {
		_, err = DB.Exec(`INSERT INTO Countries (name) VALUES ('INDIA'),('AFGHANISTHAN'),('FRANCE')`)
		if err != nil {
			log.Println("Error")
		}
	}

	var adminCount int
	_ = DB.QueryRow("SELECT COUNT(*) FROM AdminNew").Scan(&adminCount)
	if adminCount == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("admin1001"), bcrypt.DefaultCost)
		_, _ = DB.Exec("INSERT INTO AdminNew(email, password) VALUES (?, ?)", "gargi.soni@loginradius.com", hashed)
	}
}
