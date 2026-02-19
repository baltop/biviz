# Changelog

## 2026-02-19

### AI 채팅 기능 추가

- 대시보드에 AI 채팅 위젯 추가 (플로팅 버튼 → 사이드 패널)
- SSE 스트리밍 방식으로 실시간 응답 표시
- 마크다운 렌더링 지원 (코드블록, 볼드)
- 대화 히스토리 유지 (세션 내)
- 다크/라이트 모드 완벽 대응

### AI 백엔드: Gemini → Anthropic 마이그레이션

- Google Gemini API → Anthropic Claude API (claude-sonnet-4-20250514)로 전환
- 공식 SDK 사용: `github.com/anthropics/anthropic-sdk-go v1.25.0`
- 환경변수: `GEMINI_API_KEY` → `ANTHROPIC_API_KEY`
- 시스템 프롬프트를 Anthropic 네이티브 `system` 파라미터로 전달 (Gemini의 fake user/model 턴 방식 제거)
- Google Cloud 관련 의존성 20+ 패키지 제거로 빌드 크기 감소

### CSS 수정

- AI 위젯에서 사용하는 누락된 Tailwind 유틸리티 클래스 추가
  - 사이즈: `w-14`, `h-14`, `w-6`, `h-6`, `w-12`, `h-12`, `w-2`, `h-2`, `w-1.5`, `h-1.5`
  - 위치: `bottom-6`, `right-6`, `bottom-2`, `right-2`, `bottom-0`
  - 레이아웃: `rounded-2xl`, `max-w-[85%]`, `max-w-[200px]`, `min-h-[300px]`
  - 패딩: `p-1`, `p-1.5`, `py-2.5`
  - 텍스트: `text-[10px]`, `leading-relaxed`
  - 애니메이션: `animate-bounce`, `animate-pulse`
  - 기타: `resize-none`, `bg-transparent`
- 다크/라이트 모드 CSS 커스텀 속성 추가 (`--primary`, `--card`, `--bg`, `--border`, `--fg` 등)

### 서버 라우팅

- `POST /api/ai/chat` 엔드포인트 추가 (인증 미들웨어 적용)
- `ai.Init()` / `ai.Close()` 서버 생명주기에 통합

### 변경 파일

| 파일 | 변경 |
|---|---|
| `internal/ai/gemini.go` | Gemini → Anthropic SDK 전면 교체 |
| `internal/handlers/ai.go` | AI 채팅 핸들러 (신규) |
| `cmd/server/main.go` | AI 초기화 + 라우트 추가 |
| `templates/pages/dashboard.html` | AI 채팅 위젯 UI + JS (Alpine.js) |
| `static/css/app.css` | 누락 유틸리티 클래스 + CSS 변수 |
| `go.mod` / `go.sum` | Anthropic SDK 추가, Gemini 제거 |
