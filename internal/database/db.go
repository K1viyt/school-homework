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
	DB = db

}
