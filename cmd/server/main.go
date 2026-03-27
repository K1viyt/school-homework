package main

import (
	"log"
	"net/http"

	"github.com/K1viyt/school-homework/internal/database"
	"github.com/K1viyt/school-homework/internal/handlers"
)

func main() {
	database.Init()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Сервер работает!"))
	})
	http.HandleFunc("/upload", handlers.UploadHandler)
	http.HandleFunc("/homeworks", handlers.ListHomeworksHandler)
	http.HandleFunc("/homeworks/delete/", handlers.DeleteHomeworkHandler)
	http.HandleFunc("/homeworks/update/", handlers.UpdateHomeworkHandler)
	http.HandleFunc("/homeworks/replace/", handlers.ReplaceHomeworkHandler)

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
