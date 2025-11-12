# Radio Container

This repository bundles the components that make up the Radio Controller stack.

## Projects

### rcc
- Go implementation of the Radio Control Container API and orchestration services.
- Contains the production service, tests, and supporting tooling.

### web-ui
- Lightweight Go/HTML/JS web interface for interactive control and telemetry views.
- Provides static assets and helper scripts for running UI-driven workflows.

### silvus-mock
- Silvus-compatible mock radio used for integration, contract, and performance testing.
- Includes Docker assets and scripts for running the mock radio locally or in CI.

## Documentation
Shared documentation lives in `docs/`. The `rcc` package refers to these files directly; if you move documentation make sure to update those references.

## Getting Started
Clone the repo and work within the subproject you need:

```bash
# Example: run Go unit tests for RCC
cd rcc
make test
```

Refer to each subproject's own README for detailed setup and workflow guidance.
