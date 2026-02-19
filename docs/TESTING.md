# 테스트 전략

## 테스트 원칙

### 1. 테스트 피라미드
```
        /  E2E  \          ← 적게, 핵심 시나리오만
       / 통합 테스트 \       ← DB·핸들러 연동
      / 단위 테스트     \    ← 많이, 빠르게
```

- **단위 테스트 (Unit)**: 함수·메서드 단위 격리 테스트. 외부 의존성 모킹.
- **통합 테스트 (Integration)**: 실제 DB 연결, HTTP 핸들러 호출, 미들웨어 체인 검증.
- **E2E 테스트 (End-to-End)**: 브라우저 자동화로 사용자 시나리오 검증.

### 2. 핵심 규칙

1. **새 기능 = 테스트 먼저** (TDD 지향)
   - Red → Green → Refactor 사이클
   - 실패하는 테스트 작성 → 최소 구현 → 리팩토링

2. **버그 = 재현 테스트 먼저**
   - 버그 보고 → 실패하는 테스트 작성 → 수정 → 테스트 통과 확인

3. **커버리지 목표**: 비즈니스 로직 80%+, 핸들러 70%+

4. **테스트는 독립적**
   - 순서 의존성 없음
   - 각 테스트가 자체 데이터 셋업/클린업
   - `t.Parallel()` 가능한 곳은 병렬 실행

5. **빠른 피드백**
   - 단위 테스트: < 1초
   - 통합 테스트: < 10초
   - E2E: < 60초

---

## TDD 워크플로우

```
1. 요구사항 정의
    ↓
2. 실패하는 테스트 작성 (Red)
    ↓
3. 최소한의 코드로 통과 (Green)
    ↓
4. 리팩토링 (Refactor)
    ↓
5. 반복
```

### 예시: 새 기능 "데이터셋 업로드"

```go
// 1단계: 실패하는 테스트 먼저
func TestUploadDataset_CSVFile(t *testing.T) {
    // CSV 파일 생성
    body := createMultipartCSV(t, "test.csv", "name,age\n홍길동,30\n")

    // 업로드 요청
    req := httptest.NewRequest("POST", "/api/datasets/upload", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    rec := httptest.NewRecorder()

    handler.HandleUpload(rec, req)

    // 검증
    assert.Equal(t, 200, rec.Code)
    // DB에 데이터셋 레코드 생성 확인
    // 파일 저장 확인
}

// 2단계: 테스트를 통과시키는 최소 구현
// 3단계: 리팩토링 (에러 처리, 검증 로직 분리 등)
```

---

## 테스트 계획서

### Phase 1: 인증 모듈 (현재 구현 완료 → 테스트 추가)

#### 단위 테스트 (`internal/auth/auth_test.go`)

| ID | 테스트 | 입력 | 기대 결과 |
|----|--------|------|-----------|
| A-01 | 회원가입 성공 | 유효한 이메일+비밀번호 | User 객체 반환, DB에 저장 |
| A-02 | 중복 이메일 가입 | 이미 존재하는 이메일 | `ErrEmailExists` 반환 |
| A-03 | 로그인 성공 | 올바른 이메일+비밀번호 | User 객체 반환 |
| A-04 | 로그인 실패 — 잘못된 비밀번호 | 올바른 이메일+틀린 비밀번호 | `ErrInvalidCreds` 반환 |
| A-05 | 로그인 실패 — 미존재 이메일 | 없는 이메일 | `ErrInvalidCreds` 반환 |
| A-06 | 사용자 조회 | 유효한 ID | User 객체 반환 |
| A-07 | 사용자 조회 실패 | 없는 ID | `ErrUserNotFound` 반환 |

#### 미들웨어 테스트 (`internal/middleware/session_test.go`)

| ID | 테스트 | 시나리오 | 기대 결과 |
|----|--------|----------|-----------|
| M-01 | 세션 생성/조회 | CreateSession → GetSession | 동일 세션 반환 |
| M-02 | 세션 삭제 | DestroySession → GetSession | nil 반환 |
| M-03 | 인증 미들웨어 — 유효 세션 | 유효한 쿠키 | next 핸들러 호출, context에 세션 |
| M-04 | 인증 미들웨어 — 세션 없음 | 쿠키 없음 | `/login`으로 리다이렉트 |
| M-05 | 인증 미들웨어 — 만료 세션 | 무효한 토큰 | `/login`으로 리다이렉트 |
| M-06 | 인증 미들웨어 — HTMX 요청 | 세션 없음 + HX-Request | `HX-Redirect` 헤더 |

#### 핸들러 통합 테스트 (`internal/handlers/auth_test.go`)

| ID | 테스트 | 시나리오 | 기대 결과 |
|----|--------|----------|-----------|
| H-01 | 로그인 페이지 | GET /login | 200, HTML 렌더링 |
| H-02 | 로그인 처리 성공 | POST /api/login (올바른 자격) | 세션 쿠키 + 리다이렉트 |
| H-03 | 로그인 처리 실패 | POST /api/login (틀린 비밀번호) | 에러 메시지 |
| H-04 | 회원가입 처리 성공 | POST /api/signup (유효한 입력) | 세션 쿠키 + 리다이렉트 |
| H-05 | 회원가입 — 비밀번호 짧음 | POST /api/signup (7자) | "8자 이상" 에러 |
| H-06 | 회원가입 — 비밀번호 불일치 | POST /api/signup (confirm ≠ password) | "일치하지 않음" 에러 |
| H-07 | HTMX 로그인 에러 | POST /api/login + HX-Request | 422 + 부분 HTML |
| H-08 | 로그아웃 | POST /api/logout | 쿠키 삭제 + 리다이렉트 |
| H-09 | 대시보드 접근 | GET /dashboard (인증됨) | 200, 대시보드 HTML |
| H-10 | 대시보드 미인증 | GET /dashboard (쿠키 없음) | `/login` 리다이렉트 |

### Phase 2: 데이터셋 모듈 (향후)

| ID | 테스트 | 설명 |
|----|--------|------|
| D-01 | CSV 업로드 | CSV 파일 파싱 및 DB 저장 |
| D-02 | Excel 업로드 | XLSX 파일 처리 |
| D-03 | 파일 크기 제한 | 초과 시 에러 반환 |
| D-04 | 잘못된 형식 | 파싱 불가 파일 거부 |
| D-05 | 데이터셋 목록 | 사용자별 데이터셋 조회 |
| D-06 | 데이터셋 삭제 | 파일 + DB 레코드 삭제 |

### Phase 3: 대시보드/차트 모듈 (향후)

| ID | 테스트 | 설명 |
|----|--------|------|
| C-01 | 대시보드 CRUD | 생성/조회/수정/삭제 |
| C-02 | 차트 생성 | 데이터셋 연결 + 차트 설정 |
| C-03 | 차트 렌더링 | 올바른 차트 데이터 반환 |
| C-04 | 대시보드 공유 | 공유 링크 생성/접근 |

### Phase 4: E2E 테스트 (향후)

| ID | 시나리오 | 설명 |
|----|----------|------|
| E-01 | 가입 → 로그인 → 대시보드 | 전체 인증 플로우 |
| E-02 | 데이터 업로드 → 차트 생성 | 핵심 BI 워크플로우 |
| E-03 | 다크/라이트 모드 전환 | UI 상태 유지 확인 |

---

## 테스트 실행

```bash
# 전체 테스트
go test ./...

# 특정 패키지
go test ./internal/auth/

# 상세 출력
go test -v ./internal/auth/

# 커버리지
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 특정 테스트만
go test -run TestSignup ./internal/auth/

# 병렬 실행
go test -parallel 4 ./...
```

## 테스트 유틸리티

### 테스트 DB 헬퍼 (`internal/testutil/db.go` 예정)

```go
package testutil

// SetupTestDB 테스트용 DB 연결 및 트랜잭션 시작
// 각 테스트마다 트랜잭션으로 감싸고 끝나면 롤백 → 클린 상태 유지
func SetupTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    // 테스트용 DB 또는 트랜잭션 기반 격리
}

// CleanupTestDB 테스트 후 정리
func CleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
    t.Helper()
    pool.Exec(context.Background(), "DELETE FROM users")
}
```

### HTTP 테스트 헬퍼

```go
// MakeRequest 테스트용 HTTP 요청 생성 (HTMX 헤더 포함)
func MakeHTMXRequest(method, path string, body io.Reader) *http.Request {
    req := httptest.NewRequest(method, path, body)
    req.Header.Set("HX-Request", "true")
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    return req
}
```

---

## CI 통합 (향후)

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports: ["5432:5432"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - run: go test -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out
```
