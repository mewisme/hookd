# hookd

Minimal pluggable webhook runner (Go + YAML). **Triggers** verify inbound webhooks; **actions** run outbound steps. Enable or disable providers by editing the workflow file.

**Images** (Go 1.26): [`mewisme/hookd`](https://hub.docker.com/r/mewisme/hookd) · [`ghcr.io/mewisme/hookd`](https://github.com/mewisme/hookd/pkgs/container/hookd)

```bash
docker pull mewisme/hookd:latest
# or
docker pull ghcr.io/mewisme/hookd:latest
```

## Quick start (Resend → Discord)

1. Copy env and fill secrets:

```bash
cp .env.example .env
# RESEND_WEBHOOK_SECRET=whsec_...
# DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
```

2. Run (pulls `mewisme/hookd:latest`, or build locally with `--build`):

```bash
docker compose up -d
```

3. In [Resend → Webhooks](https://resend.com/webhooks), set endpoint:

```
https://<your-host>/hooks/resend-to-discord
```

Subscribe to the email events you care about (`email.bounced`, `email.failed`, …).

Health check: `GET /healthz` → `ok`.

## Workflows

Each entry under `triggers:` is one inbound endpoint. Delete a block to disable it (restart compose). YAML-only changes need no rebuild.

| Field | Meaning |
|-------|---------|
| `type` | Plugin: `resend` or `generic` |
| `path` | HTTP route (e.g. `/hooks/resend-to-discord`) |
| `secret` | Signing / shared secret (`${ENV}` ok) |
| `hooks[].event` | Exact event name, or `*` for all |
| `hooks[].actions` | List of action plugins (`http`, `log`) |

See [config/workflows.example.yaml](config/workflows.example.yaml) for a multi-trigger sample. Live Resend→Discord config: [config/workflows.yaml](config/workflows.yaml).

## Built-in plugins

**Triggers** — [docs/triggers/](docs/triggers/)

- [`resend`](docs/triggers/resend.md) — Svix-signed Resend webhooks
- [`generic`](docs/triggers/generic.md) — shared-secret header

**Actions** — [docs/actions/](docs/actions/)

- [`http`](docs/actions/http.md) — outbound HTTP request
- [`log`](docs/actions/log.md) — structured log line

Templates: [docs/templates.md](docs/templates.md)

## Local run (no Docker)

```bash
export RESEND_WEBHOOK_SECRET=whsec_...
export DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
export CONFIG_PATH=./config/workflows.yaml
go run ./cmd/server
```

## New plugin type (rebuild once)

1. Implement `plugin.Trigger` or `plugin.Action` under `internal/plugin/...`
2. `plugin.RegisterTrigger` / `RegisterAction` in `init()`
3. Blank-import the package from `cmd/server/main.go`
4. Reference the new `type:` in YAML

## Env

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | Listen port |
| `CONFIG_PATH` | `/config/workflows.yaml` | Workflows file |
| `RESEND_WEBHOOK_SECRET` | — | Svix secret |
| `DISCORD_WEBHOOK_URL` | — | Discord webhook URL |

## Publish images

Push a tag `v*` (or run the **Release** workflow manually). Multi-arch manifest (`linux/amd64`, `linux/arm64`) via Buildx cross-compile — same tag works on Intel/AMD and Apple Silicon / ARM servers:

- `mewisme/hookd:<version>` + `:latest`
- `ghcr.io/mewisme/hookd:<version>` + `:latest`

Repo secrets for Docker Hub: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`. GHCR uses `GITHUB_TOKEN`.
