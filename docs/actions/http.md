# Action: `http`

Sends an outbound HTTP request. URL, headers, and body support Go templates.

## Config

| Field | Required | Description |
|-------|----------|-------------|
| `type` | yes | Must be `http` |
| `url` | yes | Request URL (templated, `${ENV}` ok) |
| `method` | no | HTTP method (default `POST`) |
| `headers` | no | Map of header name → value (templated) |
| `body` | no | Request body string (templated) |

Timeout: 15s. Non-2xx responses are errors. If `body` is set and `Content-Type` is missing, defaults to `application/json`.

```yaml
actions:
  - type: http
    method: POST
    url: ${DISCORD_WEBHOOK_URL}
    headers:
      Content-Type: application/json
      Authorization: "Bearer ${ALERT_TOKEN}"
    body: |
      {"content": "{{.event}} to {{join .data.to ", "}}"}
```

## Templates

See [templates](../templates.md) for context fields and helpers.
