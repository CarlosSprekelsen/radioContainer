# Route → Scope Matrix

**Source**: OpenAPI v1 §1.2  
**Quote**: "viewer: read-only (list radios, get state, subscribe to telemetry)"  
**Quote**: "controller: all viewer privileges plus control actions (select radio, set power, set channel)"

## Overview

This document defines the authorization requirements for each API endpoint in the Radio Control Container. All endpoints except `/health` require authentication via Bearer token.

## Route Authorization Matrix

| Endpoint | Method | Required Scope | Required Role | Description |
|----------|--------|----------------|---------------|-------------|
| `/api/v1/health` | GET | None | None | Health check (no auth required) |
| `/api/v1/capabilities` | GET | `read` | `viewer` | Get API capabilities |
| `/api/v1/radios` | GET | `read` | `viewer` | List all radios |
| `/api/v1/radios/select` | POST | `control` | `controller` | Select active radio |
| `/api/v1/radios/{id}` | GET | `read` | `viewer` | Get specific radio details |
| `/api/v1/radios/{id}/power` | GET | `read` | `viewer` | Get radio power setting |
| `/api/v1/radios/{id}/power` | POST | `control` | `controller` | Set radio power |
| `/api/v1/radios/{id}/channel` | GET | `read` | `viewer` | Get radio channel |
| `/api/v1/radios/{id}/channel` | POST | `control` | `controller` | Set radio channel |
| `/api/v1/telemetry` | GET | `telemetry` | `viewer` | Subscribe to telemetry stream |

## Scope Definitions

### `read` Scope
- **Purpose**: Read-only access to radio information and status
- **Allowed Operations**: GET requests to read radio state, capabilities, and status
- **Required For**: `viewer` role and above

### `control` Scope  
- **Purpose**: Control operations on radios
- **Allowed Operations**: POST requests to change radio settings (power, channel, selection)
- **Required For**: `controller` role only

### `telemetry` Scope
- **Purpose**: Access to real-time telemetry streams
- **Allowed Operations**: GET requests to subscribe to Server-Sent Events
- **Required For**: `viewer` role and above

## Role Hierarchy

### `viewer` Role
- **Scopes**: `read`, `telemetry`
- **Access**: Read-only access to radio information and telemetry
- **Endpoints**: All GET endpoints except control operations

### `controller` Role  
- **Scopes**: `read`, `control`, `telemetry`
- **Access**: Full access to all operations
- **Endpoints**: All endpoints including control operations

## Implementation Details

### Authentication Flow
1. **Token Extraction**: Bearer token from `Authorization` header
2. **Token Verification**: JWT signature verification (RS256/HS256)
3. **Claims Extraction**: Subject, roles, and scopes from token
4. **Scope Validation**: Check required scope against token scopes
5. **Role Validation**: Verify role has required permissions

### Error Responses

#### 401 Unauthorized
- Missing `Authorization` header
- Invalid token format
- Token verification failed
- Expired token

#### 403 Forbidden  
- Valid token but insufficient scopes
- Valid token but insufficient role
- Token missing required claims

### Example Token Claims

#### Viewer Token
```json
{
  "sub": "user-123",
  "roles": ["viewer"],
  "scopes": ["read", "telemetry"],
  "iat": 1640995200,
  "exp": 1641081600
}
```

#### Controller Token
```json
{
  "sub": "admin-456", 
  "roles": ["controller"],
  "scopes": ["read", "control", "telemetry"],
  "iat": 1640995200,
  "exp": 1641081600
}
```

## Security Considerations

### Default Deny
- All endpoints require explicit authentication except `/health`
- No anonymous access to any protected resources
- Missing or invalid tokens result in 401 Unauthorized

### Scope Isolation
- Scopes are strictly enforced per endpoint
- No privilege escalation through endpoint manipulation
- Each operation requires the minimum necessary scope

### Token Validation
- JWT signature verification with RS256 (production) or HS256 (tests)
- Token expiration validation
- Required claims validation (sub, roles, scopes)
- Algorithm restriction to prevent confusion attacks

## Testing

### Test Scenarios
1. **No Token**: All protected endpoints return 401
2. **Invalid Token**: All protected endpoints return 401  
3. **Viewer Token on Control**: Control endpoints return 403
4. **Controller Token**: All endpoints return 200
5. **Health Endpoint**: Always returns 200 (no auth required)

### Test Tokens
- `viewer-token`: Mock token with viewer role and read/telemetry scopes
- `controller-token`: Mock token with controller role and all scopes
- `invalid-token`: Mock token that fails verification

## Compliance

This matrix ensures compliance with:
- **OpenAPI v1 §1.1**: Bearer token authentication
- **OpenAPI v1 §1.2**: Role-based access control
- **Architecture §5**: Token validation and RBAC enforcement
- **Security Best Practices**: Default deny, least privilege, scope isolation
