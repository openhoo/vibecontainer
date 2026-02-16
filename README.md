# Vibecontainer

`vibecontainer` is a provider-aware runtime for AI CLI workflows.

This single repository now publishes:
- base image
- Claude flavor image
- Codex flavor image

## Quick Start

Run prebuilt images from GHCR:

```sh
# Base
docker run -d --name vibecontainer \
  --cap-add NET_ADMIN --cap-add NET_RAW \
  -e TMUX_WEB_ENABLE=1 \
  -p 127.0.0.1:7681:7681 \
  ghcr.io/openhoo/vibecontainer:latest
```

```sh
# Claude flavor
docker run -d --name vibecontainer-claude \
  --cap-add NET_ADMIN --cap-add NET_RAW \
  -e CLAUDE_CODE_OAUTH_TOKEN=<your-token> \
  -e TMUX_WEB_ENABLE=1 \
  -p 127.0.0.1:7681:7681 \
  ghcr.io/openhoo/vibecontainer:claude
```

```sh
# Codex flavor
docker run -d --name vibecontainer-codex \
  --cap-add NET_ADMIN --cap-add NET_RAW \
  -e TMUX_WEB_ENABLE=1 \
  -p 127.0.0.1:7681:7681 \
  ghcr.io/openhoo/vibecontainer:codex
```

Open read-only stream: [http://127.0.0.1:7681](http://127.0.0.1:7681)

## Build from Source

```sh
git clone https://github.com/openhoo/vibecontainer.git
cd vibecontainer
```

```sh
# Base target
docker build --target base -t vibecontainer:base-local .

# Claude target
docker build --target claude -t vibecontainer:claude-local .

# Codex target
docker build --target codex -t vibecontainer:codex-local .
```

## Runtime Behavior

`entrypoint.sh` continues to own tmux, ttyd, and firewall lifecycle for all images.

Flavor images use `claude-entrypoint.sh` / `codex-entrypoint.sh`, both backed by
shared logic in `provider-entrypoint-base.sh`:
- if command args are passed, args are used directly
- if no args are passed, provider CLI auto-starts

## Environment Variables

### Shared (`base`, `claude`, `codex`)

| Variable | Default | Description |
|----------|---------|-------------|
| `TMUX_SESSION_NAME` | `vibe` | Name of the tmux session |
| `TMUX_WEB_ENABLE` | `0` | Enable ttyd browser streaming |
| `TMUX_WEB_BIND_ADDRESS` | `0.0.0.0` | Bind address for ttyd inside container |
| `TMUX_WEB_READONLY_PORT` | `7681` | Read-only stream port |
| `TMUX_WEB_INTERACTIVE_ENABLE` | `0` | Enable interactive stream |
| `TMUX_WEB_INTERACTIVE_PORT` | `7682` | Interactive stream port |
| `TTYD_CREDENTIAL` | *(none)* | Basic auth for ttyd (`user:password`) |
| `FIREWALL_ENABLE` | `1` | `1` enables strict ufw policy, `0` skips firewall setup |

All `*_FILE` variables are strict. If set, they must point to a readable file or startup exits with an error.

### Claude flavor only

| Variable | Description |
|----------|-------------|
| `CLAUDE_CODE_OAUTH_TOKEN` | Claude Code OAuth token |
| `CLAUDE_CODE_OAUTH_TOKEN_FILE` | Path to mounted file containing OAuth token |
| `ANTHROPIC_API_KEY` | Anthropic API key for Claude API auth |
| `ANTHROPIC_API_KEY_FILE` | Path to mounted file containing Anthropic API key |

Claude auth is only required when the startup command is `claude` (default or explicit command), and accepts either OAuth or Anthropic API key auth.

### Codex flavor only

| Variable | Description |
|----------|-------------|
| `CODEX_AUTH_JSON` | Full Codex auth JSON payload to write to `/home/dev/.codex/auth.json` |
| `CODEX_AUTH_JSON_FILE` | Path to mounted file containing full Codex auth JSON payload |
| `OPENAI_API_KEY` | OpenAI API key (supported by Codex) |
| `OPENAI_API_KEY_FILE` | Path to mounted file containing OpenAI API key |
| `CODEX_API_KEY` | Codex-specific API key env alias supported by Codex |
| `CODEX_API_KEY_FILE` | Path to mounted file containing Codex API key |

Codex auth precedence when launching `codex`:
1. `CODEX_AUTH_JSON` / `CODEX_AUTH_JSON_FILE`
2. API keys (`OPENAI_API_KEY*`, `CODEX_API_KEY*`)

`CODEX_AUTH_JSON*` must be valid JSON object with one supported auth shape:
- API key shape: `OPENAI_API_KEY` is a non-empty string
- Token shape: `tokens.id_token`, `tokens.access_token`, and `tokens.refresh_token` are strings
- Optional `auth_mode` (if set) must be one of: `apikey`, `chatgpt`, `chatgptAuthTokens`

Malformed explicit OAuth input (invalid file path, invalid JSON, or invalid shape) fails startup with a clear error.

Examples:

```sh
# OAuth via full auth.json payload (recommended)
docker run --rm -it \
  -e CODEX_AUTH_JSON_FILE=/run/secrets/codex_auth.json \
  -v "$(pwd)/codex_auth.json:/run/secrets/codex_auth.json:ro" \
  ghcr.io/openhoo/vibecontainer:codex
```

```json
{
  "auth_mode": "chatgptAuthTokens",
  "tokens": {
    "id_token": "<jwt>",
    "access_token": "<jwt>",
    "refresh_token": ""
  }
}
```

```sh
# API key mode
docker run --rm -it \
  -e OPENAI_API_KEY=<your-api-key> \
  ghcr.io/openhoo/vibecontainer:codex
```

## Tags

Published from one package: `ghcr.io/openhoo/vibecontainer`

- base image:
  - `latest`
  - `vX.Y.Z`
  - `sha-<commit>`
- Claude flavor:
  - `claude`
  - `claude-vX.Y.Z`
  - `claude-code-<version>`
- Codex flavor:
  - `codex`
  - `codex-vX.Y.Z`
  - `codex-cli-<version>`

## Security

When `FIREWALL_ENABLE=1` (default), startup enables a strict `ufw` profile.
This requires Docker capabilities:
- `NET_ADMIN`
- `NET_RAW`

For local development, you can bypass firewall setup with `FIREWALL_ENABLE=0`.

## CI Quality Gates

Pull requests and pushes to `main` run:
- `actionlint` for workflow validation
- `shellcheck` for shell scripts
- `hadolint` for Dockerfile

## License

MIT
