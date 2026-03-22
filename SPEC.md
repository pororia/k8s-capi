# K8s CMP 프로그램 명세서

> **버전:** 1.0 | **작성일:** 2026년 3월 21일
> OpenStack + CAPI/CAPO 기반 Kubernetes 클러스터 통합 관리 웹 애플리케이션

---

## 1. 프로젝트 개요

### 1.1 프로젝트 목적

OpenStack 기반 인프라 환경에서 Cluster API(CAPI)와 Cluster API Provider for OpenStack(CAPO)를 활용하여 Kubernetes 클러스터를 선언적(Declarative)으로 생성, 관리, 삭제할 수 있는 웹 기반 Cloud Management Platform을 개발한다.

**주요 목표:**
- GUI 기반으로 CAPI/CAPO 매니페스트를 자동 생성하여 Kubernetes 클러스터 프로비저닝
- Multi-Cluster 환경의 통합 모니터링 및 라이프사이클 관리
- 운영자/개발자 대상의 직관적인 대시보드 및 운영 자동화
- 멀티 테넌트 지원으로 조직별 클러스터 격리 및 권한 관리

### 1.2 시스템 범위

본 CMP가 커버하는 기능 영역:

- Management Cluster에 대한 CAPI/CAPO 컨트롤러 설치 및 부트스트랩 지원
- Workload Cluster 생성/수정/삭제 (CRUD)
- Cluster Scaling (Control Plane / Worker Node)
- Cluster Upgrade (Kubernetes 버전 업그레이드)
- OpenStack 리소스 연동 (Flavor, Image, Network, Subnet, Security Group)
- 클러스터 상태 모니터링 및 이벤트/알림
- 사용자 인증/인가 (RBAC)
- Kubeconfig 발급 및 다운로드
- Audit Log 및 작업 이력 추적

### 1.3 기술 아키텍처 개요

CMP는 다음과 같은 계층 구조를 가진다:

| 계층 | 구성요소 | 역할 |
|------|----------|------|
| Presentation | React (Next.js) | SPA 기반 웹 UI, 대시보드, 폼 |
| API Gateway | Kong / Nginx | API 라우팅, Rate Limiting, 인증 프록시 |
| Application | Go (Gin/Echo) | Backend API 서버, 비즈니스 로직 |
| Domain | CAPI/CAPO Controller | K8s 클러스터 라이프사이클 관리 |
| Infrastructure | OpenStack APIs | Compute, Network, Image 리소스 프로비저닝 |
| Data | PostgreSQL + Redis | 영속적 저장 + 캐시/세션 |
| Messaging | NATS JetStream | 비동기 작업 큐, 이벤트 스트림 |

### 1.4 시스템 아키텍처 다이어그램

```
┌─────────────────────────────────────────────────────────────────┐
│                        사용자 (브라우저)                          │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTPS
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    API Gateway (Kong/Nginx)                      │
│              Rate Limiting · CORS · Auth Proxy                  │
└──────────┬──────────────────────────────────┬───────────────────┘
           │                                  │
           ▼                                  ▼
┌─────────────────────┐           ┌─────────────────────────┐
│   Next.js Frontend  │           │    Go API Server (Gin)   │
│   (SSR/Static)      │           │                         │
│                     │           │  ┌───────────────────┐  │
│  - Dashboard        │◄─────────┤  │   Handler Layer   │  │
│  - Cluster Wizard   │  REST    │  └────────┬──────────┘  │
│  - Monitoring       │  + WS    │  ┌────────▼──────────┐  │
│  - Admin Pages      │           │  │  Service Layer    │  │
└─────────────────────┘           │  └───┬────┬────┬────┘  │
                                  │      │    │    │        │
                                  └──────┼────┼────┼────────┘
                                         │    │    │
                    ┌────────────────────┘    │    └──────────────────┐
                    │                         │                       │
                    ▼                         ▼                       ▼
        ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
        │   PostgreSQL 16  │    │  Management K8s   │    │   OpenStack API  │
        │   + Redis 7      │    │  Cluster          │    │                  │
        │   + NATS         │    │                   │    │  - Nova          │
        │                  │    │  - CAPI Controller │    │  - Neutron       │
        │  - Users         │    │  - CAPO Controller │    │  - Glance        │
        │  - Tenants       │    │  - Cluster Objects │    │  - Keystone      │
        │  - Clusters      │    │  - Machine Objects │    │  - Cinder        │
        │  - Audit Logs    │    │                   │    │                  │
        └──────────────────┘    └────────┬──────────┘    └──────────────────┘
                                         │
                                         │ Provisions
                                         ▼
                                ┌──────────────────┐
                                │  Workload K8s    │
                                │  Clusters        │
                                │                  │
                                │  - Cluster A     │
                                │  - Cluster B     │
                                │  - Cluster C     │
                                └──────────────────┘
```

---

## 2. 기술 스택 상세

### 2.1 Backend — Go (Golang)

**선택 이유:**
- Kubernetes 생태계의 표준 언어로, client-go, controller-runtime 등 공식 라이브러리가 Go로 작성됨
- CAPI/CAPO가 Go로 개발되어 있어 동일 언어로 커스텀 컨트롤러 확장 용이
- 높은 동시성 처리 성능 (goroutine)으로 다수 클러스터 동시 작업에 적합
- 컴파일 바이너리 배포로 컨테이너 이미지 최소화

**주요 프레임워크/라이브러리:**

| 용도 | 라이브러리 | 설명 |
|------|-----------|------|
| Web Framework | `github.com/gin-gonic/gin` | RESTful API 서버, 미들웨어 지원 |
| K8s Client | `k8s.io/client-go` | Kubernetes API 접근 공식 클라이언트 |
| CAPI SDK | `sigs.k8s.io/cluster-api` | Cluster API 타입 및 유틸리티 |
| CAPO SDK | `sigs.k8s.io/cluster-api-provider-openstack` | CAPO 타입 정의 |
| OpenStack SDK | `github.com/gophercloud/gophercloud/v2` | OpenStack API 클라이언트 |
| ORM | `gorm.io/gorm` + `gorm.io/driver/postgres` | PostgreSQL ORM |
| Auth | `github.com/golang-jwt/jwt/v5` | JWT 토큰 발급/검증 |
| WebSocket | `github.com/gorilla/websocket` | 실시간 상태 스트림 |
| Config | `github.com/spf13/viper` | 설정 파일 관리 (config.yaml) |
| Logger | `go.uber.org/zap` | 구조화된 로깅 |
| Validation | `github.com/go-playground/validator/v10` | 구조체 유효성 검증 |
| Redis | `github.com/redis/go-redis/v9` | Redis 클라이언트 |
| NATS | `github.com/nats-io/nats.go` | NATS JetStream 클라이언트 |
| UUID | `github.com/google/uuid` | UUID 생성 |
| Crypto | `golang.org/x/crypto` | bcrypt 패스워드 해싱 |

### 2.2 Frontend — React (Next.js) + TypeScript

**선택 이유:**
- Server-Side Rendering(SSR)으로 초기 로딩 성능 확보
- TypeScript로 타입 안정성 및 개발 생산성 향상
- Kubernetes 대시보드 등 복잡한 UI에 적합한 컴포넌트 기반 개발

**주요 라이브러리:**

| 용도 | 라이브러리 | 설명 |
|------|-----------|------|
| UI Framework | `next@14+` | App Router, SSR/SSG 지원 |
| UI Components | `shadcn/ui` + `tailwindcss` | 커스터마이즈 가능한 컴포넌트 |
| 상태관리 | `zustand` | 경량 클라이언트 상태 관리 |
| 서버 상태 | `@tanstack/react-query` | API 캐시, 자동 리패치 |
| 폼 관리 | `react-hook-form` + `zod` | 폼 유효성 검증 |
| 차트 | `recharts` | 모니터링 대시보드 시각화 |
| 실시간 통신 | `socket.io-client` | WebSocket 기반 상태 스트림 |
| 터미널 | `xterm.js` | 웹 기반 kubectl 터미널 |
| 아이콘 | `lucide-react` | 아이콘 라이브러리 |
| 날짜 | `date-fns` | 날짜 포매팅 |
| 테이블 | `@tanstack/react-table` | 고성능 데이터 테이블 |

### 2.3 데이터베이스

| DB | 용도 | 선택 이유 |
|----|------|-----------|
| PostgreSQL 16 | 주 데이터 저장소 (사용자, 클러스터, 설정, 감사로그) | JSONB 지원으로 K8s 매니페스트 스냅샷 저장에 적합 |
| Redis 7 | 캐시, 세션, 작업 락 | 클러스터 상태 캐싱, 분산 락으로 동시 작업 충돌 방지 |
| NATS JetStream | 메시지 큐 / 이벤트 스트림 | 경량이며 K8s 생태계 표준, 작업 비동기 처리에 최적 |

### 2.4 인프라/DevOps

| 용도 | 도구 | 설명 |
|------|------|------|
| 컨테이너 런타임 | Docker + containerd | 개발/프로덕션 컨테이너화 |
| CI/CD | GitHub Actions / GitLab CI | 빌드, 테스트, 배포 자동화 |
| IaC | Helm Charts | CMP 자체의 K8s 배포 관리 |
| 모니터링 | Prometheus + Grafana | CMP 및 클러스터 메트릭 모니터링 |
| 로깅 | Loki + Promtail | 중앙화된 로그 수집/검색 |
| 시크릿 관리 | HashiCorp Vault | OpenStack 자격증명, kubeconfig 암호화 저장 |

---

## 3. 프로젝트 구조

### 3.1 모노레포 디렉토리 구조

프로젝트는 모노레포로 구성하여 코드 공유와 버전 관리를 단순화한다.

```
k8s-cmp/
├── cmd/                              # 엔트리포인트
│   ├── api-server/
│   │   └── main.go                   # API 서버 부팅 (Gin, DB, Redis, NATS 초기화)
│   ├── worker/
│   │   └── main.go                   # 비동기 작업 워커 (NATS Consumer)
│   └── migrator/
│       └── main.go                   # DB 마이그레이션 실행기
│
├── internal/                         # 비공개 패키지 (외부 import 불가)
│   ├── domain/                       # 도메인 모델
│   │   ├── user.go                   # User 엔티티
│   │   ├── tenant.go                 # Tenant 엔티티
│   │   ├── cluster.go                # Cluster 엔티티
│   │   ├── node.go                   # Node 엔티티
│   │   ├── cluster_event.go          # ClusterEvent 엔티티
│   │   └── audit_log.go              # AuditLog 엔티티
│   │
│   ├── repository/                   # 데이터 접근 계층
│   │   ├── interfaces.go             # Repository 인터페이스 정의
│   │   ├── user_repo.go
│   │   ├── tenant_repo.go
│   │   ├── cluster_repo.go
│   │   └── audit_log_repo.go
│   │
│   ├── service/                      # 비즈니스 로직 계층
│   │   ├── auth_service.go           # 인증/인가 로직
│   │   ├── tenant_service.go         # 테넌트 관리
│   │   ├── cluster_service.go        # 클러스터 라이프사이클
│   │   ├── openstack_service.go      # OpenStack 리소스 조회
│   │   └── notification_service.go   # 알림 처리
│   │
│   ├── handler/                      # HTTP 핸들러 (Controller)
│   │   ├── auth_handler.go
│   │   ├── tenant_handler.go
│   │   ├── cluster_handler.go
│   │   ├── openstack_handler.go
│   │   └── websocket_handler.go
│   │
│   ├── middleware/                    # 미들웨어
│   │   ├── auth.go                   # JWT 인증 미들웨어
│   │   ├── rbac.go                   # 역할 기반 인가
│   │   ├── tenant.go                 # 테넌트 컨텍스트 주입
│   │   ├── logger.go                 # 요청 로깅
│   │   ├── cors.go                   # CORS 설정
│   │   └── audit.go                  # 감사 로그 자동 기록
│   │
│   ├── k8s/                          # CAPI/CAPO 연동
│   │   ├── client.go                 # Management Cluster 클라이언트 초기화
│   │   ├── manifest_generator.go     # CAPI 매니페스트 자동 생성
│   │   ├── manifest_templates/       # Go 템플릿 파일
│   │   │   ├── cluster.yaml.tmpl
│   │   │   ├── openstack_cluster.yaml.tmpl
│   │   │   ├── kubeadm_control_plane.yaml.tmpl
│   │   │   ├── machine_deployment.yaml.tmpl
│   │   │   ├── kubeadm_config_template.yaml.tmpl
│   │   │   └── openstack_machine_template.yaml.tmpl
│   │   ├── cluster_manager.go        # 매니페스트 Apply/Delete
│   │   ├── watcher.go                # Informer 기반 상태 감시
│   │   └── kubeconfig.go             # Kubeconfig 추출
│   │
│   ├── openstack/                    # OpenStack API 연동
│   │   ├── client.go                 # gophercloud 클라이언트 팩토리
│   │   ├── compute.go                # Nova (Flavors, Keypairs)
│   │   ├── network.go                # Neutron (Networks, Subnets, SG)
│   │   ├── image.go                  # Glance (Images)
│   │   └── quota.go                  # Quota 조회
│   │
│   └── websocket/                    # 실시간 통신
│       ├── hub.go                    # WebSocket 연결 관리 허브
│       ├── client.go                 # 개별 클라이언트 연결
│       └── events.go                 # 이벤트 타입 정의
│
├── pkg/                              # 공개 유틸리티
│   ├── errors/
│   │   └── errors.go                 # 커스텀 에러 타입 (AppError, ErrorCode)
│   ├── response/
│   │   └── response.go              # 표준 API 응답 래퍼
│   ├── crypto/
│   │   └── aes.go                    # AES-256-GCM 암호화/복호화
│   ├── pagination/
│   │   └── pagination.go            # 페이지네이션 유틸
│   └── validator/
│       └── custom.go                 # 커스텀 유효성 검증 규칙
│
├── web/                              # Next.js 프론트엔드
│   ├── package.json
│   ├── next.config.js
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── src/
│   │   ├── app/                      # App Router 페이지
│   │   │   ├── layout.tsx            # 루트 레이아웃
│   │   │   ├── page.tsx              # 리다이렉트 → /dashboard
│   │   │   ├── login/
│   │   │   │   └── page.tsx
│   │   │   ├── dashboard/
│   │   │   │   └── page.tsx
│   │   │   ├── clusters/
│   │   │   │   ├── page.tsx          # 클러스터 목록
│   │   │   │   ├── new/
│   │   │   │   │   └── page.tsx      # 클러스터 생성 위저드
│   │   │   │   └── [id]/
│   │   │   │       ├── page.tsx      # 클러스터 상세
│   │   │   │       ├── scale/
│   │   │   │       │   └── page.tsx
│   │   │   │       └── upgrade/
│   │   │   │           └── page.tsx
│   │   │   ├── admin/
│   │   │   │   ├── tenants/
│   │   │   │   │   └── page.tsx
│   │   │   │   ├── users/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── audit-logs/
│   │   │   │       └── page.tsx
│   │   │   └── settings/
│   │   │       └── page.tsx
│   │   │
│   │   ├── components/
│   │   │   ├── ui/                   # shadcn/ui 기본 컴포넌트
│   │   │   ├── layout/
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   ├── Header.tsx
│   │   │   │   └── MainLayout.tsx
│   │   │   ├── clusters/
│   │   │   │   ├── ClusterCard.tsx
│   │   │   │   ├── ClusterTable.tsx
│   │   │   │   ├── ClusterStatusBadge.tsx
│   │   │   │   ├── CreateClusterWizard.tsx
│   │   │   │   ├── wizard-steps/
│   │   │   │   │   ├── BasicInfoStep.tsx
│   │   │   │   │   ├── NetworkStep.tsx
│   │   │   │   │   ├── NodeConfigStep.tsx
│   │   │   │   │   └── ReviewStep.tsx
│   │   │   │   ├── ClusterDetail.tsx
│   │   │   │   ├── NodeList.tsx
│   │   │   │   └── ClusterEvents.tsx
│   │   │   ├── dashboard/
│   │   │   │   ├── ClusterOverview.tsx
│   │   │   │   ├── ResourceUsageChart.tsx
│   │   │   │   └── RecentEvents.tsx
│   │   │   └── common/
│   │   │       ├── DataTable.tsx
│   │   │       ├── ConfirmDialog.tsx
│   │   │       ├── LoadingSpinner.tsx
│   │   │       └── EmptyState.tsx
│   │   │
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   ├── useClusters.ts
│   │   │   ├── useOpenStackResources.ts
│   │   │   ├── useWebSocket.ts
│   │   │   └── useAuditLogs.ts
│   │   │
│   │   ├── lib/
│   │   │   ├── api-client.ts         # Axios 인스턴스 + 인터셉터
│   │   │   ├── auth.ts               # 토큰 관리
│   │   │   ├── websocket.ts          # WebSocket 연결 관리
│   │   │   └── utils.ts
│   │   │
│   │   ├── types/
│   │   │   ├── api.ts                # API 응답 공통 타입
│   │   │   ├── cluster.ts
│   │   │   ├── tenant.ts
│   │   │   ├── user.ts
│   │   │   ├── openstack.ts
│   │   │   └── audit.ts
│   │   │
│   │   └── stores/
│   │       ├── auth-store.ts         # Zustand 인증 상태
│   │       └── ui-store.ts           # UI 상태 (사이드바, 테마)
│   │
│   └── public/                       # 정적 파일
│
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile.api            # Go API 서버
│   │   ├── Dockerfile.worker         # Worker
│   │   └── Dockerfile.web            # Next.js
│   ├── helm/
│   │   └── k8s-cmp/                  # Helm Chart
│   │       ├── Chart.yaml
│   │       ├── values.yaml
│   │       └── templates/
│   └── docker-compose.yaml           # 로컬 개발 환경
│
├── migrations/                       # SQL 마이그레이션 (순차 번호)
│   ├── 001_create_tenants.up.sql
│   ├── 001_create_tenants.down.sql
│   ├── 002_create_users.up.sql
│   ├── 002_create_users.down.sql
│   ├── 003_create_clusters.up.sql
│   ├── 003_create_clusters.down.sql
│   ├── 004_create_nodes.up.sql
│   ├── 004_create_nodes.down.sql
│   ├── 005_create_cluster_events.up.sql
│   ├── 005_create_cluster_events.down.sql
│   ├── 006_create_audit_logs.up.sql
│   └── 006_create_audit_logs.down.sql
│
├── configs/
│   ├── config.yaml                   # 기본 설정
│   └── config.example.yaml           # 설정 샘플
│
├── docs/
│   ├── api.md                        # API 문서
│   └── architecture.md               # 아키텍처 문서
│
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.md
├── SPEC.md                           # (이 파일)
└── README.md
```

### 3.2 config.yaml 구조

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"  # debug | release

database:
  host: "localhost"
  port: 5432
  name: "k8s_cmp"
  user: "cmp"
  password: "${DB_PASSWORD}"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 10

redis:
  host: "localhost"
  port: 6379
  password: "${REDIS_PASSWORD}"
  db: 0

nats:
  url: "nats://localhost:4222"
  stream_name: "CMP_TASKS"

auth:
  jwt_secret: "${JWT_SECRET}"
  access_token_ttl: "15m"
  refresh_token_ttl: "7d"

management_cluster:
  kubeconfig_path: "/etc/cmp/kubeconfig"
  # 또는 in-cluster 모드 시 비워둠

encryption:
  key: "${ENCRYPTION_KEY}"  # AES-256-GCM 키 (32 bytes, base64 인코딩)

openstack:
  # 테넌트별 자격증명은 DB에서 관리
  # 여기에는 기본/공통 설정만

cors:
  allowed_origins:
    - "http://localhost:3000"
    - "https://cmp.example.com"
```

---

## 4. 기능 상세 명세

### 4.1 인증/인가 시스템

#### 4.1.1 요구사항

- JWT 기반 인증 (Access Token + Refresh Token)
- RBAC (Role-Based Access Control) 역할 체계:
  - **Super Admin:** 전체 시스템 관리, 테넌트 생성/삭제
  - **Tenant Admin:** 소속 테넌트 내 클러스터 및 사용자 관리
  - **Operator:** 클러스터 생성/수정/삭제, 스케일링
  - **Viewer:** 읽기 전용, 대시보드 조회
- (Optional) LDAP/OIDC 외부 IdP 연동 지원
- 세션 관리: Redis에 Refresh Token 저장, 블랙리스트 지원

#### 4.1.2 API 엔드포인트

| 메서드 | 경로 | 설명 | 인증 필요 |
|--------|------|------|-----------|
| `POST` | `/api/v1/auth/login` | 로그인 (이메일/비밀번호) | No |
| `POST` | `/api/v1/auth/refresh` | 토큰 갱신 | No (Refresh Token) |
| `POST` | `/api/v1/auth/logout` | 로그아웃 (토큰 무효화) | Yes |
| `GET` | `/api/v1/auth/me` | 현재 사용자 정보 조회 | Yes |

#### 4.1.3 JWT 토큰 구조

**Access Token Payload:**
```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "role": "operator",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**미들웨어 체인:**
```
Request → CORS → Logger → JWT Auth → RBAC → Tenant Context → Handler
```

#### 4.1.4 RBAC 권한 매트릭스

| 리소스/액션 | Super Admin | Tenant Admin | Operator | Viewer |
|-------------|:-----------:|:------------:|:--------:|:------:|
| 테넌트 CRUD | ✅ | ❌ | ❌ | ❌ |
| 사용자 관리 | ✅ | ✅ (소속 테넌트) | ❌ | ❌ |
| 클러스터 생성 | ✅ | ✅ | ✅ | ❌ |
| 클러스터 삭제 | ✅ | ✅ | ✅ | ❌ |
| 클러스터 조회 | ✅ | ✅ | ✅ | ✅ |
| 스케일링 | ✅ | ✅ | ✅ | ❌ |
| 업그레이드 | ✅ | ✅ | ✅ | ❌ |
| Kubeconfig 다운로드 | ✅ | ✅ | ✅ | ❌ |
| 감사 로그 조회 | ✅ | ✅ (소속 테넌트) | ❌ | ❌ |
| 시스템 설정 | ✅ | ❌ | ❌ | ❌ |

---

### 4.2 테넌트 관리

#### 4.2.1 요구사항

- 테넌트 CRUD (Super Admin 전용)
- 테넌트별 OpenStack 자격증명 관리 (Keystone Project 매핑)
- 테넌트별 리소스 쿼터 설정 (최대 클러스터 수, 최대 노드 수)
- 테넌트 멤버 초대 및 권한 부여

#### 4.2.2 데이터 모델 — Tenant

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | `UUID` PK | 테넌트 고유 ID |
| `name` | `VARCHAR(100)` UNIQUE | 테넌트 이름 |
| `description` | `TEXT` | 설명 |
| `os_auth_url` | `VARCHAR(255)` | OpenStack Keystone URL |
| `os_project_id` | `VARCHAR(64)` | OpenStack Project ID |
| `os_project_name` | `VARCHAR(100)` | OpenStack Project 이름 |
| `os_domain_name` | `VARCHAR(100)` | OpenStack Domain 이름 |
| `os_credentials` | `TEXT` (encrypted) | 암호화된 OpenStack 자격증명 (JSON) |
| `max_clusters` | `INT` DEFAULT 10 | 최대 클러스터 수 제한 |
| `max_nodes_per_cluster` | `INT` DEFAULT 20 | 클러스터당 최대 노드 수 |
| `status` | `VARCHAR(20)` | `active` / `suspended` / `deleted` |
| `created_at` | `TIMESTAMP` | 생성일시 |
| `updated_at` | `TIMESTAMP` | 수정일시 |
| `deleted_at` | `TIMESTAMP` NULL | 삭제일시 (soft delete) |

**os_credentials 암호화 JSON 구조:**
```json
{
  "username": "admin",
  "password": "encrypted...",
  "user_domain_name": "Default"
}
```

#### 4.2.3 API 엔드포인트

| 메서드 | 경로 | 설명 | 권한 |
|--------|------|------|------|
| `GET` | `/api/v1/tenants` | 테넌트 목록 | Super Admin |
| `GET` | `/api/v1/tenants/:id` | 테넌트 상세 | Super Admin, Tenant Admin(소속) |
| `POST` | `/api/v1/tenants` | 테넌트 생성 | Super Admin |
| `PUT` | `/api/v1/tenants/:id` | 테넌트 수정 | Super Admin |
| `DELETE` | `/api/v1/tenants/:id` | 테넌트 삭제 | Super Admin |
| `POST` | `/api/v1/tenants/:id/verify-connection` | OpenStack 연결 테스트 | Super Admin |
| `GET` | `/api/v1/tenants/:id/members` | 멤버 목록 | Tenant Admin+ |
| `POST` | `/api/v1/tenants/:id/members` | 멤버 추가 | Tenant Admin+ |
| `DELETE` | `/api/v1/tenants/:id/members/:userId` | 멤버 제거 | Tenant Admin+ |

---

### 4.3 클러스터 라이프사이클 관리 (핵심 기능)

#### 4.3.1 클러스터 생성 — 입력 파라미터

사용자가 웹 폼에서 입력한 값으로 CAPI/CAPO 매니페스트를 자동 생성하여 Management Cluster에 적용한다.

| 파라미터 | 타입 | 필수 | 설명 | 예시 | 유효성 검증 |
|----------|------|:----:|------|------|-------------|
| `cluster_name` | `string` | ✅ | 클러스터 이름 (tenant 내 unique) | `my-prod-cluster` | `^[a-z][a-z0-9-]{2,62}$` |
| `kubernetes_version` | `string` | ✅ | K8s 버전 | `v1.29.2` | 지원 버전 목록 검증 |
| `control_plane_count` | `int` | ✅ | CP 노드 수 (홀수) | `3` | `1`, `3`, `5` 중 택1 |
| `worker_count` | `int` | ✅ | Worker 노드 수 | `5` | `1` ~ `max_nodes_per_cluster` |
| `control_plane_flavor` | `string` | ✅ | CP용 OpenStack Flavor | `m1.xlarge` | OpenStack Flavor 존재 검증 |
| `worker_flavor` | `string` | ✅ | Worker용 OpenStack Flavor | `m1.large` | OpenStack Flavor 존재 검증 |
| `os_image` | `string` | ✅ | OS 이미지 이름 | `ubuntu-2204-kube-v1.29.2` | OpenStack Image 존재 검증 |
| `network_id` | `string` | ✅ | OpenStack Network ID | `UUID` | OpenStack Network 존재 검증 |
| `subnet_id` | `string` | ✅ | OpenStack Subnet ID | `UUID` | 선택된 Network에 속하는지 검증 |
| `external_network_id` | `string` | ✅ | 외부 네트워크 (Floating IP) | `UUID` | External Network 존재 검증 |
| `ssh_key_name` | `string` | ✅ | SSH 키페어 이름 | `my-keypair` | OpenStack Keypair 존재 검증 |
| `pod_cidr` | `string` | ❌ | Pod 네트워크 CIDR | `10.244.0.0/16` | CIDR 형식 검증, 기본값 제공 |
| `service_cidr` | `string` | ❌ | Service 네트워크 CIDR | `10.96.0.0/12` | CIDR 형식 검증, 기본값 제공 |
| `cni_plugin` | `string` | ❌ | CNI 플러그인 | `calico` | `calico` / `cilium` 중 택1 |
| `dns_nameservers` | `[]string` | ❌ | DNS 서버 목록 | `["8.8.8.8"]` | IP 형식 검증, 기본값 제공 |

#### 4.3.2 CAPI/CAPO 매니페스트 생성 로직

Backend는 입력값을 바탕으로 다음 CAPI/CAPO 리소스를 자동 생성한다.
Go의 `text/template`을 사용하여 `internal/k8s/manifest_templates/` 디렉토리의 템플릿에서 생성한다.

**생성되는 리소스 목록 (생성 순서):**

| 순서 | CAPI/CAPO 리소스 | Kind | API Group | 설명 |
|:----:|-----------------|------|-----------|------|
| 1 | Cluster | `Cluster` | `cluster.x-k8s.io/v1beta1` | CAPI 최상위 클러스터 정의 |
| 2 | OpenStackCluster | `OpenStackCluster` | `infrastructure.cluster.x-k8s.io/v1beta1` | OpenStack 인프라 설정 |
| 3 | KubeadmControlPlane | `KubeadmControlPlane` | `controlplane.cluster.x-k8s.io/v1beta1` | CP 구성 및 kubeadm 설정 |
| 4 | OpenStackMachineTemplate (CP) | `OpenStackMachineTemplate` | `infrastructure.cluster.x-k8s.io/v1beta1` | CP 노드용 VM 스펙 |
| 5 | MachineDeployment | `MachineDeployment` | `cluster.x-k8s.io/v1beta1` | Worker 노드 그룹 정의 |
| 6 | KubeadmConfigTemplate | `KubeadmConfigTemplate` | `bootstrap.cluster.x-k8s.io/v1beta1` | Worker 노드 kubeadm 설정 |
| 7 | OpenStackMachineTemplate (Worker) | `OpenStackMachineTemplate` | `infrastructure.cluster.x-k8s.io/v1beta1` | Worker 노드용 VM 스펙 |

**추가 생성 리소스:**
- `Secret` (OpenStack 클라우드 자격증명) — `clouds.yaml` 형식으로 생성

**매니페스트 생성 함수 시그니처:**
```go
// internal/k8s/manifest_generator.go

type ClusterCreateParams struct {
    ClusterName        string
    Namespace          string   // 테넌트별 네임스페이스
    KubernetesVersion  string
    ControlPlaneCount  int
    WorkerCount        int
    CPFlavor           string
    WorkerFlavor       string
    OSImage            string
    NetworkID          string
    SubnetID           string
    ExternalNetworkID  string
    SSHKeyName         string
    PodCIDR            string
    ServiceCIDR        string
    CNIPlugin          string
    DNSNameservers     []string
    CloudName          string   // clouds.yaml 내 클라우드 이름
}

// GenerateManifests는 CAPI/CAPO 리소스 매니페스트를 생성한다.
// 반환값은 Kubernetes unstructured 오브젝트 슬라이스.
func (g *ManifestGenerator) GenerateManifests(params ClusterCreateParams) ([]unstructured.Unstructured, error)

// ApplyManifests는 생성된 매니페스트를 Management Cluster에 적용한다.
func (m *ClusterManager) ApplyManifests(ctx context.Context, manifests []unstructured.Unstructured) error

// DeleteCluster는 클러스터 관련 모든 CAPI 리소스를 삭제한다.
func (m *ClusterManager) DeleteCluster(ctx context.Context, namespace, clusterName string) error
```

**매니페스트 Apply 프로세스:**
1. 입력 파라미터 유효성 검증
2. OpenStack 리소스 존재 여부 확인 (Flavor, Image, Network 등)
3. 테넌트 쿼터 확인 (클러스터 수, 노드 수)
4. Go 템플릿으로 매니페스트 생성
5. OpenStack 클라우드 자격증명 Secret 생성
6. `client-go`의 dynamic client로 Management Cluster에 Apply
7. DB에 클러스터 레코드 생성 (status: `Pending`)
8. NATS에 프로비저닝 모니터링 작업 발행
9. 매니페스트 스냅샷을 DB의 `capi_manifest_snapshot` 필드에 JSONB로 저장

#### 4.3.3 클러스터 상태 머신

클러스터는 다음과 같은 상태 전이를 가진다:

```
                    ┌──────────┐
                    │ Pending  │
                    └────┬─────┘
                         │ 매니페스트 Apply 완료
                         ▼
                ┌────────────────┐
                │ Provisioning   │──────────────┐
                └───────┬────────┘              │
                        │ 모든 노드 Ready         │ 타임아웃/에러
                        ▼                        ▼
                 ┌──────────┐              ┌──────────┐
        ┌───────│ Running   │              │  Failed  │
        │       └─────┬─────┘              └────┬─────┘
        │             │                         │
        │     Scale/  │                    Retry │ Delete
        │    Upgrade  │                         │
        │             ▼                         │
        │     ┌───────────┐                     │
        │     │ Updating  │                     │
        │     └─────┬─────┘                     │
        │           │ 완료                       │
        │           ▼                           │
        │     ┌──────────┐                      │
        └────►│ Running  │                      │
              └─────┬────┘                      │
                    │ Delete                     │
                    ▼                           │
             ┌───────────┐                      │
             │ Deleting  │◄─────────────────────┘
             └─────┬─────┘
                   │ 완료
                   ▼
             ┌───────────┐
             │  Deleted   │
             └───────────┘
```

**상태 정의:**

| 상태 | 설명 | 허용 액션 |
|------|------|-----------|
| `Pending` | 생성 요청됨, 매니페스트 준비 중 | 취소 |
| `Provisioning` | CAPI가 VM 및 K8s 컴포넌트 프로비저닝 중 | 취소 |
| `Running` | 정상 운영 중 | 스케일링, 업그레이드, 삭제, kubeconfig 다운로드 |
| `Updating` | 스케일링 또는 업그레이드 진행 중 | 대기 (모니터링만 가능) |
| `Failed` | 프로비저닝 실패 | 재시도, 삭제 |
| `Deleting` | 삭제 진행 중 | 대기 |
| `Deleted` | 완전 삭제됨 (soft delete) | - |

#### 4.3.4 클러스터 API 엔드포인트

| 메서드 | 경로 | 설명 | 권한 |
|--------|------|------|------|
| `GET` | `/api/v1/clusters` | 클러스터 목록 조회 (페이지네이션, 필터링) | Viewer+ |
| `GET` | `/api/v1/clusters/:id` | 클러스터 상세 조회 | Viewer+ |
| `POST` | `/api/v1/clusters` | 클러스터 생성 | Operator+ |
| `DELETE` | `/api/v1/clusters/:id` | 클러스터 삭제 | Operator+ |
| `GET` | `/api/v1/clusters/:id/kubeconfig` | Kubeconfig 다운로드 | Operator+ |
| `GET` | `/api/v1/clusters/:id/events` | 클러스터 이벤트 조회 | Viewer+ |
| `GET` | `/api/v1/clusters/:id/nodes` | 노드 목록 및 상태 | Viewer+ |
| `POST` | `/api/v1/clusters/:id/scale` | 노드 수 조정 | Operator+ |
| `POST` | `/api/v1/clusters/:id/upgrade` | K8s 버전 업그레이드 | Operator+ |
| `POST` | `/api/v1/clusters/:id/retry` | 실패한 클러스터 재시도 | Operator+ |

**클러스터 목록 쿼리 파라미터:**

| 파라미터 | 타입 | 설명 |
|----------|------|------|
| `page` | `int` | 페이지 번호 (기본 1) |
| `page_size` | `int` | 페이지 크기 (기본 20, 최대 100) |
| `status` | `string` | 상태 필터 (comma separated) |
| `search` | `string` | 이름 검색 |
| `sort_by` | `string` | 정렬 필드 (`name`, `created_at`, `status`) |
| `sort_order` | `string` | `asc` / `desc` |

**스케일링 요청 본문:**
```json
{
  "control_plane_count": 3,
  "worker_count": 10
}
```

**업그레이드 요청 본문:**
```json
{
  "target_version": "v1.30.0"
}
```

#### 4.3.5 스케일링 구현 상세

**Worker Node 스케일링:**
- `MachineDeployment`의 `.spec.replicas` 값을 변경
- CAPI가 자동으로 Machine을 추가/제거
- Scale Down 시 가장 최근 생성된 노드부터 제거 (기본 동작)

**Control Plane 스케일링:**
- `KubeadmControlPlane`의 `.spec.replicas` 값을 변경
- 홀수만 허용 (1, 3, 5)
- etcd 멤버 자동 관리

**구현 함수:**
```go
// internal/service/cluster_service.go

type ScaleParams struct {
    ControlPlaneCount *int  // nil이면 변경 안함
    WorkerCount       *int  // nil이면 변경 안함
}

func (s *ClusterService) ScaleCluster(ctx context.Context, clusterID uuid.UUID, params ScaleParams) error
```

#### 4.3.6 업그레이드 구현 상세

**업그레이드 프로세스:**
1. 대상 버전 유효성 검증 (현재 버전보다 높은지, 지원 버전인지)
2. 해당 버전의 OS 이미지 존재 여부 확인
3. `KubeadmControlPlane`의 `.spec.version` 변경 → CP Rolling Update
4. CP 업그레이드 완료 후 `MachineDeployment`의 `.spec.template.spec.version` 변경 → Worker Rolling Update
5. `OpenStackMachineTemplate`의 이미지도 새 버전에 맞게 교체 (새 Template 생성 → 참조 변경)

**주의사항:**
- CAPI는 한 단계씩 마이너 버전 업그레이드만 지원 (1.28 → 1.29 → 1.30)
- 업그레이드 중 클러스터는 `Updating` 상태

---

### 4.4 OpenStack 리소스 연동

#### 4.4.1 요구사항

CMP는 클러스터 생성 시 사용자가 선택할 수 있는 OpenStack 리소스 목록을 동적으로 조회한다.
각 테넌트의 OpenStack 자격증명을 사용하여 해당 프로젝트의 리소스만 조회한다.

#### 4.4.2 연동 리소스 목록

| OpenStack 리소스 | API | 용도 | 캐시 TTL |
|-----------------|-----|------|----------|
| Flavors | Nova API | VM 사양 목록 (CPU/Memory/Disk) | 5분 |
| Images | Glance API | OS 이미지 목록 (K8s 호환 이미지 필터링) | 5분 |
| Networks | Neutron API | 네트워크 목록 | 5분 |
| Subnets | Neutron API | 서브넷 목록 (Network별 필터링) | 5분 |
| Security Groups | Neutron API | 보안 그룹 목록 | 5분 |
| Key Pairs | Nova API | SSH 키페어 목록 | 5분 |
| Floating IPs | Neutron API | 외부 IP 할당 현황 | 1분 |
| Quotas | Nova/Neutron/Cinder API | 테넌트별 리소스 사용량 | 1분 |

**이미지 필터링 규칙:**
- 이미지 이름에 `kube-v` 패턴이 포함된 이미지만 표시
- 또는 메타데이터 태그 `k8s_version` 이 있는 이미지

#### 4.4.3 API 엔드포인트

| 메서드 | 경로 | 설명 |
|--------|------|------|
| `GET` | `/api/v1/openstack/flavors` | Flavor 목록 |
| `GET` | `/api/v1/openstack/images` | K8s 호환 이미지 목록 |
| `GET` | `/api/v1/openstack/networks` | 네트워크 목록 |
| `GET` | `/api/v1/openstack/networks/:id/subnets` | 특정 네트워크의 서브넷 목록 |
| `GET` | `/api/v1/openstack/security-groups` | 보안 그룹 목록 |
| `GET` | `/api/v1/openstack/keypairs` | SSH 키페어 목록 |
| `GET` | `/api/v1/openstack/external-networks` | 외부 네트워크 목록 |
| `GET` | `/api/v1/openstack/quotas` | 리소스 쿼터 사용량 |

#### 4.4.4 OpenStack 클라이언트 구현

```go
// internal/openstack/client.go

// NewClient는 테넌트의 암호화된 자격증명을 복호화하여 gophercloud 클라이언트를 생성한다.
func NewClient(ctx context.Context, tenant *domain.Tenant, decryptor *crypto.AES) (*gophercloud.ProviderClient, error)

// 각 서비스별 클라이언트 래퍼
type ComputeService struct { client *gophercloud.ServiceClient }
type NetworkService struct { client *gophercloud.ServiceClient }
type ImageService struct { client *gophercloud.ServiceClient }
```

**캐싱 전략:**
- Redis에 `openstack:{tenant_id}:{resource_type}` 키로 캐시
- TTL은 리소스 타입별로 차등 적용 (위 테이블 참조)
- 클러스터 생성/삭제 시 관련 캐시 무효화

---

### 4.5 모니터링 및 알림

#### 4.5.1 모니터링 대상

**Kubernetes Informer를 통한 실시간 감시:**
- Cluster 오브젝트: `status.phase` 변경 감지
- Machine 오브젝트: Ready Condition 감지
- KubeadmControlPlane: replicas/readyReplicas 비교
- MachineDeployment: replicas/readyReplicas 비교

**OpenStack VM 상태 동기화:**
- 주기적으로 (30초) Nova API로 VM 상태 확인
- VM이 ERROR 상태이면 해당 노드를 Failed로 표시

#### 4.5.2 Watcher 구현

```go
// internal/k8s/watcher.go

type ClusterWatcher struct {
    dynamicClient dynamic.Interface
    informers     map[string]cache.SharedIndexInformer
    hub           *websocket.Hub
    eventChan     chan ClusterEvent
}

// Start는 CAPI 리소스에 대한 Informer를 시작한다.
func (w *ClusterWatcher) Start(ctx context.Context) error

// 감시 대상 리소스:
// - clusters.cluster.x-k8s.io
// - machines.cluster.x-k8s.io
// - machinedeployments.cluster.x-k8s.io
// - kubeadmcontrolplanes.controlplane.cluster.x-k8s.io
// - openstackclusters.infrastructure.cluster.x-k8s.io
```

#### 4.5.3 알림 체계

| 알림 유형 | 트리거 | 채널 |
|-----------|--------|------|
| 클러스터 생성 완료 | Cluster phase = Running | WebSocket, 이메일 |
| 클러스터 생성 실패 | Cluster phase = Failed | WebSocket, 이메일, Webhook |
| 노드 비정상 | Machine Ready = False > 5분 | WebSocket, 이메일 |
| 스케일링 완료 | MachineDeployment replicas 달성 | WebSocket |
| 업그레이드 완료 | KCP version 변경 확인 | WebSocket, 이메일 |
| 쿼터 경고 | 테넌트 리소스 사용량 > 80% | 이메일, 대시보드 |

#### 4.5.4 실시간 상태 스트림 (WebSocket)

**WebSocket 엔드포인트:** `ws(s)://{host}/ws/v1/clusters/:id/stream`

**인증:** 연결 시 쿼리 파라미터로 JWT 토큰 전달
```
ws://api.example.com/ws/v1/clusters/{id}/stream?token={jwt_token}
```

**전송 메시지 형식:**
```json
{
  "type": "cluster.status.changed",
  "cluster_id": "uuid",
  "data": {
    "phase": "Running",
    "ready": true,
    "control_plane_ready": true,
    "nodes_ready": 5,
    "nodes_total": 5,
    "conditions": [
      {
        "type": "InfrastructureReady",
        "status": "True",
        "last_transition_time": "2026-03-21T10:00:00Z"
      }
    ]
  },
  "timestamp": "2026-03-21T10:00:00Z"
}
```

**이벤트 타입:**
- `cluster.status.changed` — 클러스터 상태 변경
- `cluster.node.added` — 새 노드 추가
- `cluster.node.removed` — 노드 제거
- `cluster.node.ready` — 노드 Ready 상태 변경
- `cluster.provision.progress` — 프로비저닝 진행률
- `cluster.error` — 에러 발생

**WebSocket Hub 구현:**
```go
// internal/websocket/hub.go

type Hub struct {
    // 클러스터ID → 클라이언트 목록
    clusters map[string]map[*Client]bool
    
    register   chan *Subscription
    unregister chan *Subscription
    broadcast  chan *Event
}

type Subscription struct {
    ClusterID string
    Client    *Client
}
```

**전체 클러스터 상태 스트림 (대시보드용):**
`ws(s)://{host}/ws/v1/clusters/stream` — 소속 테넌트의 모든 클러스터 상태 변경을 수신

---

### 4.6 감사 로그 (Audit Log)

모든 주요 작업을 기록하여 보안 감사 및 문제 추적을 지원한다.

#### 4.6.1 데이터 모델 — AuditLog

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | `UUID` PK | 로그 ID |
| `tenant_id` | `UUID` FK | 테넌트 ID |
| `user_id` | `UUID` FK | 작업 수행 사용자 |
| `action` | `VARCHAR(50)` | 액션 (`cluster.create`, `cluster.delete`, `cluster.scale` 등) |
| `resource_type` | `VARCHAR(50)` | 대상 리소스 타입 (`cluster`, `tenant`, `user`) |
| `resource_id` | `VARCHAR(100)` | 대상 리소스 ID |
| `resource_name` | `VARCHAR(200)` | 대상 리소스 이름 (검색용) |
| `request_body` | `JSONB` | 요청 내용 (민감정보 마스킹) |
| `response_status` | `INT` | HTTP 상태 코드 |
| `result` | `VARCHAR(20)` | `success` / `failure` |
| `error_message` | `TEXT` | 실패 시 오류 메시지 |
| `ip_address` | `VARCHAR(45)` | 요청자 IP |
| `user_agent` | `VARCHAR(500)` | User-Agent |
| `duration_ms` | `INT` | 요청 처리 시간 (ms) |
| `created_at` | `TIMESTAMP` | 발생 시간 |

**기록 대상 액션:**
- `auth.login`, `auth.logout`
- `tenant.create`, `tenant.update`, `tenant.delete`
- `user.create`, `user.update`, `user.delete`, `user.role.change`
- `cluster.create`, `cluster.delete`, `cluster.scale`, `cluster.upgrade`, `cluster.retry`
- `cluster.kubeconfig.download`

**민감정보 마스킹:**
- `os_credentials` → `"***"`
- `password` → `"***"`
- JWT 토큰 → `"***"`

#### 4.6.2 API 엔드포인트

| 메서드 | 경로 | 설명 | 권한 |
|--------|------|------|------|
| `GET` | `/api/v1/audit-logs` | 감사 로그 검색/필터링 | Super Admin, Tenant Admin |
| `GET` | `/api/v1/audit-logs/:id` | 감사 로그 상세 | Super Admin, Tenant Admin |

**쿼리 파라미터:**

| 파라미터 | 타입 | 설명 |
|----------|------|------|
| `page`, `page_size` | `int` | 페이지네이션 |
| `action` | `string` | 액션 필터 |
| `resource_type` | `string` | 리소스 타입 필터 |
| `user_id` | `string` | 사용자 필터 |
| `result` | `string` | 결과 필터 (`success`/`failure`) |
| `from`, `to` | `datetime` | 기간 필터 |
| `search` | `string` | 리소스 이름 검색 |

---

### 4.7 사용자 관리

#### 4.7.1 데이터 모델 — User

| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | `UUID` PK | 사용자 고유 ID |
| `tenant_id` | `UUID` FK (nullable) | 소속 테넌트 (Super Admin은 null) |
| `email` | `VARCHAR(255)` UNIQUE | 이메일 |
| `password_hash` | `VARCHAR(255)` | bcrypt 해시된 비밀번호 |
| `name` | `VARCHAR(100)` | 이름 |
| `role` | `VARCHAR(20)` | `super_admin` / `tenant_admin` / `operator` / `viewer` |
| `status` | `VARCHAR(20)` | `active` / `inactive` / `suspended` |
| `last_login_at` | `TIMESTAMP` NULL | 마지막 로그인 |
| `created_at` | `TIMESTAMP` | 생성일시 |
| `updated_at` | `TIMESTAMP` | 수정일시 |
| `deleted_at` | `TIMESTAMP` NULL | 삭제일시 (soft delete) |

#### 4.7.2 API 엔드포인트

| 메서드 | 경로 | 설명 | 권한 |
|--------|------|------|------|
| `GET` | `/api/v1/users` | 사용자 목록 | Tenant Admin+ |
| `GET` | `/api/v1/users/:id` | 사용자 상세 | Tenant Admin+ or 본인 |
| `POST` | `/api/v1/users` | 사용자 생성 | Tenant Admin+ |
| `PUT` | `/api/v1/users/:id` | 사용자 수정 | Tenant Admin+ or 본인 |
| `DELETE` | `/api/v1/users/:id` | 사용자 삭제 | Tenant Admin+ |
| `PUT` | `/api/v1/users/:id/password` | 비밀번호 변경 | 본인 |
| `PUT` | `/api/v1/users/:id/role` | 역할 변경 | Tenant Admin+ |

---

## 5. 데이터베이스 스키마

### 5.1 ERD 관계도

```
┌──────────┐     1:N     ┌──────────┐     1:N     ┌──────────────┐
│  Tenant  │────────────►│  User    │             │ ClusterEvent │
└────┬─────┘             └──────────┘             └──────▲───────┘
     │                                                    │ 1:N
     │ 1:N                                               │
     │                                                    │
     ▼                                                    │
┌──────────┐     1:N     ┌──────────┐                    │
│ Cluster  │────────────►│  Node    │                    │
└────┬─────┘             └──────────┘                    │
     │                                                    │
     └────────────────────────────────────────────────────┘
     │
     │ 1:N
     ▼
┌──────────┐
│ AuditLog │
└──────────┘
```

**관계 요약:**
- Tenant (1) → (N) User
- Tenant (1) → (N) Cluster
- Cluster (1) → (N) Node
- Cluster (1) → (N) ClusterEvent
- Tenant (1) → (N) AuditLog

### 5.2 테이블 — clusters

| 필드 | 타입 | 제약조건 | 설명 |
|------|------|----------|------|
| `id` | `UUID` | PK | 클러스터 고유 ID |
| `tenant_id` | `UUID` | FK → tenants.id, NOT NULL | 소속 테넌트 |
| `name` | `VARCHAR(100)` | NOT NULL | 클러스터 이름 |
| `kubernetes_version` | `VARCHAR(20)` | NOT NULL | K8s 버전 |
| `status` | `VARCHAR(20)` | NOT NULL, DEFAULT 'pending' | 현재 상태 |
| `cp_count` | `INT` | NOT NULL | Control Plane 노드 수 |
| `worker_count` | `INT` | NOT NULL | Worker 노드 수 |
| `cp_flavor` | `VARCHAR(100)` | NOT NULL | CP Flavor |
| `worker_flavor` | `VARCHAR(100)` | NOT NULL | Worker Flavor |
| `os_image` | `VARCHAR(100)` | NOT NULL | OS 이미지 이름 |
| `network_id` | `VARCHAR(64)` | NOT NULL | OpenStack Network ID |
| `subnet_id` | `VARCHAR(64)` | NOT NULL | OpenStack Subnet ID |
| `external_network_id` | `VARCHAR(64)` | NOT NULL | 외부 네트워크 ID |
| `ssh_key_name` | `VARCHAR(100)` | NOT NULL | SSH 키페어 |
| `pod_cidr` | `VARCHAR(50)` | NOT NULL, DEFAULT '10.244.0.0/16' | Pod CIDR |
| `service_cidr` | `VARCHAR(50)` | NOT NULL, DEFAULT '10.96.0.0/12' | Service CIDR |
| `cni_plugin` | `VARCHAR(20)` | NOT NULL, DEFAULT 'calico' | CNI |
| `dns_nameservers` | `JSONB` | DEFAULT '["8.8.8.8"]' | DNS 서버 목록 |
| `api_endpoint` | `VARCHAR(255)` | NULL | K8s API 엔드포인트 (프로비저닝 후 설정) |
| `capi_namespace` | `VARCHAR(100)` | NOT NULL | Management Cluster 내 네임스페이스 |
| `capi_manifest_snapshot` | `JSONB` | NULL | 적용된 CAPI 매니페스트 스냅샷 |
| `error_message` | `TEXT` | NULL | 실패 시 오류 상세 |
| `created_by` | `UUID` | FK → users.id | 생성자 |
| `created_at` | `TIMESTAMP` | NOT NULL, DEFAULT NOW() | 생성일시 |
| `updated_at` | `TIMESTAMP` | NOT NULL, DEFAULT NOW() | 수정일시 |
| `deleted_at` | `TIMESTAMP` | NULL | 삭제일시 (soft delete) |

**인덱스:**
- `UNIQUE (tenant_id, name) WHERE deleted_at IS NULL` — 테넌트 내 이름 유니크
- `INDEX (tenant_id, status)` — 테넌트별 상태 필터링
- `INDEX (created_at DESC)` — 최신순 정렬

### 5.3 테이블 — nodes

| 필드 | 타입 | 제약조건 | 설명 |
|------|------|----------|------|
| `id` | `UUID` | PK | 노드 고유 ID |
| `cluster_id` | `UUID` | FK → clusters.id, NOT NULL | 소속 클러스터 |
| `name` | `VARCHAR(100)` | NOT NULL | 노드 이름 (Machine 이름) |
| `role` | `VARCHAR(20)` | NOT NULL | `control_plane` / `worker` |
| `status` | `VARCHAR(20)` | NOT NULL | `provisioning` / `running` / `failed` / `deleting` |
| `machine_id` | `VARCHAR(100)` | NULL | CAPI Machine 이름 |
| `openstack_server_id` | `VARCHAR(64)` | NULL | Nova Server ID |
| `private_ip` | `VARCHAR(45)` | NULL | 내부 IP |
| `floating_ip` | `VARCHAR(45)` | NULL | Floating IP (CP만) |
| `flavor` | `VARCHAR(100)` | NOT NULL | Flavor |
| `ready` | `BOOLEAN` | NOT NULL, DEFAULT false | Ready 여부 |
| `kubernetes_version` | `VARCHAR(20)` | NULL | 노드의 K8s 버전 |
| `created_at` | `TIMESTAMP` | NOT NULL | 생성일시 |
| `updated_at` | `TIMESTAMP` | NOT NULL | 수정일시 |

### 5.4 테이블 — cluster_events

| 필드 | 타입 | 제약조건 | 설명 |
|------|------|----------|------|
| `id` | `UUID` | PK | 이벤트 ID |
| `cluster_id` | `UUID` | FK → clusters.id, NOT NULL | 소속 클러스터 |
| `type` | `VARCHAR(50)` | NOT NULL | 이벤트 타입 |
| `severity` | `VARCHAR(20)` | NOT NULL | `info` / `warning` / `error` |
| `message` | `TEXT` | NOT NULL | 이벤트 메시지 |
| `source` | `VARCHAR(100)` | NULL | 이벤트 소스 (CAPI/CAPO/System) |
| `metadata` | `JSONB` | NULL | 추가 메타데이터 |
| `created_at` | `TIMESTAMP` | NOT NULL | 발생 시간 |

**인덱스:**
- `INDEX (cluster_id, created_at DESC)` — 클러스터별 최신 이벤트

### 5.5 마이그레이션 SQL 예시

```sql
-- migrations/001_create_tenants.up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    os_auth_url VARCHAR(255) NOT NULL,
    os_project_id VARCHAR(64) NOT NULL,
    os_project_name VARCHAR(100) NOT NULL,
    os_domain_name VARCHAR(100) NOT NULL DEFAULT 'Default',
    os_credentials TEXT NOT NULL,  -- AES-256-GCM encrypted
    max_clusters INT NOT NULL DEFAULT 10,
    max_nodes_per_cluster INT NOT NULL DEFAULT 20,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;
```

---

## 6. UI/UX 설계

### 6.1 주요 화면 구성

| 화면 | 경로 | 주요 기능 |
|------|------|-----------|
| 로그인 | `/login` | 이메일/비밀번호 입력, OIDC 로그인 버튼 |
| 대시보드 | `/dashboard` | 클러스터 요약 카드, 상태 차트, 최근 이벤트 |
| 클러스터 목록 | `/clusters` | 테이블 뷰 + 카드 뷰 전환, 필터/검색, 상태 배지 |
| 클러스터 생성 | `/clusters/new` | 4단계 스텝 위저드 |
| 클러스터 상세 | `/clusters/:id` | 탭 구성: 개요/노드/이벤트/설정 |
| 클러스터 스케일링 | `/clusters/:id/scale` | CP/Worker 노드 수 조절 슬라이더 |
| 클러스터 업그레이드 | `/clusters/:id/upgrade` | 버전 선택, Rolling Update 진행률 |
| 테넌트 관리 | `/admin/tenants` | 테넌트 CRUD, 쿼터 설정 |
| 사용자 관리 | `/admin/users` | 사용자 목록, 역할 변경, 초대 |
| 감사 로그 | `/admin/audit-logs` | 작업 이력 검색/필터링, 상세 보기 |
| 설정 | `/settings` | 프로필, 비밀번호 변경, API 토큰 |

### 6.2 클러스터 생성 위저드 상세

클러스터 생성은 4단계 스텝 위저드로 진행된다:

#### Step 1 — 기본 정보
- 클러스터 이름 (텍스트 입력, 실시간 유효성 검증)
- Kubernetes 버전 선택 (드롭다운, 지원 버전 목록)
- CNI 플러그인 선택 (라디오 버튼: Calico / Cilium)
- Pod CIDR, Service CIDR (텍스트 입력, 기본값 프리필)

#### Step 2 — 네트워크 설정
- OpenStack Network 선택 (드롭다운, API로 동적 로드)
- Subnet 선택 (Network 선택 시 자동 필터링)
- External Network 선택 (드롭다운)
- DNS 서버 (태그 입력, 기본값 8.8.8.8)

#### Step 3 — 노드 구성
- Control Plane 노드 수 (1/3/5 선택 버튼)
- Control Plane Flavor (드롭다운, CPU/RAM/Disk 표시)
- Worker 노드 수 (숫자 입력 또는 슬라이더)
- Worker Flavor (드롭다운)
- OS 이미지 (드롭다운, K8s 버전 호환 필터)
- SSH 키페어 (드롭다운)

#### Step 4 — 검토 및 생성
- 전체 설정 요약 테이블
- 예상 리소스 사용량 (vCPU, RAM, Disk 합계)
- 현재 쿼터 대비 사용량 표시 (프로그레스 바)
- "클러스터 생성" 버튼

**위저드 UX 요구사항:**
- 각 스텝은 이전 스텝 완료 후에만 진행 가능
- "이전" / "다음" 버튼으로 스텝 간 이동
- 각 스텝에서 실시간 유효성 검증 (Zod)
- 드롭다운 데이터는 TanStack Query로 비동기 로드 (로딩 스피너 표시)
- 폼 데이터는 스텝 간 유지 (Zustand 또는 React Hook Form의 FormProvider)

### 6.3 대시보드 화면 구성

**상단 요약 카드 (4개):**
- 총 클러스터 수 / Running 수
- 총 노드 수 / Ready 수
- 리소스 사용량 (vCPU, RAM)
- 최근 알림 수

**중앙 영역:**
- 클러스터 상태 파이 차트 (Running/Provisioning/Failed/Updating)
- 클러스터별 노드 수 바 차트

**하단 영역:**
- 최근 이벤트 타임라인 (최근 20개)
- 최근 생성/삭제된 클러스터 목록

### 6.4 클러스터 상세 화면 구성

**헤더:**
- 클러스터 이름, 상태 배지, Kubernetes 버전
- 액션 버튼: 스케일링, 업그레이드, Kubeconfig 다운로드, 삭제

**탭 구성:**
1. **개요 탭:** API 엔드포인트, 네트워크 정보, 생성일, 생성자
2. **노드 탭:** 노드 테이블 (이름, 역할, 상태, IP, Flavor, 버전)
3. **이벤트 탭:** 이벤트 타임라인 (실시간 WebSocket 업데이트)
4. **설정 탭:** CAPI 매니페스트 스냅샷 뷰어 (YAML 하이라이팅)

---

## 7. 비기능 요구사항

### 7.1 보안

- HTTPS 전용 통신 (TLS 1.3)
- OpenStack 자격증명은 AES-256-GCM으로 암호화하여 DB 저장
- 비밀번호: bcrypt 해싱 (cost factor 12)
- Kubeconfig는 발급 시 사용자별 RBAC 적용, 제한된 유효기간 설정
- API Rate Limiting 적용 (Kong/Nginx)
  - 인증 API: 5 req/min per IP
  - 일반 API: 100 req/min per user
  - 클러스터 생성: 5 req/min per tenant
- CORS 정책 설정 (허용된 도메인만 접근)
- SQL Injection / XSS 방지 (ORM 사용, 입력값 sanitize)
- Audit Log를 통한 전체 작업 추적
- 민감 정보 로그 마스킹

### 7.2 성능

- 클러스터 목록 조회: 응답 시간 < 500ms (100개 클러스터 기준)
- 클러스터 생성 요청: API 응답 < 2초 (비동기 프로비저닝 시작)
- 동시 클러스터 프로비저닝: 10개 이상 병렬 처리
- WebSocket 연결: 클러스터당 최대 100개 동시 연결 지원
- 데이터베이스: Connection Pool 최적화 (pgx pool, max 25 connections)
- OpenStack API 호출: Redis 캐시로 반복 호출 최소화
- 프론트엔드: Lighthouse Performance Score > 80

### 7.3 가용성

- CMP 애플리케이션 자체를 Kubernetes에 배포하여 HA 구성
- API 서버 복수 레플리카 + 로드밸런싱
- PostgreSQL: Primary-Replica 구성 또는 CloudNativePG 사용
- Redis: Sentinel 또는 Cluster 모드
- Graceful Shutdown 구현 (진행 중 작업 완료 후 종료)
- Health Check 엔드포인트: `GET /healthz` (liveness), `GET /readyz` (readiness)

### 7.4 테스트 요구사항

| 테스트 유형 | 도구 | 커버리지 목표 |
|-------------|------|---------------|
| Unit Test (Go) | `testing` + `testify` + `gomock` | 비즈니스 로직 80%+ |
| Integration Test | `envtest` (sigs.k8s.io/controller-runtime) | CAPI 매니페스트 생성/적용 |
| E2E Test | 실제 OpenStack + CAPI 환경 | 클러스터 생성~삭제 전체 사이클 |
| Frontend Test | `vitest` + `@testing-library/react` | 주요 컴포넌트 70%+ |
| API Test | REST Client / Postman 컬렉션 | 전체 엔드포인트 |
| Load Test | k6 | 동시 사용자 50명 기준 |

---

## 8. 개발 단계 및 마일스톤

| 단계 | 기간 | 주요 산출물 |
|------|------|-------------|
| **Phase 1:** 기반 구축 | 4주 | 프로젝트 스캐폴딩, DB 스키마/마이그레이션, 인증/인가, 테넌트 CRUD |
| **Phase 2:** 핵심 기능 | 6주 | CAPI/CAPO 연동, 매니페스트 자동생성, 클러스터 CRUD, OpenStack 연동 |
| **Phase 3:** 운영 기능 | 4주 | Scaling, Upgrade, 모니터링, WebSocket 실시간 스트림, 감사로그 |
| **Phase 4:** UI/UX | 4주 | 대시보드, 클러스터 생성 위저드, 상세 화면, 관리 페이지 |
| **Phase 5:** 보안/안정성 | 3주 | 보안 강화, HA 구성, 성능 최적화, 에러 핸들링 고도화 |
| **Phase 6:** 테스트/배포 | 3주 | E2E 테스트, Helm Chart, CI/CD 파이프라인, 문서화 |

**총 예상 기간: 약 24주 (6개월)**

---

## 9. Claude Code 지시 가이드

본 명세서를 바탕으로 Claude Code에게 지시할 때의 프롬프트 예시입니다.

### 9.1 단계별 프롬프트 예시

#### ▶ 1단계: 프로젝트 초기화

```
SPEC.md의 섹션 3(Project Structure)을 참고하여 Go 백엔드 프로젝트를 초기화해줘.

구체적으로:
1. go mod init github.com/{org}/k8s-cmp 실행
2. 섹션 2.1의 모든 Go 의존성 설치
3. 섹션 3.1의 디렉토리 구조 생성 (빈 .go 파일에 package 선언만)
4. cmd/api-server/main.go에 Gin 서버 부팅 코드 작성
   - viper로 configs/config.yaml 읽기
   - zap 로거 초기화
   - PostgreSQL (GORM) 연결
   - Redis 연결
   - NATS 연결
   - 라우터 설정 (빈 핸들러)
   - Graceful Shutdown
5. configs/config.example.yaml 작성 (SPEC.md §3.2 참고)
6. deploy/docker-compose.yaml 작성 (PostgreSQL, Redis, NATS)
7. Makefile 작성 (build, run, test, migrate, docker-up, docker-down)
```

#### ▶ 2단계: DB 스키마 및 마이그레이션

```
SPEC.md 섹션 5를 참고하여 DB 스키마를 구현해줘.

구체적으로:
1. migrations/ 디렉토리에 up/down SQL 파일 생성 (001~006)
   - SPEC.md §5.2~5.4의 모든 테이블, 인덱스, 제약조건 포함
   - §5.5의 SQL 예시 형식 따르기
2. internal/domain/ 에 Go 구조체 정의
   - GORM 태그 포함 (테이블명, 컬럼명, 제약조건)
   - JSON 시리얼라이즈 태그
   - 모든 엔티티에 soft delete 지원 (gorm.DeletedAt)
3. cmd/migrator/main.go 구현
   - golang-migrate 사용
   - up, down, version 서브커맨드
4. internal/repository/interfaces.go에 모든 Repository 인터페이스 정의
5. internal/repository/ 에 GORM 기반 구현체 작성
```

#### ▶ 3단계: 인증/인가

```
SPEC.md 섹션 4.1을 참고하여 인증/인가 시스템을 구현해줘.

구체적으로:
1. internal/service/auth_service.go
   - Login: 이메일/비밀번호 검증 → Access Token + Refresh Token 발급
   - Refresh: Refresh Token 검증 → 새 토큰 쌍 발급
   - Logout: Redis에서 Refresh Token 삭제, 블랙리스트 추가
   - JWT Payload는 §4.1.3 구조 따르기
2. internal/handler/auth_handler.go
   - §4.1.2의 4개 엔드포인트 구현
   - pkg/response 래퍼 사용
3. internal/middleware/auth.go — JWT 검증 미들웨어
4. internal/middleware/rbac.go — §4.1.4 권한 매트릭스 구현
   - RequireRole(roles ...string) 미들웨어 팩토리
5. internal/middleware/tenant.go — 테넌트 컨텍스트 주입
6. internal/middleware/audit.go — 감사 로그 자동 기록
   - §4.6.1의 기록 대상 액션에 해당하는 요청 자동 기록
   - 민감정보 마스킹 적용
7. 미들웨어 체인: §4.1.3의 순서 따르기
```

#### ▶ 4단계: 테넌트 관리

```
SPEC.md 섹션 4.2를 참고하여 테넌트 관리 CRUD를 구현해줘.

구체적으로:
1. internal/service/tenant_service.go
   - CRUD + OpenStack 연결 테스트 + 멤버 관리
   - 생성 시 os_credentials를 pkg/crypto/aes.go로 암호화
   - 조회 시 os_credentials 복호화 (응답에는 마스킹)
2. internal/handler/tenant_handler.go
   - §4.2.3의 모든 엔드포인트 구현
   - 입력 유효성 검증 (validator 태그)
3. pkg/crypto/aes.go — AES-256-GCM 암호화/복호화 유틸
4. POST /tenants/:id/verify-connection 구현
   - gophercloud로 Keystone 인증 시도
   - 성공/실패 응답
```

#### ▶ 5단계: OpenStack 리소스 연동

```
SPEC.md 섹션 4.4를 참고하여 OpenStack 리소스 연동 모듈을 구현해줘.

구체적으로:
1. internal/openstack/client.go
   - §4.4.4의 NewClient 함수: 테넌트 자격증명으로 gophercloud 클라이언트 생성
   - 클라이언트 캐싱 (테넌트별 1개)
2. internal/openstack/compute.go — Flavors, Keypairs 조회
3. internal/openstack/network.go — Networks, Subnets, Security Groups, External Networks, Floating IPs
4. internal/openstack/image.go — Images 조회 (§4.4.2의 필터링 규칙 적용)
5. internal/openstack/quota.go — Quotas 조회
6. internal/service/openstack_service.go
   - Redis 캐싱 적용 (§4.4.2의 TTL 따르기)
   - 캐시 키: openstack:{tenant_id}:{resource_type}
7. internal/handler/openstack_handler.go — §4.4.3의 모든 엔드포인트
```

#### ▶ 6단계: CAPI/CAPO 매니페스트 생성 엔진

```
SPEC.md 섹션 4.3.1~4.3.2를 참고하여 CAPI/CAPO 매니페스트 생성 엔진을 구현해줘.

구체적으로:
1. internal/k8s/manifest_templates/ 디렉토리에 Go 템플릿 파일 작성
   - §4.3.2 테이블의 7개 리소스 각각에 대한 .yaml.tmpl 파일
   - CAPI v1beta1 스키마 준수
   - OpenStack clouds.yaml Secret 템플릿
2. internal/k8s/manifest_generator.go
   - §4.3.2의 ClusterCreateParams 구조체 정의
   - GenerateManifests 함수: 템플릿 렌더링 → []unstructured.Unstructured 반환
   - 각 리소스의 이름 규칙: {cluster_name}-{suffix}
3. internal/k8s/client.go
   - Management Cluster kubeconfig 로드
   - dynamic.Interface 생성
4. internal/k8s/cluster_manager.go
   - ApplyManifests: §4.3.2의 매니페스트 Apply 프로세스 9단계 구현
   - DeleteCluster: 클러스터 관련 모든 CAPI 리소스 삭제
5. internal/k8s/kubeconfig.go
   - Workload Cluster의 kubeconfig Secret에서 추출
   - Secret 이름: {cluster_name}-kubeconfig
```

#### ▶ 7단계: 클러스터 라이프사이클 API

```
SPEC.md 섹션 4.3.3~4.3.6을 참고하여 클러스터 라이프사이클 API를 구현해줘.

구체적으로:
1. internal/service/cluster_service.go
   - CreateCluster: 유효성 검증 → 쿼터 확인 → 매니페스트 생성 → Apply → DB 저장 → NATS 발행
   - DeleteCluster: 상태 확인 → CAPI 리소스 삭제 → DB 상태 업데이트
   - ScaleCluster: §4.3.5 구현 (MachineDeployment/KCP replicas 변경)
   - UpgradeCluster: §4.3.6 구현 (CP → Worker 순차 업그레이드)
   - GetKubeconfig: Secret에서 kubeconfig 추출
   - RetryCluster: Failed 클러스터 재시도
2. internal/handler/cluster_handler.go — §4.3.4의 모든 엔드포인트
   - 목록 조회: 페이지네이션, 필터링, 정렬 파라미터 처리
3. cmd/worker/main.go
   - NATS Consumer로 비동기 작업 처리
   - 클러스터 프로비저닝 모니터링 (상태 폴링 → DB 업데이트)
```

#### ▶ 8단계: 모니터링/WebSocket

```
SPEC.md 섹션 4.5를 참고하여 모니터링 및 WebSocket 실시간 스트림을 구현해줘.

구체적으로:
1. internal/k8s/watcher.go — §4.5.2 구현
   - Kubernetes Informer로 CAPI 리소스 변경 감지
   - 변경 감지 시 WebSocket Hub로 이벤트 전송
   - DB의 클러스터/노드 상태 동기화
2. internal/websocket/hub.go — §4.5.4의 Hub 구현
   - 클러스터별 구독 관리
   - JWT 인증 (쿼리 파라미터)
3. internal/websocket/client.go — 개별 WebSocket 클라이언트 연결
4. internal/websocket/events.go — §4.5.4의 이벤트 타입 정의
5. internal/handler/websocket_handler.go
   - /ws/v1/clusters/:id/stream 엔드포인트
   - /ws/v1/clusters/stream 엔드포인트 (대시보드용)
6. internal/service/notification_service.go — §4.5.3의 알림 체계 구현
```

#### ▶ 9단계: Frontend UI

```
SPEC.md 섹션 6을 참고하여 Next.js 프론트엔드를 구현해줘.

구체적으로:
1. web/ 디렉토리에 Next.js 14 프로젝트 초기화
   - TypeScript strict mode, Tailwind CSS, shadcn/ui 설정
   - §2.2의 모든 프론트엔드 라이브러리 설치
2. 레이아웃 구현 (§6.1 참고)
   - Sidebar: 네비게이션 메뉴 (대시보드, 클러스터, 관리, 설정)
   - Header: 사용자 정보, 테넌트 선택, 알림
3. 로그인 페이지 구현
4. 대시보드 구현 (§6.3 참고)
   - 요약 카드 4개, 상태 파이 차트, 최근 이벤트
5. 클러스터 목록 페이지 (테이블 뷰 + 카드 뷰)
6. 클러스터 생성 위저드 (§6.2 상세 구현)
   - 4단계 스텝, React Hook Form + Zod 검증
   - OpenStack 리소스 동적 로드 (TanStack Query)
7. 클러스터 상세 페이지 (§6.4 참고)
   - 4개 탭 구현, WebSocket 실시간 업데이트
8. 관리 페이지: 테넌트, 사용자, 감사 로그
9. web/src/lib/api-client.ts — Axios 인스턴스, JWT 인터셉터
10. web/src/lib/websocket.ts — WebSocket 연결 관리
```

#### ▶ 10단계: 테스트 및 배포

```
SPEC.md 섹션 7.4와 8을 참고하여 테스트와 배포 환경을 구성해줘.

구체적으로:
1. 주요 서비스 Unit Test 작성
   - auth_service_test.go, cluster_service_test.go, tenant_service_test.go
   - testify + gomock 사용, 테이블 드리븐 방식
2. CAPI 매니페스트 생성 테스트
   - manifest_generator_test.go: 입력값 → 올바른 YAML 생성 검증
3. API Integration Test
   - httptest로 핸들러 테스트 (DB mock)
4. deploy/docker/Dockerfile.api — 멀티스테이지 빌드
5. deploy/docker/Dockerfile.web — Next.js 빌드
6. deploy/helm/k8s-cmp/ — Helm Chart
   - Deployment, Service, Ingress, ConfigMap, Secret
   - values.yaml에 주요 설정
7. .github/workflows/ci.yaml — CI 파이프라인
   - lint (golangci-lint), test, build, docker push
8. Health Check 엔드포인트: GET /healthz, GET /readyz
```

---

## 10. API 응답 형식 표준

### 10.1 성공 응답

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "my-cluster",
    "status": "running"
  }
}
```

### 10.2 목록 응답 (페이지네이션)

```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_items": 42,
      "total_pages": 3
    }
  }
}
```

### 10.3 에러 응답

```json
{
  "success": false,
  "error": {
    "code": "CLUSTER_NOT_FOUND",
    "message": "클러스터를 찾을 수 없습니다.",
    "details": {}
  }
}
```

### 10.4 에러 코드 목록

| 코드 | HTTP 상태 | 설명 |
|------|-----------|------|
| `UNAUTHORIZED` | 401 | 인증 실패 |
| `FORBIDDEN` | 403 | 권한 부족 |
| `NOT_FOUND` | 404 | 리소스 없음 |
| `VALIDATION_ERROR` | 422 | 입력값 유효성 오류 |
| `DUPLICATE_ERROR` | 409 | 중복 (이름 등) |
| `QUOTA_EXCEEDED` | 429 | 쿼터 초과 |
| `OPENSTACK_ERROR` | 502 | OpenStack API 오류 |
| `CAPI_ERROR` | 502 | CAPI 작업 오류 |
| `INTERNAL_ERROR` | 500 | 내부 서버 오류 |
