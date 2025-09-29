### Lazy Lagoon API

Base path: `/lazy-lagoon`

Purpose: Transform and truncate CSV/JSON/JSONL/SQL with rules; optional webhook callbacks. Returns preview content and extracted attribute paths.

Endpoints
- POST `/truncate`: Truncate input (CSV/JSONL/SQL) to a preview; stores in output; returns preview content string.
- POST `/transform`: Apply rules to CSV/JSON/JSONL and return preview plus `attributes.paths`. If webhook provided, posts status payload.
- GET `/healthz/ready`: Readiness.

Requests
- `RequestBodyTruncate`:
  - `input`: `Input { storageType, dataType, reference, credential }`
  - `output`: `Output { storageType, dataType, reference, credential }`
- `RequestBodyTransform`:
  - `input`
  - `output?`
  - `rules: Rule[]`
  - `webhook?`

Rules
- `Rule { expression?, actions[] }`
- `Action { actionType, fieldName }`
- `Expression { logicalOperator, expressions[] }`

OpenAPI spec: `docs/lazy-lagoon/openapi.yaml`


