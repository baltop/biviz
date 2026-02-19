package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"biviz/internal/auth"
	"biviz/internal/middleware"
)

var PageTemplates map[string]*template.Template

func render(w http.ResponseWriter, page string, data interface{}) {
	tmpl, ok := PageTemplates[page]
	if !ok {
		http.Error(w, "페이지를 찾을 수 없습니다", 500)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("템플릿 렌더링 실패: %v", err)
		http.Error(w, "서버 오류", 500)
	}
}

// renderHTMXError HTMX 에러 응답 (부분 HTML)
func renderHTMXError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	fmt.Fprintf(w, `<div class="flex items-center gap-3 p-4 mb-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm">
		<span>%s</span>
	</div>`, template.HTMLEscapeString(message))
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func ShowLogin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		if session := middleware.GetSession(cookie.Value); session != nil {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}
	render(w, "login.html", map[string]interface{}{})
}

func ShowSignup(w http.ResponseWriter, r *http.Request) {
	render(w, "signup.html", map[string]interface{}{})
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || password == "" {
		if isHTMX(r) {
			renderHTMXError(w, "이메일과 비밀번호를 입력해주세요")
			return
		}
		render(w, "login.html", map[string]interface{}{"Error": "이메일과 비밀번호를 입력해주세요"})
		return
	}

	user, err := auth.Login(r.Context(), email, password)
	if err != nil {
		if isHTMX(r) {
			renderHTMXError(w, "이메일 또는 비밀번호가 올바르지 않습니다")
			return
		}
		render(w, "login.html", map[string]interface{}{"Error": "이메일 또는 비밀번호가 올바르지 않습니다"})
		return
	}

	token, err := middleware.CreateSession(user.ID, user.Email)
	if err != nil {
		log.Printf("세션 생성 실패: %v", err)
		if isHTMX(r) {
			renderHTMXError(w, "서버 오류가 발생했습니다")
			return
		}
		render(w, "login.html", map[string]interface{}{"Error": "서버 오류가 발생했습니다"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func HandleSignup(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if email == "" || password == "" {
		if isHTMX(r) {
			renderHTMXError(w, "이메일과 비밀번호를 입력해주세요")
			return
		}
		render(w, "signup.html", map[string]interface{}{"Error": "이메일과 비밀번호를 입력해주세요"})
		return
	}

	if len(password) < 8 {
		if isHTMX(r) {
			renderHTMXError(w, "비밀번호는 8자 이상이어야 합니다")
			return
		}
		render(w, "signup.html", map[string]interface{}{"Error": "비밀번호는 8자 이상이어야 합니다"})
		return
	}

	if password != confirmPassword {
		if isHTMX(r) {
			renderHTMXError(w, "비밀번호가 일치하지 않습니다")
			return
		}
		render(w, "signup.html", map[string]interface{}{"Error": "비밀번호가 일치하지 않습니다"})
		return
	}

	user, err := auth.Signup(r.Context(), email, password)
	if err != nil {
		msg := "서버 오류가 발생했습니다"
		if err == auth.ErrEmailExists {
			msg = "이미 등록된 이메일입니다"
		}
		if isHTMX(r) {
			renderHTMXError(w, msg)
			return
		}
		render(w, "signup.html", map[string]interface{}{"Error": msg})
		return
	}

	token, err := middleware.CreateSession(user.ID, user.Email)
	if err != nil {
		log.Printf("세션 생성 실패: %v", err)
		if isHTMX(r) {
			renderHTMXError(w, "가입 완료되었지만 로그인에 실패했습니다")
			return
		}
		render(w, "signup.html", map[string]interface{}{"Error": "가입 완료되었지만 로그인에 실패했습니다"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		middleware.DestroySession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/login")
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
