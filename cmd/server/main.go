package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"biviz/internal/db"
	"biviz/internal/handlers"
	"biviz/internal/middleware"
)

func main() {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgresql://dev:devpass@localhost:5432/devdb")

	if err := db.Connect(dbURL); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatal(err)
	}

	// 페이지별 독립 템플릿 로드
	handlers.PageTemplates = map[string]*template.Template{
		"login.html": template.Must(template.ParseFiles(
			"templates/layouts/base.html",
			"templates/pages/login.html",
		)),
		"signup.html": template.Must(template.ParseFiles(
			"templates/layouts/base.html",
			"templates/pages/signup.html",
		)),
		"dashboard.html": template.Must(template.ParseFiles(
			"templates/layouts/base.html",
			"templates/pages/dashboard.html",
		)),
	}

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /login", handlers.ShowLogin)
	mux.HandleFunc("GET /signup", handlers.ShowSignup)
	mux.HandleFunc("POST /api/login", handlers.HandleLogin)
	mux.HandleFunc("POST /api/signup", handlers.HandleSignup)
	mux.HandleFunc("POST /api/logout", handlers.HandleLogout)

	mux.Handle("GET /dashboard", middleware.AuthMiddleware(http.HandlerFunc(handlers.ShowDashboard)))

	log.Printf("🚀 BiViz 서버 시작: http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
