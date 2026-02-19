package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB 전역 커넥션 풀
var Pool *pgxpool.Pool

// Connect PostgreSQL 연결 초기화
func Connect(databaseURL string) error {
	var err error
	Pool, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("DB 연결 실패: %w", err)
	}

	// 연결 테스트
	if err := Pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("DB ping 실패: %w", err)
	}

	log.Println("✅ PostgreSQL 연결 성공")
	return nil
}

// Migrate 테이블 생성
func Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id          SERIAL PRIMARY KEY,
		email       VARCHAR(255) UNIQUE NOT NULL,
		password    VARCHAR(255) NOT NULL,
		name        VARCHAR(100) DEFAULT '',
		created_at  TIMESTAMPTZ DEFAULT NOW(),
		updated_at  TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := Pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("마이그레이션 실패: %w", err)
	}

	log.Println("✅ 마이그레이션 완료")
	return nil
}

// Close DB 연결 종료
func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
