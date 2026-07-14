# Templates

`http` and `log` actions render strings with Go `text/template`.

## Context

| Field | Meaning |
|-------|---------|
| `.event` | Event name from the trigger (e.g. `email.bounced`) |
| `.created_at` | Timestamp from the payload when present |
| `.data` | Event data object (map) |
| `.raw` | Raw JSON body as a string |

`${ENV}` in workflow YAML is expanded at config load time (before templates).

## Helpers

| Helper | Example | Notes |
|--------|---------|-------|
| `default` | `{{default .data.from "N/A"}}` | Empty / missing → fallback |
| `join` | `{{join .data.to ", "}}` | Arrays → string; empty → `N/A` |
| `json` | `{{json .data.tags}}` | Marshal value; nil → `None` |
| `eventTitle` | `{{eventTitle .event}}` | Resend-oriented Discord titles |
| `eventColor` | `{{eventColor .event}}` | Discord embed color ints |

### `eventTitle` / `eventColor` map

| Event | Title | Color |
|-------|-------|-------|
| `email.failed` | Email Failed | `15158332` |
| `email.bounced` | Email Bounced | `16753920` |
| `email.clicked` | Email Clicked | `3447003` |
| `email.opened` | Email Opened | `5763719` |
| `email.received` | Email Received | `3066993` |
| other | Resend Event | `9807270` |
