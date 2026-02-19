package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"biviz/internal/db"
	"biviz/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists    = errors.New("이미 등록된 이메일입니다")
	ErrInvalidCreds   = errors.New("이메일 또는 비밀번호가 올바르지 않습니다")
	ErrUserNotFound   = errors.New("사용자를 찾을 수 없습니다")
)

// Signup 회원가입 처리
func Signup(ctx context.Context, email, password string) (*models.User, error) {
	// 비밀번호 해싱
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("비밀번호 해싱 실패: %w", err)
	}

	var user models.User
	err = db.Pool.QueryRow(ctx,
		`INSERT INTO users (email, password) VALUES ($1, $2)
		 RETURNING id, email, name, created_at, updated_at`,
		email, string(hashed),
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("회원가입 실패: %w", err)
	}

	return &user, nil
}

// Login 로그인 처리
func Login(ctx context.Context, email, password string) (*models.User, error) {
	var user models.User
	var hashedPassword string

	err := db.Pool.QueryRow(ctx,
		`SELECT id, email, password, name, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &hashedPassword, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, ErrInvalidCreds
	}

	// 비밀번호 검증
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return &user, nil
}

// GetUserByID ID로 사용자 조회
func GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := db.Pool.QueryRow(ctx,
		`SELECT id, email, name, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

// UpdateUser 사용자 정보 업데이트
func UpdateUser(ctx context.Context, id int, name string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE users SET name = $1, updated_at = $2 WHERE id = $3`,
		name, time.Now(), id,
	)
	return err
}
