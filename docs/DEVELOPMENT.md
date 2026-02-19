# 개발 환경 설정

## 요구 사항

- Go 1.25+
- Docker (PostgreSQL 컨테이너)
- Node.js 22+ (선택, 프론트엔드 빌드 도구 사용 시)

## PostgreSQL 시작

```bash
cd docker/postgres
docker compose up -d
```

컨테이너 상태 확인:
```bash
docker ps | grep dev-postgres
```

## 빌드 및 실행

```bash
cd works/biviz

# 빌드
go build -o biviz ./cmd/server

# 실행
./biviz
# → http://localhost:8080

# 환경변수로 설정 변경
PORT=9090 DATABASE_URL=postgresql://... ./biviz
```

## 개발 중 자동 재시작

[air](https://github.com/air-verse/air) 사용 권장:

```bash
go install github.com/air-verse/air@latest

# 프로젝트 루트에서
air
```

`.air.toml` 예시:
```toml
[build]
  cmd = "go build -o ./tmp/biviz ./cmd/server"
  bin = "./tmp/biviz"
  include_ext = ["go", "html", "css", "js"]
  exclude_dir = ["tmp", "docs"]
```

## 환경 변수

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `PORT` | `8080` | 서버 포트 |
| `DATABASE_URL` | `postgresql://dev:devpass@localhost:5432/devdb` | DB 연결 문자열 |

## SSH 포트포워드 접속

```bash
ssh -L 8080:localhost:8080 user@server
# 브라우저에서 http://localhost:8080
```

HTMX, Alpine.js는 로컬 파일로 서빙하므로 CDN 접근 불필요.

## 디렉토리 규칙

- `cmd/` — 실행 바이너리 엔트리포인트
- `internal/` — 외부 import 불가 패키지 (Go 규약)
- `static/` — 정적 파일 (CSS, JS, 이미지)
- `templates/` — Go html/template 파일
- `docs/` — 개발 문서
