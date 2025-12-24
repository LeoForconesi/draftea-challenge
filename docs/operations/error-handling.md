# Error Handling Strategy
↩️ [Return to README](../../README.md)

## Error Response Shape
```json
{
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "...",
    "details": {}
  }
}
```

## HTTP Mapping
- VALIDATION_ERROR -> 400
- UNAUTHORIZED -> 401
- NOT_FOUND -> 404
- INSUFFICIENT_FUNDS -> 409
- GATEWAY_TIMEOUT -> 504
- GATEWAY_ERROR -> 502
- INTERNAL -> 500

## Validation
- Explicit field checks for required fields, UUIDs, and amounts.
- Validation errors return details to help clients correct requests.

## Logging
- Structured logs with request_id and business identifiers.
- Levels: debug/info/warn/error per severity.
