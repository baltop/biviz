package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"biviz/internal/auth"
	"biviz/internal/db"
	"biviz/internal/middleware"
)

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://dev:devpass@localhost:5432/devdb"
	}

	if err := db.Connect(dbURL); err != nil {
		fmt.Fprintf(os.Stderr, "테스트 DB 연결 실패: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "마이그레이션 실패: %v\n", err)
		os.Exit(1)
	}

	// 테스트용 최소 템플릿 등록
	PageTemplates = map[string]*template.Template{
		"login.html": template.Must(template.New("base.html").Parse(
			`{{define "base.html"}}LOGIN_PAGE {{.Error}}{{end}}`,
		)),
		"signup.html": template.Must(template.New("base.html").Parse(
			`{{define "base.html"}}SIGNUP_PAGE {{.Error}}{{end}}`,
		)),
		"dashboard.html": template.Must(template.New("base.html").Parse(
			`{{define "base.html"}}DASHBOARD {{.User.Email}}{{end}}`,
		)),
	}

	code := m.Run()

	db.Pool.Exec(context.Background(), "DELETE FROM users WHERE email LIKE '%@handler.test.local'")
	os.Exit(code)
}

func testEmail(name string) string {
	return fmt.Sprintf("%s@handler.test.local", name)
}

func cleanup(t *testing.T, email string) {
	t.Helper()
	db.Pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)
}

func formBody(vals map[string]string) *strings.Reader {
	v := url.Values{}
	for k, val := range vals {
		v.Set(k, val)
	}
	return strings.NewReader(v.Encode())
}

func formRequest(method, path string, vals map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, formBody(vals))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// ── 로그인 페이지 ──

// H-01: 로그인 페이지 렌더링
func TestShowLogin_Render(t *testing.T) {
	req := httptest.NewRequest("GET", "/login", nil)
	rec := httptest.NewRecorder()
	ShowLogin(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "LOGIN_PAGE") {
		t.Error("로그인 페이지 렌더링 안 됨")
	}
}

// H-01b: 이미 로그인된 사용자 → 대시보드 리다이렉트
func TestShowLogin_AlreadyLoggedIn(t *testing.T) {
	token, _ := middleware.CreateSession(1, "already@test.local")
	defer middleware.DestroySession(token)

	req := httptest.NewRequest("GET", "/login", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rec := httptest.NewRecorder()
	ShowLogin(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}
	if rec.Header().Get("Location") != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", rec.Header().Get("Location"))
	}
}

// ── 회원가입 ──

// H-02: 회원가입 성공
func TestHandleSignup_Success(t *testing.T) {
	email := testEmail("h-signup-ok")
	defer cleanup(t, email)

	req := formRequest("POST", "/api/signup", map[string]string{
		"email": email, "password": "password123", "confirm_password": "password123",
	})
	rec := httptest.NewRecorder()
	HandleSignup(rec, req)

	// 쿠키 설정 확인
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session_token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("session_token 쿠키 없음")
	}

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}
}

// H-03: 회원가입 — 빈 입력
func TestHandleSignup_Empty(t *testing.T) {
	req := formRequest("POST", "/api/signup", map[string]string{
		"email": "", "password": "", "confirm_password": "",
	})
	rec := httptest.NewRecorder()
	HandleSignup(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200 (에러 페이지 렌더링)", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "입력해주세요") {
		t.Errorf("에러 메시지 없음: %s", body)
	}
}

// H-04: 회원가입 — 비밀번호 너무 짧음
func TestHandleSignup_ShortPassword(t *testing.T) {
	req := formRequest("POST", "/api/signup", map[string]string{
		"email": "short@test.local", "password": "1234567", "confirm_password": "1234567",
	})
	rec := httptest.NewRecorder()
	HandleSignup(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "8자 이상") {
		t.Errorf("비밀번호 길이 에러 없음: %s", body)
	}
}

// H-05: 회원가입 — 비밀번호 불일치
func TestHandleSignup_PasswordMismatch(t *testing.T) {
	req := formRequest("POST", "/api/signup", map[string]string{
		"email": "mismatch@test.local", "password": "password123", "confirm_password": "different",
	})
	rec := httptest.NewRecorder()
	HandleSignup(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "일치하지") {
		t.Errorf("불일치 에러 없음: %s", body)
	}
}

// H-06: 회원가입 — 중복 이메일
func TestHandleSignup_DuplicateEmail(t *testing.T) {
	email := testEmail("h-dup")
	defer cleanup(t, email)

	auth.Signup(context.Background(), email, "password123")

	req := formRequest("POST", "/api/signup", map[string]string{
		"email": email, "password": "password123", "confirm_password": "password123",
	})
	rec := httptest.NewRecorder()
	HandleSignup(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "이미 등록된") {
		t.Errorf("중복 이메일 에러 없음: %s", body)
	}
}

// H-07: HTMX 회원가입 성공 → HX-Redirect
func TestHandleSignup_HTMX_Success(t *testing.T) {
	email := testEmail("h-htmx-signup")
	defer cleanup(t, email)

	req := formRequest("POST", "/api/signup", map[string]string{
		"email": email, "password": "password123", "confirm_password": "password123",
	})
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	HandleSignup(rec, req)

	if rec.Header().Get("HX-Redirect") != "/dashboard" {
		t.Errorf("HX-Redirect = %q, want /dashboard", rec.Header().Get("HX-Redirect"))
	}
}

// H-08: HTMX 회원가입 에러 → 422 + 부분 HTML
func TestHandleSignup_HTMX_Error(t *testing.T) {
	req := formRequest("POST", "/api/signup", map[string]string{
		"email": "", "password": "", "confirm_password": "",
	})
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	HandleSignup(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
}

// ── 로그인 처리 ──

// H-09: 로그인 성공
func TestHandleLogin_Success(t *testing.T) {
	email := testEmail("h-login-ok")
	defer cleanup(t, email)

	auth.Signup(context.Background(), email, "mypassword")

	req := formRequest("POST", "/api/login", map[string]string{
		"email": email, "password": "mypassword",
	})
	rec := httptest.NewRecorder()
	HandleLogin(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session_token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("session_token 쿠키 없음")
	}
}

// H-10: 로그인 실패 — 잘못된 비밀번호
func TestHandleLogin_WrongPassword(t *testing.T) {
	email := testEmail("h-login-wrongpw")
	defer cleanup(t, email)

	auth.Signup(context.Background(), email, "correct")

	req := formRequest("POST", "/api/login", map[string]string{
		"email": email, "password": "wrong",
	})
	rec := httptest.NewRecorder()
	HandleLogin(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "올바르지 않습니다") {
		t.Errorf("에러 메시지 없음: %s", body)
	}
}

// H-11: 로그인 — 빈 입력
func TestHandleLogin_Empty(t *testing.T) {
	req := formRequest("POST", "/api/login", map[string]string{
		"email": "", "password": "",
	})
	rec := httptest.NewRecorder()
	HandleLogin(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "입력해주세요") {
		t.Errorf("에러 메시지 없음: %s", body)
	}
}

// H-12: HTMX 로그인 에러 → 422
func TestHandleLogin_HTMX_Error(t *testing.T) {
	req := formRequest("POST", "/api/login", map[string]string{
		"email": "nope@test.local", "password": "wrong",
	})
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	HandleLogin(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
}

// ── 로그아웃 ──

// H-13: 로그아웃 → 쿠키 삭제 + 리다이렉트
func TestHandleLogout(t *testing.T) {
	token, _ := middleware.CreateSession(99, "logout@test.local")

	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rec := httptest.NewRecorder()
	HandleLogout(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}

	// 쿠키 만료 확인
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session_token" && c.MaxAge >= 0 {
			t.Errorf("session_token MaxAge = %d, want < 0", c.MaxAge)
		}
	}

	// 세션 삭제 확인
	if s := middleware.GetSession(token); s != nil {
		t.Error("세션이 삭제되지 않음")
	}
}

// H-14: HTMX 로그아웃 → HX-Redirect
func TestHandleLogout_HTMX(t *testing.T) {
	token, _ := middleware.CreateSession(100, "htmx-logout@test.local")

	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	HandleLogout(rec, req)

	if rec.Header().Get("HX-Redirect") != "/login" {
		t.Errorf("HX-Redirect = %q, want /login", rec.Header().Get("HX-Redirect"))
	}
}

// ── 대시보드 ──

// H-15: 대시보드 — 인증된 사용자
func TestShowDashboard_Authenticated(t *testing.T) {
	email := testEmail("h-dash-ok")
	defer cleanup(t, email)

	user, _ := auth.Signup(context.Background(), email, "password123")
	token, _ := middleware.CreateSession(user.ID, email)
	defer middleware.DestroySession(token)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	// context에 세션 주입 (미들웨어 통과 시뮬레이션)
	ctx := context.WithValue(req.Context(), middleware.SessionKey, &middleware.Session{
		UserID: user.ID, Email: email,
	})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	ShowDashboard(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), email) {
		t.Error("대시보드에 사용자 이메일 없음")
	}
}

// H-16: 대시보드 — 세션 없음 → 리다이렉트
func TestShowDashboard_NoSession(t *testing.T) {
	req := httptest.NewRequest("GET", "/dashboard", nil)
	rec := httptest.NewRecorder()
	ShowDashboard(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}
}
