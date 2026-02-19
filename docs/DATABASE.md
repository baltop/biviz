# 데이터베이스

## 연결 정보

| 항목 | 값 |
|------|------|
| DBMS | PostgreSQL 16 |
| 호스트 | localhost:5432 |
| 데이터베이스 | devdb |
| 사용자 | dev |
| 비밀번호 | devpass |
| 인코딩 | UTF-8 |
| 컨테이너 | dev-postgres (Docker) |

```
DATABASE_URL=postgresql://dev:devpass@localhost:5432/devdb
```

## 스키마

### users

```sql
CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,        -- bcrypt 해시
    name        VARCHAR(100) DEFAULT '',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

## 마이그레이션

현재 자동 마이그레이션 방식 사용 (`db.Migrate()` → `CREATE TABLE IF NOT EXISTS`).

향후 마이그레이션 도구 도입 계획:
- [golang-migrate](https://github.com/golang-migrate/migrate) 또는
- [goose](https://github.com/pressly/goose)

## 커넥션 관리

- `pgxpool.Pool` 사용 (커넥션 풀)
- 앱 시작 시 `Connect()` → 종료 시 `Close()`
- `context.Background()` 기반 쿼리 실행

## 향후 테이블 (예정)

```sql
-- 데이터셋
CREATE TABLE datasets (
    id          SERIAL PRIMARY KEY,
    user_id     INT REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    file_path   VARCHAR(500),
    row_count   INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 대시보드
CREATE TABLE dashboards (
    id          SERIAL PRIMARY KEY,
    user_id     INT REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    config      JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- 차트
CREATE TABLE charts (
    id           SERIAL PRIMARY KEY,
    dashboard_id INT REFERENCES dashboards(id) ON DELETE CASCADE,
    dataset_id   INT REFERENCES datasets(id) ON DELETE SET NULL,
    type         VARCHAR(50) NOT NULL,   -- bar, line, pie, scatter, ...
    config       JSONB DEFAULT '{}',
    position     INT DEFAULT 0,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);
```
