package handlers

import (
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/K1viyt/school-homework/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type Homework struct {
	ID          int     `json:"id"`
	Filename    string  `json:"filename"`
	Filepath    string  `json:"filepath"`
	Subject     *string `json:"subject"`
	Description *string `json:"description"`
	UploadedAt  string  `json:"uploaded_at"`
}

type User struct {
	ID       int
	Username string
	Role     string
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	//Проверка на POST
	if r.Method != http.MethodPost {
		http.Error(w, "Разрешён только POST", http.StatusMethodNotAllowed)
		return // ← без этого код ниже выполнится!
	}
	// Проверка авторизации
	user, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	// Проверка роли (только для хэндлеров учителя)
	if user.Role != "teacher" {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
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
	// Проверка авторизации
	user, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	// Проверка роли (только для хэндлеров учителя)
	if user.Role != "teacher" {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
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
	// Проверка авторизации
	user, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	// Проверка роли (только для хэндлеров учителя)
	if user.Role != "teacher" {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
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
	// Проверка авторизации
	user, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	// Проверка роли (только для хэндлеров учителя)
	if user.Role != "teacher" {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
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

func ReplaceHomeworkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Разрешон только PUT", http.StatusMethodNotAllowed)
		return
	}
	// Проверка авторизации
	user, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	// Проверка роли (только для хэндлеров учителя)
	if user.Role != "teacher" {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}

	idSrs := strings.TrimPrefix(r.URL.Path, "/homeworks/replace/")
	id, err := strconv.Atoi(idSrs)
	if err != nil {
		http.Error(w, "Некоректный тип данных", http.StatusBadRequest)
		return
	}
	var fpath string
	err = database.DB.QueryRow(`SELECT filepath FROM homework WHERE id=?`, id).Scan(&fpath)
	if err != nil {
		http.Error(w, "Такого id нету", http.StatusNotFound)
		return
	}
	os.Remove(fpath)

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
		http.Error(w, "Ошибка копирования в новый файл", http.StatusBadRequest)
		return
	}
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
	_, err = database.DB.Exec(`UPDATE homework SET filename=?,filepath=?,subject=?,description=? WHERE id=?`, h.Filename, filepath.Join("uploads", h.Filename), subjectVal, descriptionVal, id)
	if err != nil {
		http.Error(w, "Ошибка перезаписи файла в базу", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Задание заменено",
	})
}
func genarateUsername(fullName string) string {
	// берём первое слово из полного имени (имя)
	parts := strings.Fields(fullName)
	base := strings.ToLower(parts[0])
	// добавляем 4 случайных цифры
	suffix := fmt.Sprintf("%04d", rand.Intn(10000))
	return base + "_" + suffix

}

// Данный токен выдается user при успешной авторизации
func genaratToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := crand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func RegistrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Разрешон только POST", http.StatusMethodNotAllowed)
		return
	}
	var role string
	//Достаем нужные значения из БД
	password := r.FormValue("password")
	fullName := r.FormValue("full_name")
	login := genarateUsername(fullName)
	//Защита для роли учителя
	if r.FormValue("invite_code") == "SECRET123" {
		role = "teacher"
	} else {
		role = "student"
	}
	//Пишем пароль в hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка записи pasword в hash", http.StatusBadRequest)
		return
	}
	//Записываем данные в БД
	_, err = database.DB.Exec(`INSERT INTO users(full_name,role,password,username) VALUES(?,?,?,?)`, fullName, role, hash, login)
	if err != nil {
		http.Error(w, "Ошибка загрузки данных пользователя в базу", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Регистрация успешна",
		"username": login,
	})

}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Разрешон только Post", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var storedHash string
	var role string
	var userID int
	err := database.DB.QueryRow(
		`SELECT password, role,id FROM users WHERE username=?`, username,
	).Scan(&storedHash, &role, &userID)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}
	token, err := genaratToken()
	if err != nil {
		http.Error(w, "Ошибка генерации token", http.StatusInternalServerError)
		return
	}
	_, err = database.DB.Exec(`INSERT INTO sessions(user_id,token),VALUES(?,?)`, userID, token)
	if err != nil {
		http.Error(w, "Ошибка передачи данных сессии пользователя", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Вы успешно авторизованны",
		"username": username,
		"role":     role,
		"token":    token,
	})
}
func getUserFromToken(r *http.Request) (*User, error) {
	var user User
	// 1. Достать заголовок Authorization
	authHeader := r.Header.Get("Authorization")

	// 2. Убрать префикс "Bearer "
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return nil, fmt.Errorf("токен отсутствует")
	}

	err := database.DB.QueryRow(`SELECT users.id,users.username,users.role FROM sessions JOIN users ON sessions.user_id = users.id WHERE sessions.token = ? AND sessions.created_at>datetime('now','-24 hours')`, token).Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {

		return &user, fmt.Errorf("сессия не найдена или истекла")
	}
	return &user, nil

}
