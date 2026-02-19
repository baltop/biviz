# API 엔드포인트

## 라우팅 규칙

- **페이지 렌더링**: `GET /path` → HTML 전체 페이지 반환
- **폼 처리**: `POST /api/path` → HTMX 부분 응답 또는 리다이렉트
- **정적 파일**: `GET /static/*` → CSS, JS, 이미지

## 엔드포인트 목록

### 페이지

| 메서드 | 경로 | 인증 | 설명 |
|--------|------|------|------|
| GET | `/` | ✗ | `/login`으로 리다이렉트 |
| GET | `/login` | ✗ | 로그인 페이지 (로그인 상태면 `/dashboard`로) |
| GET | `/signup` | ✗ | 회원가입 페이지 |
| GET | `/dashboard` | ✓ | 대시보드 (AuthMiddleware) |
| GET | `/interview-upload` | ✓ | 인터뷰 업로드 페이지 |

### API

| 메서드 | 경로 | 인증 | 설명 |
|--------|------|------|------|
| POST | `/api/login` | ✗ | 로그인 처리 |
| POST | `/api/signup` | ✗ | 회원가입 처리 |
| POST | `/api/logout` | ✗ | 로그아웃 (세션 삭제) |
| POST | `/api/ai/chat` | ✓ | AI 채팅 (SSE 스트리밍) |
| POST | `/api/interview/upload` | ✓ | 인터뷰 파일 업로드 (multipart/form-data) |

## 요청/응답 상세

### POST /api/signup

**요청** (form-urlencoded):
```
email=user@example.com&password=12345678&confirm_password=12345678
```

**성공 응답**:
- 쿠키 `session_token` 설정
- HTMX: `HX-Redirect: /dashboard` (204)
- 일반: `303 See Other → /dashboard`

**에러 응답** (422 Unprocessable Entity):
```html
<div class="... bg-red-500/10 ...">이미 등록된 이메일입니다</div>
```

### POST /api/login

**요청** (form-urlencoded):
```
email=user@example.com&password=12345678
```

**성공/에러**: signup과 동일한 패턴

### POST /api/logout

**동작**: 세션 삭제 + 쿠키 만료 + `/login`으로 리다이렉트

### POST /api/ai/chat

**요청** (JSON):
```json
{
  "message": "매출 데이터를 분석해줘",
  "history": [
    {"role": "user", "content": "이전 질문"},
    {"role": "assistant", "content": "이전 답변"}
  ]
}
```

**응답** (SSE text/event-stream):
```
data: {"text":"안녕"}

data: {"text":"하세요"}

data: [DONE]
```

**에러 응답**:
- 401: 미인증 `{"error":"Unauthorized"}`
- 503: AI 미설정 `{"error":"AI service is not configured"}`
- 400: 잘못된 요청 `{"error":"Invalid request body"}`

**모델**: Claude claude-sonnet-4-20250514 (Anthropic)
**환경변수**: `ANTHROPIC_API_KEY` 필요

### GET /interview-upload

인터뷰 업로드 페이지를 렌더링합니다. 업로드된 파일 목록도 함께 표시됩니다.

### POST /api/interview/upload

**요청** (multipart/form-data):
```
files: (바이너리 파일, 복수 가능)
```

**지원 형식**: TXT, CSV, PDF, DOC, DOCX, MP3, MP4, WAV, M4A, WEBM
**최대 크기**: 50MB

**성공 응답**: `200 OK`
**에러 응답**:
- 401: 미인증 `"인증이 필요합니다"`
- 400: 파일 없음 `"파일을 선택해주세요"` / 크기 초과 `"파일이 너무 큽니다 (최대 50MB)"`

**저장 경로**: `uploads/interviews/{userID}/{filename}`

## HTMX 통합 패턴

모든 POST 핸들러는 `HX-Request` 헤더를 확인하여 응답 분기:

```
if HX-Request == "true":
    에러 → 422 + 부분 HTML (hx-target에 삽입)
    성공 → HX-Redirect 헤더
else:
    에러 → 전체 페이지 재렌더링 (에러 메시지 포함)
    성공 → 303 리다이렉트
```
