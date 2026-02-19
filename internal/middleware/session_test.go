package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// M-01: 세션 생성 후 조회
func TestCreateAndGetSession(t *testing.T) {
	token, err := CreateSession(1, "test@example.com")
	if err != nil {
		t.Fatalf("세션 생성 실패: %v", err)
	}

	session := GetSession(token)
	if session == nil {
		t.Fatal("세션이 nil")
	}
	if session.UserID != 1 {
		t.Errorf("UserID = %d, want 1", session.UserID)
	}
	if session.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", session.Email, "test@example.com")
	}

	// 정리
	DestroySession(token)
}

// M-02: 세션 삭제 후 조회 → nil
func TestDestroySession(t *testing.T) {
	token, _ := CreateSession(2, "del@example.com")
	DestroySession(token)

	if s := GetSession(token); s != nil {
		t.Error("삭제된 세션이 여전히 존재")
	}
}

// M-03: 존재하지 않는 토큰 조회 → nil
func TestGetSession_NotFound(t *testing.T) {
	if s := GetSession("nonexistent-token-12345"); s != nil {
		t.Error("없는 토큰에 대해 세션 반환됨")
	}
}

// M-04: 토큰 고유성
func TestCreateSession_UniqueTokens(t *testing.T) {
	t1, _ := CreateSession(1, "a@example.com")
	t2, _ := CreateSession(1, "a@example.com")
	if t1 == t2 {
		t.Error("동일한 토큰이 생성됨")
	}
	DestroySession(t1)
	DestroySession(t2)
}

// M-05: AuthMiddleware — 유효한 세션
func TestAuthMiddleware_ValidSession(t *testing.T) {
	token, _ := CreateSession(10, "valid@example.com")
	defer DestroySession(token)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		s := GetCurrentSession(r)
		if s == nil {
			t.Error("context에 세션 없음")
		}
		if s.UserID != 10 {
			t.Errorf("UserID = %d, want 10", s.UserID)
		}
	})

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rec := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(rec, req)

	if !called {
		t.Error("next 핸들러가 호출되지 않음")
	}
}

// M-06: AuthMiddleware — 쿠키 없음 → 리다이렉트
func TestAuthMiddleware_NoCookie(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("인증 없이 next 호출됨")
	})

	req := httptest.NewRequest("GET", "/dashboard", nil)
	rec := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/login" {
		t.Errorf("Location = %q, want /login", loc)
	}
}

// M-07: AuthMiddleware — 무효한 토큰 → 리다이렉트
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("무효 토큰으로 next 호출됨")
	})

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "bad-token"})
	rec := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rec.Code)
	}
}

// M-08: AuthMiddleware — HTMX 요청 + 세션 없음 → HX-Redirect
func TestAuthMiddleware_HTMX_NoSession(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("HTMX 미인증으로 next 호출됨")
	})

	req := httptest.NewRequest("GET", "/dashboard", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	AuthMiddleware(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	hxRedirect := rec.Header().Get("HX-Redirect")
	if hxRedirect != "/login" {
		t.Errorf("HX-Redirect = %q, want /login", hxRedirect)
	}
}
