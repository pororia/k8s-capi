# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/migrator ./cmd/migrator

# Final stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api-server .
COPY --from=builder /app/worker .
COPY --from=builder /app/migrator .
COPY --from=builder /app/migrations ./migrations
# configs/ 는 선택적 (없으면 환경변수로 대체됨)
COPY --from=builder /app/configs/config.example.yaml ./configs/config.example.yaml
# manifest_templates 는 //go:embed 로 바이너리에 포함되어 있으므로 불필요
RUN mkdir -p /etc/cmp

EXPOSE 8080

CMD ["./api-server"]
