package main

import (
	"log"
	"net/http"

	"github.com/K1viyt/school-homework/internal/handlers"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Сервер работает!"))
	})
	http.HandleFunc("/upload", handlers.UploadHandler)

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
