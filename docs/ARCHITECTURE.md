# 아키텍처

## 기술 스택

| 레이어 | 기술 | 버전 | 역할 |
|--------|------|------|------|
| 서버 | Go (net/http) | 1.25 | HTTP 서버, 라우팅, 비즈니스 로직 |
| 뷰 | html/template | stdlib | 서버 사이드 렌더링 (SSR) |
| 인터랙션 | HTMX | 2.0 | AJAX 요청, 부분 페이지 교체 |
| UI 반응성 | Alpine.js | 3.14 | 클라이언트 사이드 상태 관리 |
| DB | PostgreSQL | 16 | 데이터 저장 |
| DB 드라이버 | pgx/v5 | 5.8 | 커넥션 풀, 쿼리 실행 |
| AI | Anthropic Claude | claude-sonnet-4-20250514 | AI 채팅 어시스턴트 |
| 인증 | bcrypt | — | 비밀번호 해싱 |

## 아키텍처 원칙

### 하이퍼미디어 중심 (Hypermedia-Driven)
- 서버가 HTML을 반환하고, HTMX가 DOM을 부분 교체
- JSON API 대신 HTML 응답 → 프론트엔드 프레임워크 불필요
- 클라이언트 상태 최소화

### 점진적 향상 (Progressive Enhancement)
- HTMX 없이도 기본 동작 (폼 서브밋 + 리다이렉트)
- Alpine.js는 UI 편의 기능에만 사용 (다크모드 토글, 드롭다운 등)

### 서버 중심 상태 관리
- 세션은 서버 메모리에 저장 (추후 Redis 전환 가능)
- 클라이언트는 쿠키 토큰만 보유

## 프로젝트 구조

```
works/biviz/
├── cmd/
│   └── server/
│       └── main.go              # 엔트리포인트, 라우터 등록
├── internal/
│   ├── auth/
│   │   └── auth.go              # 회원가입/로그인/사용자 조회
│   ├── db/
│   │   └── db.go                # PostgreSQL 연결, 마이그레이션
│   ├── handlers/
│   │   ├── ai.go                # AI 채팅 핸들러 (SSE 스트리밍)
│   │   ├── auth.go              # 인증 관련 핸들러
│   │   ├── dashboard.go         # 대시보드 핸들러
│   │   └── interview.go         # 인터뷰 업로드 핸들러
│   ├── middleware/
│   │   └── session.go           # 세션 관리, 인증 미들웨어
│   └── models/
│       └── user.go              # 사용자 모델
├── static/
│   ├── css/
│   │   └── app.css              # 글로벌 스타일 (Tailwind 포함)
│   └── js/
│       ├── app.js               # 클라이언트 JS
│       ├── htmx.min.js          # HTMX (로컬 서빙)
│       └── alpine.min.js        # Alpine.js (로컬 서빙)
├── templates/
│   ├── layouts/
│   │   └── base.html            # 공통 레이아웃
│   ├── pages/
│   │   ├── login.html           # 로그인 페이지
│   │   ├── signup.html          # 회원가입 페이지
│   │   ├── dashboard.html       # 대시보드 페이지
│   │   └── interview-upload.html # 인터뷰 업로드 페이지
│   └── components/
│       └── error-message.html   # 에러 메시지 컴포넌트
├── docs/                        # 개발 문서 (현재 디렉토리)
├── go.mod
├── go.sum
└── biviz                        # 빌드된 바이너리
```

## 요청 흐름

```
[브라우저] → GET /dashboard
  → [net/http 라우터]
  → [AuthMiddleware] → 쿠키에서 세션 토큰 확인
  → [ShowDashboard 핸들러]
  → [html/template 렌더링] → base.html + dashboard.html
  → HTML 응답

[브라우저] → POST /api/login (HTMX hx-post)
  → [HandleLogin 핸들러]
  → [auth.Login] → DB 조회 + bcrypt 검증
  → 세션 생성, 쿠키 설정
  → HX-Redirect: /dashboard
```

## Go 템플릿 설계 결정

**문제**: `html/template`에서 같은 이름의 `{{define "content"}}`가 여러 파일에 존재하면 마지막 파싱된 것만 유지됨.

**해결**: 페이지별 독립 `template.Template` 인스턴스 생성

```go
// ❌ 잘못된 방식 — 모든 페이지를 하나의 템플릿에 파싱
tmpl := template.Must(template.ParseGlob("templates/**/*.html"))

// ✅ 올바른 방식 — 페이지별 독립 인스턴스
PageTemplates = map[string]*template.Template{
    "login.html": template.Must(template.ParseFiles(
        "templates/layouts/base.html",
        "templates/pages/login.html",
    )),
}
```

## HTMX + Alpine.js 로컬 서빙

CDN 대신 로컬 파일을 사용하는 이유:
- SSH 포트포워드 환경에서 CDN 로드 실패 가능
- 네트워크 의존성 제거
- 오프라인 개발 지원
