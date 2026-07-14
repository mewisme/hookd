# Trigger: `resend`

Verifies [Resend](https://resend.com/docs/webhooks) webhooks using Svix signatures.

## Config

| Field | Required | Description |
|-------|----------|-------------|
| `type` | yes | Must be `resend` |
| `path` | yes | HTTP route, e.g. `/hooks/resend-to-discord` |
| `secret` | yes | Signing secret from Resend (`whsec_...`), supports `${ENV}` |
| `fail_on_error` | no | If `true`, return 500 when an action fails (default: still `200`) |
| `hooks` | yes | Event → actions |

```yaml
triggers:
  - id: resend-to-discord
    type: resend
    path: /hooks/resend-to-discord
    secret: ${RESEND_WEBHOOK_SECRET}
    hooks:
      - event: "*"
        actions:
          - type: http
            method: POST
            url: ${DISCORD_WEBHOOK_URL}
            body: |
              {"content": "{{.event}}"}
```

## Verification

Expects headers:

- `svix-id`
- `svix-timestamp`
- `svix-signature`

Uses the **raw request body** for signature checks. Invalid signatures → `400`.

Dedup key: `svix-id` (in-memory TTL).

## Payload

Parses Resend JSON:

```json
{
  "type": "email.bounced",
  "created_at": "2024-01-01T00:00:00.000Z",
  "data": { ... }
}
```

- Event name = top-level `type`
- Template context: `.event`, `.created_at`, `.data`, `.raw`

Common events: `email.sent`, `email.delivered`, `email.opened`, `email.clicked`, `email.bounced`, `email.failed`, `email.received`, …

## Setup

1. Set `RESEND_WEBHOOK_SECRET` (from Resend → Webhooks → signing secret).
2. Point Resend at `https://<host><path>` (e.g. `/hooks/resend-to-discord`).
3. Subscribe to the events you need.

See [config/workflows.yaml](../config/workflows.yaml) for a Resend → Discord example.
