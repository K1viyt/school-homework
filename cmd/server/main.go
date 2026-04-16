package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/K1viyt/school-homework/internal/database"
	"github.com/K1viyt/school-homework/internal/handlers"
)

func main() {
	// ИСПРАВЛЕНИЕ БАГА #3:
	// Создаём папку uploads при старте сервера, если её нет.
	// os.MkdirAll — создаёт папку и все промежуточные папки в пути.
	// 0755 — права доступа: владелец может читать/писать/выполнять,
	//         остальные — только читать и "заходить" в папку.
	// Если папка уже существует — ошибки не будет, всё нормально.
	err := os.MkdirAll("uploads", 0755)
	if err != nil {
		log.Fatal("Не удалось создать папку uploads: ", err)
	}
	// Пытаемся подгрузить .env, а если его нет — password.env.
	// Обе ошибки молча игнорируем: если файла нет, переменные могут быть заданы в окружении.
	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("password.env")
	}

	database.Init()

	// Отдаём фронтенд из папки web/ на все неразмеченные пути.
	// Более специфичные маршруты ниже перехватывают /login, /upload и т.д.
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/upload", handlers.UploadHandler)
	http.HandleFunc("/homeworks", handlers.ListHomeworksHandler)
	http.HandleFunc("/homeworks/delete/", handlers.DeleteHomeworkHandler)
	http.HandleFunc("/homeworks/update/", handlers.UpdateHomeworkHandler)
	http.HandleFunc("/homeworks/replace/", handlers.ReplaceHomeworkHandler)
	http.HandleFunc("/registr", handlers.RegistrHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/profile", handlers.ProfileHandler)
	http.HandleFunc("GET /schedule", handlers.GetScheduleHandler)
	http.HandleFunc("POST /schedule", handlers.CreateScheduleHandler)
	http.HandleFunc("PATCH /schedule/update/{id}", handlers.UpdateScheduleHandler)
	http.HandleFunc("DELETE /schedule/delete/{id}", handlers.DeleteScheduleHandler)
	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
