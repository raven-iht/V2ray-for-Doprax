# Echo Decision Backend (GoRules-style)

Run:

```bash
cd /workspace
GO111MODULE=on go run ./cmd/server
```

Test:

```bash
curl -s -X POST http://localhost:8080/api/decisions/pricing/approveOrder/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"basketTotal": 249.9, "country": "IR"}' | jq .
```

Health:

```bash
curl -s http://localhost:8080/healthz
```

Endpoint shape mirrors GoRules:

- POST `/api/decisions/:model/:decision/evaluate`

Sample model/decision is registered in `internal/http/wire.go` as `pricing/approveOrder`.
