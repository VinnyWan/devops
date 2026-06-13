#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# DevOps Platform - Docker 一键启动脚本
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# 检查 Docker 环境
check_deps() {
  if ! command -v docker &>/dev/null; then
    log_error "Docker 未安装，请先安装 Docker (https://docs.docker.com/get-docker/)"
    exit 1
  fi
  if ! docker compose version &>/dev/null; then
    log_error "Docker Compose 未安装或版本过旧，请安装 Docker Compose v2+"
    exit 1
  fi
  log_info "Docker 环境检查通过"
}

# 生成 .env 文件
ensure_env() {
  if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
      cp .env.example .env
      log_info "已从 .env.example 创建 .env 配置文件"
      log_warn "请按需修改 .env 中的配置项（如数据库密码等）"
    else
      log_warn ".env.example 不存在，使用 docker-compose.yml 中的默认值"
    fi
  else
    log_info "已存在 .env 配置文件"
  fi
}

# 启动服务
start_services() {
  log_info "正在构建并启动服务..."
  docker compose up -d --build

  log_info "等待服务就绪..."
  local max_wait=120
  local waited=0
  while [ $waited -lt $max_wait ]; do
    if curl -sf http://localhost:8000/healthz &>/dev/null; then
      log_info "后端服务已就绪"
      break
    fi
    sleep 5
    waited=$((waited + 5))
  done

  if [ $waited -ge $max_wait ]; then
    log_error "后端服务启动超时（${max_wait}s），请检查日志: docker compose logs backend"
    exit 1
  fi

  log_info "============================================"
  log_info "  DevOps 运维平台启动成功！"
  log_info "  Web UI:  http://localhost:${FRONTEND_PORT:-8080}"
  log_info "  API:     http://localhost:8000/api/v1"
  log_info "  Swagger: http://localhost:8000/swagger/index.html"
  log_info "============================================"
}

# 显示状态
show_status() {
  echo ""
  log_info "当前服务状态："
  docker compose ps
}

# 主流程
check_deps
ensure_env
start_services
show_status
