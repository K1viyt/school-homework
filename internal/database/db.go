package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	db, err := sql.Open("sqlite3", "homework.db")
	if err != nil {
		log.Fatal("Ошибка соеденения с BD: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("BD не отвечает: ", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS homework(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	filename TEXT NOT NULL,
	filepath TEXT NOT NULL,
	subject TEXT,
	description TEXT,
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)

	if err != nil {
		log.Fatal("Невозможно создать BD: ", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	full_name TEXT NOT NULL,
	username TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	role TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		log.Fatal("Невозможно созадть BD под users: ", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sessions(id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		log.Fatal("Невозможно созадть BD под sessions: ", err)
	}
	DB = db

}
