# ---- 构建阶段 ----
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o devops-server cmd/server/main.go

# ---- 运行阶段 ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/devops-server .
COPY --from=builder /app/config/config.yaml.example ./config/config.yaml

RUN mkdir -p logs

EXPOSE 8000

ENTRYPOINT ["./devops-server"]
