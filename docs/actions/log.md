# Action: `log`

Writes a structured log line (`slog.Info`). Useful for debugging workflows.

## Config

| Field | Required | Description |
|-------|----------|-------------|
| `type` | yes | Must be `log` |
| `message` | no | Log message template (default `{{.event}}`) |

```yaml
actions:
  - type: log
    message: "got {{.event}} for {{default .data.from "unknown"}}"
```

Also logs attribute `trigger_event` with the event name.

## Templates

See [templates](../templates.md) for context fields and helpers.
