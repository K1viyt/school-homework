package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/K1viyt/school-homework/internal/database"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Разрешон только GET", http.StatusMethodNotAllowed)
		return
	}
	var fullName, username, role string
	var class, subject sql.NullString
	user, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Пользователь по данному токену не найдет", http.StatusUnauthorized)
		return
	}
	row := database.DB.QueryRow(`SELECT full_name, username, role, class, subject FROM users WHERE id=$1`, user.ID)
	err = row.Scan(&fullName, &username, &role, &class, &subject)
	if err != nil {
		http.Error(w, "Профиль ненайдет", http.StatusNotFound)
		return
	}

	if role == "teacher" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"ФИО":      fullName,
			"username": username,
			"role":     role,
			"subject":  subject.String,
		})
	}

	if role == "student" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"full_name": fullName,
			"username":  username,
			"role":      role,
			"class":     class.String,
		})
	}

}
