# K8s CMP (Cloud Management Platform)

## 프로젝트 개요

OpenStack 기반 인프라 환경에서 CAPI(Cluster API) + CAPO(Cluster API Provider for OpenStack)를 활용하여
Kubernetes 클러스터를 선언적으로 생성·관리·삭제할 수 있는 웹 기반 Cloud Management Platform.

**상세 기능 명세, DB 스키마, API 엔드포인트, UI 설계는 반드시 `SPEC.md`를 참조할 것.**

---

## 기술 스택

### Backend
- **언어:** Go 1.22+
- **웹 프레임워크:** Gin
- **ORM:** GORM (PostgreSQL)
- **K8s 클라이언트:** client-go, sigs.k8s.io/cluster-api
- **OpenStack SDK:** gophercloud
- **인증:** golang-jwt/jwt (JWT Access + Refresh Token)
- **WebSocket:** gorilla/websocket
- **설정 관리:** viper (config.yaml)
- **로거:** zap (구조화된 JSON 로깅)
- **메시지 큐:** NATS JetStream

### Frontend
- **프레임워크:** Next.js 14+ (App Router)
- **언어:** TypeScript (strict mode)
- **UI 컴포넌트:** shadcn/ui + Tailwind CSS
- **상태 관리:** Zustand (클라이언트) + TanStack Query (서버 상태)
- **폼:** React Hook Form + Zod (유효성 검증)
- **차트:** Recharts
- **실시간 통신:** Socket.IO client
- **터미널:** xterm.js

### 데이터베이스
- **PostgreSQL 16:** 주 데이터 저장소 (사용자, 클러스터, 테넌트, 감사로그)
- **Redis 7:** 캐시, 세션, 분산 락
- **NATS JetStream:** 비동기 작업 큐, 이벤트 스트림

---

## 디렉토리 구조

```
k8s-cmp/
├── cmd/                        # 엔트리포인트
│   ├── api-server/             # API 서버 main.go
│   ├── worker/                 # 비동기 작업 워커
│   └── migrator/               # DB 마이그레이션
├── internal/                   # 비공개 패키지
│   ├── domain/                 # 도메인 모델 (Cluster, User, Tenant)
│   ├── service/                # 비즈니스 로직 계층
│   ├── repository/             # DB/외부 저장소 접근
│   ├── handler/                # HTTP 핸들러 (Controller)
│   ├── middleware/             # 인증, 로깅, CORS 미들웨어
│   ├── k8s/                    # CAPI/CAPO 연동 로직
│   │   ├── manifest_generator.go   # CAPI 매니페스트 자동 생성
│   │   ├── cluster_manager.go      # 클러스터 CRUD (Apply/Delete)
│   │   └── watcher.go              # Informer 기반 상태 감시
│   ├── openstack/              # OpenStack API 연동
│   └── websocket/              # 실시간 상태 스트림
├── pkg/                        # 공개 유틸리티
│   ├── errors/                 # 커스텀 에러 타입
│   ├── response/               # 표준 API 응답 래퍼
│   └── crypto/                 # 암호화 유틸 (AES-256-GCM)
├── web/                        # Next.js 프론트엔드
│   ├── src/app/                # App Router 페이지
│   ├── src/components/         # UI 컴포넌트
│   │   ├── ui/                 # shadcn/ui 기본 컴포넌트
│   │   ├── clusters/           # 클러스터 관련 컴포넌트
│   │   ├── dashboard/          # 대시보드 컴포넌트
│   │   └── layout/             # 레이아웃 (Sidebar, Header)
│   ├── src/hooks/              # Custom Hooks
│   ├── src/lib/                # API 클라이언트, 유틸
│   └── src/types/              # TypeScript 타입 정의
├── deploy/                     # Helm charts, Dockerfile
├── migrations/                 # SQL 마이그레이션 파일
├── configs/                    # config.yaml 샘플
├── docs/                       # API 문서, 설계 문서
├── CLAUDE.md                   # (이 파일)
├── SPEC.md                     # 상세 프로그램 명세서
└── Makefile                    # 빌드/테스트 명령
```

---

## 코딩 규칙

### Go Backend
- API 응답 형식: `{ "success": bool, "data": {}, "error": { "code": string, "message": string } }`
- DB 필드명: `snake_case` / Go 구조체: `CamelCase`
- 모든 API 경로: `/api/v1/` 프리픽스
- 에러 처리: `pkg/errors`의 커스텀 에러 타입 사용
- 테스트 파일: `*_test.go`, 테이블 드리븐 방식
- 컨텍스트: 모든 서비스/레포지토리 메서드의 첫 번째 인자는 `context.Context`
- 로깅: `zap.Logger`를 DI로 주입, 구조화된 필드 사용
- OpenStack 자격증명: 평문 저장 금지, AES-256-GCM 암호화 필수

### Frontend
- 컴포넌트: 함수형 컴포넌트 + Hooks만 사용
- 스타일: Tailwind CSS 유틸리티 클래스 우선 (인라인 style 지양)
- API 호출: TanStack Query의 `useQuery`/`useMutation` 사용
- 폼 검증: Zod 스키마 정의 → React Hook Form에 연결
- 페이지 파일: `web/src/app/[route]/page.tsx`
- 타입: API 응답 타입은 `web/src/types/`에 정의

### Git 컨벤션
- 커밋 메시지: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:` 접두사 사용
- 브랜치: `feature/`, `fix/`, `refactor/` 접두사

---

## 개발 순서 (Claude Code 지시 순서)

아래 순서대로 단계별로 지시합니다. 각 단계의 상세 요구사항은 SPEC.md를 참조하세요.

1. **프로젝트 초기화** → SPEC.md §3 참조
2. **DB 스키마 및 마이그레이션** → SPEC.md §5 참조
3. **인증/인가 시스템** → SPEC.md §4.1 참조
4. **테넌트 관리 CRUD** → SPEC.md §4.2 참조
5. **OpenStack 리소스 연동** → SPEC.md §4.4 참조
6. **CAPI/CAPO 매니페스트 생성 엔진** → SPEC.md §4.3.1~4.3.2 참조
7. **클러스터 라이프사이클 API** → SPEC.md §4.3.3~4.3.4 참조
8. **모니터링/WebSocket 실시간 스트림** → SPEC.md §4.5 참조
9. **Frontend UI 개발** → SPEC.md §6 참조
10. **테스트 및 배포** → SPEC.md §7.4, §8 참조

---

## 주요 참고사항

- Management Cluster가 이미 존재한다고 가정 (CAPI/CAPO 컨트롤러 설치 완료)
- CMP는 Management Cluster의 kubeconfig를 사용하여 CAPI 리소스를 관리
- Workload Cluster의 kubeconfig는 CAPI가 자동 생성한 Secret에서 추출
- OpenStack 자격증명은 테넌트별로 관리되며, Vault 또는 DB 암호화 저장
