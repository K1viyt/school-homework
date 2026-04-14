package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/K1viyt/school-homework/internal/database"
)

type Schedule struct {
	ID         int     `json:"id"`
	ClassName  string  `json:"class_name"`
	WeekParity string  `json:"week_parity"` // "odd" | "even"
	DayOfWeek  int     `json:"day_of_week"` // 1..6 (Пн..Сб)
	LessonNum  int     `json:"lesson_num"`  // 1..10
	Subject    string  `json:"subject"`
	TeacherID  int     `json:"teacher_id"`
	Room       *string `json:"room"` // может быть NULL
	CreatedAt  string  `json:"created_at"`
}

func GetScheduleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Разрешен только GET", http.StatusMethodNotAllowed)
		return
	}
	_, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "User не найден", http.StatusUnauthorized)
		return
	}
	className := r.URL.Query().Get("class")
	if className == "" {
		http.Error(w, "Класс пуст", http.StatusBadRequest)
		return
	}
	week := r.URL.Query().Get("week")
	if week != "odd" && week != "even" {
		http.Error(w, "week должен быть 'odd' или 'even'", http.StatusBadRequest)
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, class_name, week_parity, day_of_week, lesson_num, subject, teacher_id, room
		FROM schedule
		WHERE class_name = ? AND week_parity = ?
		ORDER BY day_of_week, lesson_num
	`, className, week)
	if err != nil {
		http.Error(w, "Ошибка запроса к базе данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	schedule := []Schedule{}
	for rows.Next() {
		var sh Schedule
		if err := rows.Scan(&sh.ID, &sh.ClassName, &sh.WeekParity, &sh.DayOfWeek, &sh.LessonNum, &sh.Subject, &sh.TeacherID, &sh.Room); err != nil {
			http.Error(w, "Ошибка чтения строки", http.StatusInternalServerError)
			return
		}
		schedule = append(schedule, sh)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Ошибка итерации по результатам", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}
