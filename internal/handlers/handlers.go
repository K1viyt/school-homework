package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Разрешён только POST", http.StatusMethodNotAllowed)
		return // ← без этого код ниже выполнится!
	}
	file, h, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	f, err := os.Create(filepath.Join("uploads", h.Filename))
	if err != nil {
		http.Error(w, "Не удалось сохранить файл", http.StatusInternalServerError)
		log.Println("Ошибка создания файла:", err)
		return
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "файл %s загружен", h.Filename)

}
