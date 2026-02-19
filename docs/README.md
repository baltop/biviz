# BiViz 개발 문서

BI 시각화 솔루션 — Go + HTMX + Alpine.js + PostgreSQL

## 문서 목록

| 문서 | 설명 |
|------|------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | 시스템 아키텍처 및 기술 스택 |
| [API.md](API.md) | API 엔드포인트 명세 |
| [DATABASE.md](DATABASE.md) | 데이터베이스 스키마 및 마이그레이션 |
| [DEVELOPMENT.md](DEVELOPMENT.md) | 개발 환경 설정 및 실행 방법 |
| [TESTING.md](TESTING.md) | 테스트 원칙, 전략, 계획서 |
| [ROADMAP.md](ROADMAP.md) | 개발 로드맵 및 향후 계획 |

## 빠른 시작

```bash
cd works/biviz
go build -o biviz ./cmd/server
./biviz
# → http://localhost:8080
```

필수 요건: Go 1.25+, PostgreSQL 16 (Docker)
