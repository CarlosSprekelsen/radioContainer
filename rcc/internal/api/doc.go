// Package api implements the HTTP API gateway for the Radio Control Container.
//
// The API gateway exposes northbound HTTP/JSON commands and SSE endpoints,
// translating HTTP requests into orchestrator calls with client throttling.
//
// Architecture References:
//   - OpenAPI §2: HTTP/JSON API specification
//   - Telemetry SSE §1: Server-Sent Events protocol
package api
