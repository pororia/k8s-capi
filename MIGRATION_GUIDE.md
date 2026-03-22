# K8s CMP 이관 가이드 (Windows / Linux)

이 문서는 K8s CMP(Kubernetes Cluster Management Platform) 프로젝트를
**Windows** 또는 **Linux** 환경으로 이관하고 실행하는 방법을 설명합니다.

---

## 목차

1. [사전 요구사항](#1-사전-요구사항)
2. [소스코드 이관](#2-소스코드-이관)
3. [환경 구성 (공통)](#3-환경-구성-공통)
4. [실행 방법 A — Docker Compose (권장)](#4-실행-방법-a--docker-compose-권장)
5. [실행 방법 B — 직접 빌드 및 실행](#5-실행-방법-b--직접-빌드-및-실행)
6. [Windows 특이사항](#6-windows-특이사항)
7. [Linux 특이사항](#7-linux-특이사항)
8. [초기 관리자 계정 생성](#8-초기-관리자-계정-생성)
9. [동작 확인](#9-동작-확인)
10. [트러블슈팅](#10-트러블슈팅)

---

## 1. 사전 요구사항

### 공통 (Windows / Linux 모두)

| 소프트웨어 | 최소 버전 | 용도 |
|---|---|---|
| Docker Desktop / Docker Engine | 24+ | 인프라 컨테이너 실행 |
| Docker Compose | 2.20+ | 멀티 컨테이너 오케스트레이션 |
| Go | 1.22+ | 백엔드 빌드 (직접 빌드 시) |
| Node.js | 20 LTS | 프론트엔드 빌드 (직접 빌드 시) |
| Git | 2.40+ | 소스코드 관리 |
| openssl | 3.x | 암호화 키 생성 |

### Management Cluster kubeconfig
- CAPI/CAPO 컨트롤러가 설치된 Management Cluster의 `kubeconfig` 파일이 필요합니다.
- OpenStack에 접근 가능한 네트워크 환경이어야 합니다.

---

## 2. 소스코드 이관

### 방법 1 — Git 클론 (원격 저장소가 있는 경우)

```bash
git clone https://your-git-server/k8s-cmp.git
cd k8s-cmp
```

### 방법 2 — 파일 직접 복사

현재 Mac에서 압축 후 전송:

```bash
# Mac에서 실행
cd /Users/bongsookim/project
tar -czf k8s-capi.tar.gz k8s-capi/

# 대상 서버로 전송 (SCP 예시)
scp k8s-capi.tar.gz user@target-server:/opt/

# 대상 서버에서 압축 해제
ssh user@target-server
cd /opt
tar -xzf k8s-capi.tar.gz
cd k8s-capi
```

> **Windows로 전송 시:** WinSCP, FileZilla, 또는 `scp` (PowerShell/WSL) 사용

---

## 3. 환경 구성 (공통)

### 3-1. .env 파일 생성

프로젝트 루트에 `.env` 파일을 생성합니다.

```bash
cp .env.example .env
```

`.env` 파일을 열어 아래 값을 채웁니다:

```env
# JWT 시크릿 (최소 32자 이상 랜덤 문자열)
AUTH_JWT_SECRET=여기에_랜덤_문자열_입력

# AES-256-GCM 암호화 키 (32바이트 base64 인코딩)
ENCRYPTION_KEY=여기에_base64_키_입력

# Management Cluster kubeconfig 경로
KUBECONFIG=/path/to/kubeconfig
```

**키 생성 방법:**

```bash
# Linux / Mac / Windows(Git Bash)
openssl rand -hex 32          # → AUTH_JWT_SECRET 값으로 사용
openssl rand -base64 32       # → ENCRYPTION_KEY 값으로 사용
```

> **Windows PowerShell에서 키 생성:**
> ```powershell
> # PowerShell (openssl 없는 경우)
> $bytes = New-Object byte[] 32
> [System.Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
> [Convert]::ToBase64String($bytes)   # → ENCRYPTION_KEY
> -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})  # → AUTH_JWT_SECRET
> ```

### 3-2. config.yaml 확인

`configs/config.yaml` 파일은 기본 설정 파일입니다.
Docker Compose 실행 시 환경변수가 우선 적용되므로 별도 수정 불필요합니다.

**직접 실행 시에는 로컬 환경에 맞게 수정:**

```yaml
database:
  host: "localhost"
  port: 5432        # Docker Compose 포트와 일치시킬 것
  name: "k8s_cmp"
  user: "cmp"
  password: "cmp_password"

redis:
  host: "localhost"
  port: 6379

nats:
  url: "nats://localhost:4222"

auth:
  jwt_secret: "your-jwt-secret"   # .env의 AUTH_JWT_SECRET 값

encryption:
  key: "your-base64-key"          # .env의 ENCRYPTION_KEY 값
```

### 3-3. Management Cluster kubeconfig 준비

```bash
# kubeconfig 파일을 프로젝트 내 안전한 위치에 복사
mkdir -p ~/.kube
cp /path/to/management-kubeconfig ~/.kube/config

# .env에 경로 설정
echo "KUBECONFIG=~/.kube/config" >> .env
```

---

## 4. 실행 방법 A — Docker Compose (권장)

**이 방법은 Go/Node.js 설치 없이 Docker만으로 전체 스택을 실행합니다.**

### 4-1. 인프라만 실행 (개발/테스트용)

PostgreSQL, Redis, NATS만 컨테이너로 실행하고, 백엔드/프론트는 로컬에서 직접 실행합니다.

```bash
# 인프라 컨테이너 시작
docker compose -f deploy/docker-compose.yaml up -d

# 상태 확인
docker compose -f deploy/docker-compose.yaml ps

# 로그 확인
docker compose -f deploy/docker-compose.yaml logs -f
```

### 4-2. 전체 스택 실행 (운영용)

백엔드, 프론트엔드, 인프라 모두 컨테이너로 실행합니다.

```bash
# .env 파일 확인 후 실행
docker compose -f deploy/docker-compose.full.yaml up -d

# 빌드 포함 실행 (이미지 재빌드 필요 시)
docker compose -f deploy/docker-compose.full.yaml up -d --build

# 상태 확인
docker compose -f deploy/docker-compose.full.yaml ps

# 로그 확인 (전체)
docker compose -f deploy/docker-compose.full.yaml logs -f

# 로그 확인 (특정 서비스)
docker compose -f deploy/docker-compose.full.yaml logs -f api
docker compose -f deploy/docker-compose.full.yaml logs -f web
```

**포트 매핑 확인:**

| 서비스 | 내부 포트 | 외부 포트 | 접근 URL |
|---|---|---|---|
| API 서버 | 8080 | 8080 | http://localhost:8080 |
| 프론트엔드 | 3000 | 3000 | http://localhost:3000 |
| PostgreSQL | 5432 | 5433 | localhost:5433 |
| Redis | 6379 | 6379 | localhost:6379 |
| NATS | 4222 | 4222 | nats://localhost:4222 |
| NATS Monitor | 8222 | 8222 | http://localhost:8222 |

### 4-3. 종료

```bash
# 컨테이너 중지 (데이터 유지)
docker compose -f deploy/docker-compose.full.yaml down

# 컨테이너 + 볼륨 삭제 (데이터 초기화)
docker compose -f deploy/docker-compose.full.yaml down -v
```

---

## 5. 실행 방법 B — 직접 빌드 및 실행

Go와 Node.js를 직접 설치하여 네이티브로 빌드/실행합니다.

### 5-1. 인프라 컨테이너 시작

```bash
docker compose -f deploy/docker-compose.yaml up -d
```

### 5-2. Go 백엔드 빌드 및 실행

**Linux:**
```bash
# 모듈 정리
go mod tidy

# 빌드
go build -o bin/api-server ./cmd/api-server
go build -o bin/worker ./cmd/worker
go build -o bin/migrator ./cmd/migrator

# DB 마이그레이션
./bin/migrator -action=up

# API 서버 실행
./bin/api-server

# 워커 실행 (별도 터미널)
./bin/worker
```

**Windows (PowerShell):**
```powershell
# 빌드
go build -o bin\api-server.exe .\cmd\api-server
go build -o bin\worker.exe .\cmd\worker
go build -o bin\migrator.exe .\cmd\migrator

# DB 마이그레이션
.\bin\migrator.exe -action=up

# API 서버 실행
.\bin\api-server.exe

# 워커 실행 (별도 PowerShell 창)
.\bin\worker.exe
```

**또는 go run 사용 (빌드 없이 바로 실행):**

```bash
# 마이그레이션
go run ./cmd/migrator -action=up

# API 서버
go run ./cmd/api-server

# 워커 (별도 터미널)
go run ./cmd/worker
```

### 5-3. 프론트엔드 빌드 및 실행

```bash
cd web

# 패키지 설치
npm install

# 개발 서버 (핫리로드 지원)
npm run dev

# 운영 빌드 후 실행
npm run build
npm run start
```

---

## 6. Windows 특이사항

### 6-1. Docker Desktop 설정

- Docker Desktop 설치 후 **WSL2 backend** 활성화 권장
- Settings → Resources → WSL Integration → 사용하는 WSL 배포판 활성화

### 6-2. 경로 구분자

Windows에서 직접 빌드 시 경로 구분자에 주의합니다.

```powershell
# PowerShell에서 Go 빌드
$env:GOOS = "windows"
go build -o bin\api-server.exe .\cmd\api-server

# Linux용 바이너리 크로스 컴파일 (WSL/Linux 서버 배포용)
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o bin\api-server-linux .\cmd\api-server
```

### 6-3. kubeconfig 경로 (Windows)

`.env` 파일에서 Windows 경로 지정 시 슬래시 방향 주의:

```env
# Windows 경로 (Docker Compose에서는 슬래시 사용)
KUBECONFIG=C:/Users/username/.kube/config

# 또는 WSL 경로
KUBECONFIG=/mnt/c/Users/username/.kube/config
```

### 6-4. 포트 충돌 확인

```powershell
# 사용 중인 포트 확인
netstat -ano | findstr :8080
netstat -ano | findstr :3000
netstat -ano | findstr :5432
```

### 6-5. 방화벽 설정

Windows Defender 방화벽에서 아래 포트를 허용합니다:
- 8080 (API 서버)
- 3000 (프론트엔드)
- 4222 (NATS)

```powershell
# PowerShell (관리자 권한)
New-NetFirewallRule -DisplayName "K8s CMP API" -Direction Inbound -Port 8080 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "K8s CMP Web" -Direction Inbound -Port 3000 -Protocol TCP -Action Allow
```

### 6-6. Makefile 사용 (Windows)

Windows에서는 `make` 명령이 기본 제공되지 않습니다.

**방법 1 — Git Bash 또는 WSL 사용 (권장):**
```bash
# Git Bash 또는 WSL 터미널에서 실행
make docker-up
make migrate-up
make run
```

**방법 2 — Chocolatey로 make 설치:**
```powershell
# PowerShell (관리자 권한)
choco install make
```

**방법 3 — PowerShell에서 직접 명령 실행:**
```powershell
docker compose -f deploy/docker-compose.yaml up -d
go run ./cmd/migrator -action=up
go run ./cmd/api-server
```

---

## 7. Linux 특이사항

### 7-1. Docker 설치 (Docker Desktop 없는 서버 환경)

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# Docker Compose plugin 확인
docker compose version
```

### 7-2. Go 설치

```bash
# 최신 Go 1.22 설치
wget https://go.dev/dl/go1.22.10.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.10.linux-amd64.tar.gz

# PATH 설정 (~/.bashrc 또는 ~/.zshrc에 추가)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

go version
```

### 7-3. Node.js 설치

```bash
# nvm 사용 (권장)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc
nvm install 20
nvm use 20
node --version
```

### 7-4. openssl 키 생성

```bash
# JWT Secret
openssl rand -hex 32

# AES-256 Encryption Key
openssl rand -base64 32
```

### 7-5. systemd 서비스 등록 (운영 환경)

백엔드를 systemd 서비스로 등록하면 서버 재시작 후에도 자동으로 실행됩니다.

```bash
# /etc/systemd/system/k8s-cmp-api.service
sudo tee /etc/systemd/system/k8s-cmp-api.service > /dev/null <<EOF
[Unit]
Description=K8s CMP API Server
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/k8s-capi
EnvironmentFile=/opt/k8s-capi/.env
ExecStart=/opt/k8s-capi/bin/api-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable k8s-cmp-api
sudo systemctl start k8s-cmp-api
sudo systemctl status k8s-cmp-api
```

### 7-6. 포트 확인

```bash
# 사용 중인 포트 확인
ss -tlnp | grep -E '8080|3000|5432|6379|4222'

# 방화벽 설정 (ufw)
sudo ufw allow 8080/tcp
sudo ufw allow 3000/tcp
```

---

## 8. 초기 관리자 계정 생성

마이그레이션 후 DB에 직접 Super Admin 계정을 INSERT 해야 합니다.

### 8-1. bcrypt 해시 생성

```bash
# 프로젝트 루트에서 실행
go run ./cmd/tools/genhash/main.go
# 출력된 해시 값을 복사
```

> `main.go` 내 `password` 변수값을 원하는 비밀번호로 변경 후 실행

### 8-2. 관리자 계정 INSERT

```bash
# PostgreSQL 접속 (Docker Compose 환경)
docker exec -it k8s-cmp-postgres psql -U cmp -d k8s_cmp

# 직접 실행 환경
psql -h localhost -p 5433 -U cmp -d k8s_cmp
```

```sql
INSERT INTO users (
  id, email, password_hash, name, role, status, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'admin@example.com',
  '$2a$10$여기에_genhash_출력값_입력',
  'Super Admin',
  'super_admin',
  'active',
  NOW(),
  NOW()
);
```

---

## 9. 동작 확인

### 9-1. API 서버 헬스체크

```bash
curl http://localhost:8080/health
# 기대 응답: {"status":"ok"} 또는 200 OK
```

### 9-2. 로그인 테스트

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"your-password"}'
```

### 9-3. 프론트엔드 접속

브라우저에서 `http://localhost:3000` 접속 후 로그인 화면 확인

### 9-4. 전체 서비스 상태 확인

```bash
# Docker Compose 상태
docker compose -f deploy/docker-compose.full.yaml ps

# 기대 출력 (모든 서비스 Running)
# postgres   running (healthy)
# redis      running (healthy)
# nats       running (healthy)
# api        running
# worker     running
# web        running
```

---

## 10. 트러블슈팅

### 포트 충돌 (`port is already allocated`)

로컬에 PostgreSQL이 이미 5432로 실행 중인 경우:

```bash
# deploy/docker-compose.yaml 의 postgres 포트를 변경
# "5433:5432" → 로컬 포트를 5433으로 사용

# configs/config.yaml 의 database.port 도 동일하게 변경
database:
  port: 5433
```

### AES 키 오류 (`AES key must be 32 bytes`)

`ENCRYPTION_KEY`가 base64 디코딩 후 정확히 32바이트여야 합니다.

```bash
# 올바른 키 생성 방법
openssl rand -base64 32
# 예시 출력: TSda0uQE9A4GQG5/J51pkeB5i0spb0Ci9sHpZHzGkRU=
```

### 마이그레이션 실패 (DB 연결 오류)

```bash
# DB 컨테이너가 완전히 뜰 때까지 대기 후 재시도
docker compose -f deploy/docker-compose.yaml ps
# postgres가 healthy 상태인지 확인

# 수동 마이그레이션 재실행
go run ./cmd/migrator -action=up
```

### Docker 이미지 빌드 오류 (Windows)

```powershell
# Docker Desktop에서 WSL2 backend 사용 중인지 확인
docker info | findstr -i "Operating System"

# 빌드 캐시 초기화 후 재시도
docker builder prune -f
docker compose -f deploy/docker-compose.full.yaml up -d --build
```

### OpenStack API 연결 실패

OpenStack 엔드포인트(`os_auth_url`)가 이 서버에서 접근 가능한지 확인합니다:

```bash
curl -v --max-time 10 https://your-openstack-auth-url:5000/v3
```

응답이 없으면 VPN 연결이 필요하거나 OpenStack 네트워크와 동일한 망에서 실행해야 합니다.

### 프론트엔드 API 연결 오류 (CORS)

`deploy/docker-compose.full.yaml` 또는 `configs/config.yaml`의 `cors.allowed_origins`에 프론트엔드 접속 URL을 추가합니다:

```yaml
cors:
  allowed_origins:
    - "http://localhost:3000"
    - "http://your-server-ip:3000"
```

또는 환경변수로:
```env
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://your-server-ip:3000
```

---

## 빠른 시작 요약

```bash
# 1. 소스코드 가져오기
git clone https://your-git-server/k8s-cmp.git && cd k8s-cmp

# 2. 환경변수 설정
cp .env.example .env
# .env 파일에 AUTH_JWT_SECRET, ENCRYPTION_KEY, KUBECONFIG 입력

# 3. 전체 스택 실행 (Docker만으로)
docker compose -f deploy/docker-compose.full.yaml up -d --build

# 4. 관리자 계정 생성
go run ./cmd/tools/genhash/main.go   # 비밀번호 해시 생성
docker exec -it k8s-cmp-postgres psql -U cmp -d k8s_cmp
# → INSERT INTO users ... 실행

# 5. 접속
# 프론트엔드: http://localhost:3000
# API:        http://localhost:8080
```
