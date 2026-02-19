package handlers

import (
	"net/http"

	"biviz/internal/auth"
	"biviz/internal/middleware"
)

func ShowDashboard(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetCurrentSession(r)
	if session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := auth.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	render(w, "dashboard.html", map[string]interface{}{
		"User": user,
	})
}
