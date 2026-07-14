# Trigger: `generic`

Shared-secret inbound webhook for any provider that can send a secret header.

## Config

| Field | Required | Description |
|-------|----------|-------------|
| `type` | yes | Must be `generic` |
| `path` | yes | HTTP route, e.g. `/hooks/partner` |
| `secret` | yes | Expected secret value, supports `${ENV}` |
| `header` | no | Header name to check (default: `X-Webhook-Secret`) |
| `fail_on_error` | no | If `true`, return 500 when an action fails |
| `hooks` | yes | Event → actions |

```yaml
triggers:
  - id: partner
    type: generic
    path: /hooks/partner
    secret: ${PARTNER_SECRET}
    header: X-Webhook-Secret
    hooks:
      - event: order.created
        actions:
          - type: http
            method: POST
            url: https://example.com/orders
            body: "{{.raw}}"
```

## Verification

Compares `header` (default `X-Webhook-Secret`) to `secret` with constant-time compare. Mismatch → `400`.

Dedup key (optional): `Idempotency-Key` header.

## Payload

Expects JSON with an event name in `type` or `event`:

```json
{
  "type": "order.created",
  "created_at": "2024-01-01T00:00:00Z",
  "data": { "id": 1 }
}
```

- If `data` is missing, the whole object is used as `.data`
- Template context: `.event`, `.created_at`, `.data`, `.raw`

## Example request

```bash
curl -X POST https://<host>/hooks/partner \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: $PARTNER_SECRET" \
  -d '{"type":"order.created","data":{"id":1}}'
```
