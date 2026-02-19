# 개발 로드맵

## Phase 1: 기반 ✅
- [x] 프로젝트 구조 (Go standard layout)
- [x] PostgreSQL 연결 (pgx/v5 커넥션 풀)
- [x] 자동 마이그레이션
- [x] 회원가입 / 로그인 / 로그아웃
- [x] bcrypt 비밀번호 해싱
- [x] 세션 기반 인증 (메모리)
- [x] 인증 미들웨어
- [x] 기본 대시보드 레이아웃 (사이드바, 통계 카드)
- [x] 다크/라이트 모드 (Alpine.js + localStorage)
- [x] HTMX/Alpine.js 로컬 서빙
- [x] 글래스모피즘 UI

## Phase 2: 테스트 & 안정화 🔜
- [ ] 단위 테스트 (auth, middleware, models)
- [ ] 통합 테스트 (핸들러 + DB)
- [ ] 테스트 유틸리티 (testutil 패키지)
- [ ] 세션 만료 처리 (TTL)
- [ ] 입력 검증 강화 (이메일 형식 등)
- [ ] 에러 핸들링 통일
- [ ] 마이그레이션 도구 도입 (goose/migrate)

## Phase 3: 데이터셋
- [ ] CSV/Excel 파일 업로드
- [ ] 파일 파싱 및 메타데이터 추출
- [ ] 데이터셋 CRUD
- [ ] 데이터 미리보기 (테이블)
- [ ] 데이터 정렬/필터링

## Phase 4: 차트 & 시각화
- [ ] 차트 라이브러리 선정 (Chart.js / D3.js / ECharts)
- [ ] 기본 차트 타입 (bar, line, pie, scatter)
- [ ] 차트 설정 UI (축, 범례, 색상)
- [ ] HTMX 부분 교체로 차트 인터랙션

## Phase 5: 대시보드
- [ ] 대시보드 CRUD
- [ ] 드래그 앤 드롭 레이아웃
- [ ] 다중 차트 배치
- [ ] 대시보드 공유 (공개 링크)
- [ ] PDF/이미지 내보내기

## Phase 6: 고도화
- [ ] Redis 세션 저장소
- [ ] 사용자 프로필/설정
- [ ] 팀/조직 기능
- [ ] API 키 인증
- [ ] 실시간 데이터 업데이트 (SSE/WebSocket)
