package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// 세션 저장소 (메모리 기반, 프로덕션에서는 Redis 권장)
type Session struct {
	UserID    int
	Email     string
	CreatedAt time.Time
}

var (
	sessions = make(map[string]*Session)
	mu       sync.RWMutex
)

type contextKey string

const SessionKey contextKey = "session"

// GenerateToken 랜덤 세션 토큰 생성
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession 세션 생성
func CreateSession(userID int, email string) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	mu.Lock()
	sessions[token] = &Session{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
	}
	mu.Unlock()

	return token, nil
}

// GetSession 세션 조회
func GetSession(token string) *Session {
	mu.RLock()
	defer mu.RUnlock()
	return sessions[token]
}

// DestroySession 세션 삭제
func DestroySession(token string) {
	mu.Lock()
	delete(sessions, token)
	mu.Unlock()
}

// AuthMiddleware 인증 미들웨어
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// HTMX 요청이면 로그인 페이지로 리다이렉트 헤더
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session := GetSession(cookie.Value)
		if session == nil {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// 세션 정보를 context에 저장
		ctx := context.WithValue(r.Context(), SessionKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCurrentSession 현재 요청의 세션 가져오기
func GetCurrentSession(r *http.Request) *Session {
	session, _ := r.Context().Value(SessionKey).(*Session)
	return session
}
