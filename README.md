# Vibecontainer

[![npm](https://img.shields.io/npm/v/@openhoo/vibecontainer)](https://www.npmjs.com/package/@openhoo/vibecontainer)

`vibecontainer` is a provider-aware runtime for AI CLI workflows.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/openhoo/vibecontainer/main/install.sh | sh
```

### npm

```sh
npm install -g @openhoo/vibecontainer
```

### Binary download

Download a prebuilt binary from [GitHub Releases](https://github.com/openhoo/vibecontainer/releases/latest).

| Platform | amd64 | arm64 |
|----------|-------|-------|
| Linux | `vibecontainer-linux-amd64` | `vibecontainer-linux-arm64` |
| macOS | `vibecontainer-darwin-amd64` | `vibecontainer-darwin-arm64` |
| Windows | `vibecontainer-windows-amd64.exe` | `vibecontainer-windows-arm64.exe` |

### Nix

```sh
nix run github:openhoo/vibecontainer
```

### Docker

Prebuilt images are published to GHCR with base, Claude, and Codex flavors:

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

## Cloudflare Tunnel

You can publish the read-only (or interactive) ttyd endpoint through Cloudflare Tunnel using `cloudflared`.

```sh
# Example: expose read-only ttyd through an existing named tunnel
docker run -d --name cloudflared \
  --network container:vibecontainer \
  cloudflare/cloudflared:latest \
  tunnel --no-autoupdate run --token <TUNNEL_TOKEN>
```

In Cloudflare Zero Trust, route your public hostname to:
- `http://vibecontainer:7681` for read-only stream
- `http://vibecontainer:7682` for interactive stream (if enabled)

`docker-compose.yml` example:

```yaml
services:
  vibecontainer:
    image: ghcr.io/openhoo/vibecontainer:latest
    container_name: vibecontainer
    cap_add:
      - NET_ADMIN
      - NET_RAW
    environment:
      TMUX_WEB_ENABLE: "1"
      TMUX_WEB_INTERACTIVE_ENABLE: "0"
      TTYD_CREDENTIAL: "user:change-me"
    ports:
      - "127.0.0.1:7681:7681"
      - "127.0.0.1:7682:7682"
    restart: unless-stopped

  cloudflared:
    image: cloudflare/cloudflared:latest
    container_name: cloudflared
    depends_on:
      - vibecontainer
    command: tunnel --no-autoupdate run --token ${TUNNEL_TOKEN}
    network_mode: "service:vibecontainer"
    restart: unless-stopped
```

Create `.env` next to the compose file:

```env
TUNNEL_TOKEN=<your-cloudflare-tunnel-token>
```

When exposing remotely, protect access with:
- `TTYD_CREDENTIAL` on the container, and/or
- Cloudflare Access policies in Zero Trust

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

## Nix Development

Use the flake to get a reproducible Go CLI development environment.

```sh
# enter the dev shell (Go, gopls, golangci-lint, docker client, etc.)
nix develop
```

```sh
# build the CLI package output
nix build .#vibecontainer
```

```sh
# run the CLI app output
nix run .#vibecontainer -- --version
```

```sh
# run flake checks for the current system
nix flake check
```

## Go CLI

`vibecontainer` now includes a Go CLI for creating and managing provider-aware stacks.

```sh
# local build without creating ./vibecontainer in repo root
go build -o /tmp/vibecontainer ./cmd/vibecontainer
```

```sh
# show commands
/tmp/vibecontainer --help
```

```sh
# guided TUI create flow
vibecontainer create
```

```sh
# open current directory as workspace (like `code .`)
vibecontainer create .
```

By default, `vibecontainer create` does not bind-mount a host workspace.
Pass a trailing path (for example `.`) to opt in to mapping.

```sh
# non-interactive create
vibecontainer create --yes \
  --name my-stack \
  --provider codex \
  --readonly-port 7681 \
  --tunnel-token <token> \
  --openai-api-key <key> \
  .
```

```sh
# lifecycle
vibecontainer list
vibecontainer status --name my-stack
vibecontainer stop --name my-stack
vibecontainer start --name my-stack
vibecontainer restart --name my-stack
vibecontainer logs --name my-stack --follow
vibecontainer remove --name my-stack --yes
```

### Credential Management

The CLI securely stores OAuth tokens and API keys in your system keychain (macOS Keychain, Windows Credential Manager, or Linux Secret Service) so you don't need to re-enter them every time.

```sh
# View which credentials are stored (without showing values)
vibecontainer credentials list

# Clear all stored credentials
vibecontainer credentials clear
```

When creating a stack:
- The CLI automatically loads previously saved credentials
- You can press Enter to use saved credentials or type new values
- New credentials are automatically saved for next time
- Use `--no-save-auth` to skip saving credentials to keychain

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
