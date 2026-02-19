package auth

import (
	"context"
	"fmt"
	"os"
	"testing"

	"biviz/internal/db"
)

// 테스트 DB 셋업
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

	code := m.Run()

	// 테스트 데이터 정리
	db.Pool.Exec(context.Background(), "DELETE FROM users WHERE email LIKE '%@test.biviz.local'")

	os.Exit(code)
}

// 테스트용 고유 이메일 생성
func testEmail(name string) string {
	return fmt.Sprintf("%s@test.biviz.local", name)
}

// 테스트 후 사용자 삭제
func cleanup(t *testing.T, email string) {
	t.Helper()
	db.Pool.Exec(context.Background(), "DELETE FROM users WHERE email = $1", email)
}

// A-01: 회원가입 성공
func TestSignup_Success(t *testing.T) {
	email := testEmail("signup-ok")
	defer cleanup(t, email)

	user, err := Signup(context.Background(), email, "password123")
	if err != nil {
		t.Fatalf("회원가입 실패: %v", err)
	}
	if user.ID == 0 {
		t.Error("User ID가 0")
	}
	if user.Email != email {
		t.Errorf("Email = %q, want %q", user.Email, email)
	}
}

// A-02: 중복 이메일 가입 → ErrEmailExists
func TestSignup_DuplicateEmail(t *testing.T) {
	email := testEmail("signup-dup")
	defer cleanup(t, email)

	_, err := Signup(context.Background(), email, "password123")
	if err != nil {
		t.Fatalf("첫 가입 실패: %v", err)
	}

	_, err = Signup(context.Background(), email, "other-password")
	if err != ErrEmailExists {
		t.Errorf("err = %v, want ErrEmailExists", err)
	}
}

// A-03: 로그인 성공
func TestLogin_Success(t *testing.T) {
	email := testEmail("login-ok")
	defer cleanup(t, email)

	_, err := Signup(context.Background(), email, "mypassword")
	if err != nil {
		t.Fatalf("가입 실패: %v", err)
	}

	user, err := Login(context.Background(), email, "mypassword")
	if err != nil {
		t.Fatalf("로그인 실패: %v", err)
	}
	if user.Email != email {
		t.Errorf("Email = %q, want %q", user.Email, email)
	}
}

// A-04: 로그인 실패 — 잘못된 비밀번호
func TestLogin_WrongPassword(t *testing.T) {
	email := testEmail("login-wrongpw")
	defer cleanup(t, email)

	Signup(context.Background(), email, "correct-pw")

	_, err := Login(context.Background(), email, "wrong-pw")
	if err != ErrInvalidCreds {
		t.Errorf("err = %v, want ErrInvalidCreds", err)
	}
}

// A-05: 로그인 실패 — 존재하지 않는 이메일
func TestLogin_NonexistentEmail(t *testing.T) {
	_, err := Login(context.Background(), "nobody@test.biviz.local", "password")
	if err != ErrInvalidCreds {
		t.Errorf("err = %v, want ErrInvalidCreds", err)
	}
}

// A-06: 사용자 ID로 조회
func TestGetUserByID_Success(t *testing.T) {
	email := testEmail("getuser-ok")
	defer cleanup(t, email)

	created, _ := Signup(context.Background(), email, "pw123456")

	user, err := GetUserByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("조회 실패: %v", err)
	}
	if user.Email != email {
		t.Errorf("Email = %q, want %q", user.Email, email)
	}
}

// A-07: 존재하지 않는 ID 조회 → ErrUserNotFound
func TestGetUserByID_NotFound(t *testing.T) {
	_, err := GetUserByID(context.Background(), 999999)
	if err != ErrUserNotFound {
		t.Errorf("err = %v, want ErrUserNotFound", err)
	}
}

// A-08: 사용자 업데이트
func TestUpdateUser(t *testing.T) {
	email := testEmail("update-ok")
	defer cleanup(t, email)

	user, _ := Signup(context.Background(), email, "pw123456")

	err := UpdateUser(context.Background(), user.ID, "홍길동")
	if err != nil {
		t.Fatalf("업데이트 실패: %v", err)
	}

	updated, _ := GetUserByID(context.Background(), user.ID)
	if updated.Name != "홍길동" {
		t.Errorf("Name = %q, want 홍길동", updated.Name)
	}
}

// A-09: bcrypt 해시 검증 — 원문 비밀번호가 DB에 저장되지 않음
func TestSignup_PasswordHashed(t *testing.T) {
	email := testEmail("hash-check")
	defer cleanup(t, email)

	Signup(context.Background(), email, "plaintext123")

	var stored string
	db.Pool.QueryRow(context.Background(),
		"SELECT password FROM users WHERE email = $1", email,
	).Scan(&stored)

	if stored == "plaintext123" {
		t.Error("비밀번호가 평문으로 저장됨")
	}
	if len(stored) < 50 {
		t.Errorf("해시 길이가 너무 짧음: %d", len(stored))
	}
}
