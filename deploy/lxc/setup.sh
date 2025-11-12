#!/bin/bash
# LXC Setup Script for Radio Control Container (RCC)
# Provisions an ARM32-compatible container image using LXC/LXE primitives.

set -euo pipefail

CONTAINER_NAME="rcc"
PROFILE_FILE="$(dirname "$0")/rcc.profile"
BINARY_SOURCE="./rcc"
APP_DIR="/opt/rcc"
CONFIG_DIR="/etc/rcc"
SERVICE_NAME="rcc"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
  echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
  echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
  echo -e "${RED}[ERROR]${NC} $1" >&2
  exit 1
}

require_root() {
  if [[ $EUID -ne 0 ]]; then
    error "This script must be run as root (required by LXC tooling)"
  fi
}

require_lxc() {
  if ! command -v lxc >/dev/null 2>&1; then
    error "LXC is not installed. Install with: sudo apt install lxc lxc-templates"
  fi
}

ensure_binary() {
  if [[ ! -f "$BINARY_SOURCE" ]]; then
    error "Binary $BINARY_SOURCE not found. Build it first (e.g. 'GOOS=linux GOARCH=arm go build -o rcc ./cmd/rcc')."
  fi
}

cleanup_existing() {
  if lxc list | grep -q "^${CONTAINER_NAME}\\b"; then
    warn "Container ${CONTAINER_NAME} already exists"
    read -p "Recreate it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
      log "Stopping and removing existing container..."
      lxc stop "$CONTAINER_NAME" 2>/dev/null || true
      lxc delete "$CONTAINER_NAME" 2>/dev/null || true
    else
      log "Exiting without changes"
      exit 0
    fi
  fi
}

create_container() {
  if [[ ! -f "$PROFILE_FILE" ]]; then
    error "Profile file not found at $PROFILE_FILE"
  fi

  log "Creating LXC container ${CONTAINER_NAME}"
  lxc init ubuntu:22.04 "$CONTAINER_NAME" -p "$PROFILE_FILE"
  log "Starting container"
  lxc start "$CONTAINER_NAME"
  log "Waiting for container to initialize"
  sleep 10
}

provision_runtime() {
  log "Installing runtime dependencies"
  lxc exec "$CONTAINER_NAME" -- apt update
  lxc exec "$CONTAINER_NAME" -- apt install -y ca-certificates tzdata

  log "Creating runtime directories"
  lxc exec "$CONTAINER_NAME" -- mkdir -p "$APP_DIR" "$APP_DIR/logs" "$CONFIG_DIR"
  lxc exec "$CONTAINER_NAME" -- mkdir -p /var/log/rcc /var/lib/rcc /run/secrets || true
}

push_binary_and_config() {
  log "Pushing RCC binary"
  lxc file push "$BINARY_SOURCE" "$CONTAINER_NAME$APP_DIR/rcc"
  lxc exec "$CONTAINER_NAME" -- chmod +x "$APP_DIR/rcc"

  log "Writing default configuration"
  lxc exec "$CONTAINER_NAME" -- tee "$CONFIG_DIR/config.json" >/dev/null <<'JSON'
{
  "heartbeatInterval": "30s",
  "heartbeatTimeout": "90s",
  "telemetry": {
    "bufferSize": 256
  }
}
JSON
}

configure_service() {
  log "Configuring systemd service"
  lxc exec "$CONTAINER_NAME" -- tee /etc/systemd/system/${SERVICE_NAME}.service >/dev/null <<SERVICE
[Unit]
Description=Radio Control Container Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/rcc
Restart=on-failure
RestartSec=5
Environment=RCC_ADDR=0.0.0.0:8080

[Install]
WantedBy=multi-user.target
SERVICE

  log "Enabling RCC service"
  lxc exec "$CONTAINER_NAME" -- systemctl daemon-reload
  lxc exec "$CONTAINER_NAME" -- systemctl enable ${SERVICE_NAME}.service
  lxc exec "$CONTAINER_NAME" -- systemctl start ${SERVICE_NAME}.service
}

verify_service() {
  log "Verifying service status"
  if lxc exec "$CONTAINER_NAME" -- systemctl is-active --quiet ${SERVICE_NAME}.service; then
    log "${SERVICE_NAME} service is running"
  else
    warn "${SERVICE_NAME} service is not active"
    lxc exec "$CONTAINER_NAME" -- systemctl status ${SERVICE_NAME}.service || true
  fi

  log "Container summary"
  lxc list "$CONTAINER_NAME"
}

require_root
require_lxc
ensure_binary
cleanup_existing
create_container
provision_runtime
push_binary_and_config
configure_service
verify_service

log "LXC setup complete"
