# Radio Control Container (RCC)

Radio Control Container implementation following IEEE 42010 + arc42 architecture.

## Architecture

See `docs/radio_control_container_ieee_42010_arc_42_architecture_draft_v1.md` for complete architecture documentation.

## Components

- **API Gateway** (`internal/api/`) - Northbound HTTP/JSON commands and SSE endpoint
- **Auth** (`internal/auth/`) - Token validation and scope enforcement  
- **Radio Manager** (`internal/radio/`) - Radio inventory and adapter management
- **Command Orchestrator** (`internal/command/`) - Request validation and routing
- **Telemetry Hub** (`internal/telemetry/`) - SSE event multiplexing
- **Radio Adapters** (`internal/adapter/`) - Vendor-specific radio interfaces
- **Audit Logger** (`internal/audit/`) - Action logging
- **Config Store** (`internal/config/`) - Channel maps and configuration

## Building

```bash
go build ./...
```

## Running

```bash
go run cmd/rcc/main.go
```
