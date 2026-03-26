package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/K1viyt/school-homework/internal/database"
)

type Homework struct {
	ID          int     `json:"id"`
	Filename    string  `json:"filename"`
	Filepath    string  `json:"filepath"`
	Subject     *string `json:"subject"`
	Description *string `json:"description"`
	UploadedAt  string  `json:"uploaded_at"`
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	//Проверка на POST
	if r.Method != http.MethodPost {
		http.Error(w, "Разрешён только POST", http.StatusMethodNotAllowed)
		return // ← без этого код ниже выполнится!
	}
	//Получаем мето данные файла и закидывем сам файл в память
	file, h, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()
	//Создаем файл по пути...
	f, err := os.Create(filepath.Join("uploads", h.Filename))
	if err != nil {
		http.Error(w, "Не удалось сохранить файл", http.StatusInternalServerError)
		log.Println("Ошибка создания файла:", err)
		return
	}
	defer f.Close()
	//Копируем file в f
	_, err = io.Copy(f, file)
	if err != nil {
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}
	//Считываем инфу из таблицы
	subject := r.FormValue("subject")
	description := r.FormValue("description")
	var subjectVal any
	var descriptionVal any

	if subject != "" {
		subjectVal = subject
	} else {
		subjectVal = nil
	}
	if description != "" {
		descriptionVal = description

	} else {
		descriptionVal = nil
	}
	_, err = database.DB.Exec(`INSERT INTO homework(filename, filepath, subject, description) VALUES (?, ?, ?, ?)`,
		h.Filename, filepath.Join("uploads", h.Filename), subjectVal, descriptionVal)
	if err != nil {
		http.Error(w, "Ошибка записи в базу данных", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "файл %s загружен", h.Filename)

}

func ListHomeworksHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Проверяем что метод GET
	if r.Method != http.MethodGet {
		http.Error(w, "Разрешон только GET", http.StatusMethodNotAllowed)
		return
	}
	// 2. Делаем SELECT-запрос к базе данных
	rows, err := database.DB.Query(`SELECT id,filename,filepath,subject,description FROM homework`)

	// 3. Проходим по результатам и собираем их в слайс структур
	var homeworks []Homework
	if err != nil {
		http.Error(w, "Ошибка запроса к базе данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var hw Homework

		err := rows.Scan(&hw.ID, &hw.Filename, &hw.Filepath, &hw.Subject, &hw.Description)
		if err != nil {
			http.Error(w, "Некоректный тип данных", http.StatusBadRequest)
			return
		}
		// 3. Проходим по результатам и собираем их в слайс структур
		homeworks = append(homeworks, hw)
	}

	// 4. Превращаем слайс в JSON\
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(homeworks)
	// 5. Отправляем JSON клиенту
}

func DeleteHomeworkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Разрешон только Delete", http.StatusMethodNotAllowed)
		return
	}
	//Получаю ID
	idStr := strings.TrimPrefix(r.URL.Path, "/homeworks/delete/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный id", http.StatusBadRequest)
		return
	}
	//Удаляем по ID,filepath
	var fpath string
	err = database.DB.QueryRow(`SELECT filepath FROM homework WHERE id=?`, id).Scan(&fpath)
	if err != nil {
		http.Error(w, "Задание не найдено", http.StatusNotFound)
		return
	}
	_, err = database.DB.Exec(`DELETE FROM homework WHERE id=?`, id)
	if err != nil {
		http.Error(w, "Ошибка удаления из базы", http.StatusInternalServerError)
		return
	}
	os.Remove(fpath)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Задание удалено",
	})

}
func UpdateHomeworkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Разрешон только Patch", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Subject     string `json:"subject"`
		Description string `json:"description"`
	}

	idSrs := strings.TrimPrefix(r.URL.Path, "/homeworks/update/")
	id, err := strconv.Atoi(idSrs)
	if err != nil {
		http.Error(w, "Некорректный id", http.StatusBadRequest)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	result, err := database.DB.Exec(`UPDATE homework SET subject=?, description=? WHERE id=?`, input.Subject, input.Description, id)
	if err != nil {
		http.Error(w, "Ошибка обновления в базе", http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "Задание не найдено", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Задание обновлено",
	})
}
