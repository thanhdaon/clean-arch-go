# HTTP Status Codes Design

## Overview

Design for improving HTTP error responses across the Task Manager API, aligning domain error kinds with proper HTTP status codes and documenting them in OpenAPI.

## Goals

1. Document all possible error codes in OpenAPI for API consumers
2. Ensure each domain error kind maps to the correct HTTP status code

## Error Kind → HTTP Status Mapping

| errkind | HTTP | Description |
|---------|------|-------------|
| `Authorization` | 401 | Invalid/missing credentials |
| `Permission` | 403 | Authenticated but not allowed |
| `NotExist` | 404 | Resource not found |
| `Exist` | 409 | Duplicate resource |
| `Connection` | 503 | Database/external service down |
| `Internal` / `Other` | 500 | Unexpected server error |

**Validation errors** (invalid email, missing required fields, etc.) → 400 Bad Request
- These are handled before domain logic, no errkind needed

## OpenAPI Response Documentation

Replace generic `default` error with specific response codes.

### Common Responses (defined in `components/responses`)

```yaml
components:
  responses:
    BadRequest:
      description: Invalid request body or parameters
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    Unauthorized:
      description: Missing or invalid authentication
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    Forbidden:
      description: Authenticated but lacks permission
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    Conflict:
      description: Resource already exists
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    InternalError:
      description: Unexpected server error
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    ServiceUnavailable:
      description: Service temporarily unavailable
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
```

### Endpoint Response Mapping

| Endpoint | 400 | 401 | 403 | 404 | 409 | 500 | 503 |
|----------|-----|-----|-----|-----|-----|-----|-----|
| `GET /users` | - | ✓ | - | - | - | ✓ | ✓ |
| `POST /users` | ✓ | - | - | - | ✓ | ✓ | ✓ |
| `GET /users/{userId}` | - | ✓ | - | ✓ | - | ✓ | ✓ |
| `PATCH /users/{userId}` | ✓ | ✓ | - | ✓ | - | ✓ | ✓ |
| `DELETE /users/{userId}` | - | ✓ | - | ✓ | - | ✓ | ✓ |
| `PUT /users/{userId}/role` | ✓ | ✓ | ✓ | ✓ | - | ✓ | ✓ |
| `POST /auth/login` | ✓ | - | - | - | - | ✓ | ✓ |
| `GET /tasks` | - | ✓ | - | - | - | ✓ | ✓ |
| `POST /tasks` | ✓ | ✓ | - | - | - | ✓ | ✓ |
| `PUT /tasks/{taskId}/status` | ✓ | ✓ | - | ✓ | - | ✓ | ✓ |
| `PATCH /tasks/{taskId}` | ✓ | ✓ | - | ✓ | - | ✓ | ✓ |
| `DELETE /tasks/{taskId}` | - | ✓ | - | ✓ | - | ✓ | ✓ |
| `PUT /tasks/{taskId}/assign/{assigneeId}` | - | ✓ | - | ✓ | - | ✓ | ✓ |
| `DELETE /tasks/{taskId}/assign` | - | ✓ | - | ✓ | - | ✓ | ✓ |
| `PUT /tasks/{taskId}/reopen` | - | ✓ | - | ✓ | - | ✓ | ✓ |
| `PUT /tasks/{taskId}/archive` | - | ✓ | - | ✓ | - | ✓ | ✓ |
| `PUT /tasks/{taskId}/priority` | ✓ | ✓ | - | ✓ | - | ✓ | ✓ |
| `PUT /tasks/{taskId}/due-date` | ✓ | ✓ | - | ✓ | - | ✓ | ✓ |
| `PATCH /tasks/{taskId}/description` | ✓ | ✓ | - | ✓ | - | ✓ | ✓ |
| `POST /tasks/{taskId}/tags` | ✓ | ✓ | - | ✓ | - | ✓ | ✓ |
| `DELETE /tasks/{taskId}/tags/{tagId}` | - | ✓ | - | ✓ | - | ✓ | ✓ |

## Code Implementation

### Centralized Error Mapper

Add to `ports/openapi_response.go`:

```go
func mapErrorToStatus(err error) int {
    switch {
    case errors.Is(errkind.Authorization, err):
        return http.StatusUnauthorized
    case errors.Is(errkind.Permission, err):
        return http.StatusForbidden
    case errors.Is(errkind.NotExist, err):
        return http.StatusNotFound
    case errors.Is(errkind.Exist, err):
        return http.StatusConflict
    case errors.Is(errkind.Connection, err):
        return http.StatusServiceUnavailable
    default:
        return http.StatusInternalServerError
    }
}
```

### Simplified Handlers

Remove manual helper selection. Handlers call `responseError()` which uses the mapper:

```go
// Before (manual selection):
if errors.Is(errkind.NotExist, err) {
    notFound(ctx, err, w, r)
} else {
    badRequest(ctx, err, w, r)
}

// After (automatic mapping):
responseError(ctx, err, w, r)
```

### Validation Errors

Keep explicit `badRequest()` calls for validation errors that occur before domain logic:

```go
body := PostUser{}
if err := render.Decode(r, &body); err != nil {
    badRequest(r.Context(), err, w, r)  // Invalid JSON
    return
}
```

## Changes Summary

1. **Add** `mapErrorToStatus()` in `ports/openapi_response.go`
2. **Update** `responseError()` to use the mapper
3. **Add** common response definitions in `ports/openapi.yml`
4. **Replace** `default` responses with specific codes per endpoint
5. **Refactor** handlers to use `responseError()` instead of manual helpers
6. **Remove** redundant helper functions (`unauthorised`, `internalError`, `badRequest`)
