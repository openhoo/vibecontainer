# Vibecontainer

`vibecontainer` is a provider-agnostic base runtime for AI CLI workflows.

This repository contains the base image for vibecoding containers and is the shared foundation for:
- [claude-container](https://github.com/openhoo/claude-container)
- [codex-container](https://github.com/openhoo/codex-container)
- [gemini-container](https://github.com/openhoo/gemini-container)

It provides:
- Debian-based development environment
- non-root `dev` user and `/workspace` working directory
- `tmux` session lifecycle management
- optional browser streaming via `ttyd`
- strict-by-default `ufw` firewall policy

It does not install provider-specific CLIs (Claude, Codex, etc.).

## Quick Start

```sh
# 1. Clone the repository
git clone https://github.com/openhoo/vibecontainer.git && cd vibecontainer

# 2. Create .env (required by docker run --env-file)
touch .env

# 3. Build and run
docker build -t vibecontainer .
docker run -d --name vibecontainer --env-file .env \
  --cap-add NET_ADMIN --cap-add NET_RAW \
  -e TMUX_WEB_ENABLE=1 \
  -p 127.0.0.1:7681:7681 \
  -v "$(pwd):/workspace" \
  vibecontainer

# 4. Open read-only stream
open http://127.0.0.1:7681

# Or attach directly
docker exec -it vibecontainer tmux attach -t vibe
```

## Docker Compose

A `docker-compose.yml` is provided to run `vibecontainer` alongside a [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) (`cloudflared`) for remote access.

```sh
# 1. Create .env with your Cloudflare Tunnel token
echo "TUNNEL_TOKEN=your-tunnel-token" > .env

# 2. Start both services
docker compose up -d

# 3. Attach to the tmux session
docker exec -it vibecontainer tmux attach -t vibe
```

Configure the tunnel's public hostname in the Cloudflare Zero Trust dashboard to point to `http://vibecontainer:7681` for the read-only stream. If you've enabled the interactive stream in your Docker configuration (for example by setting `TMUX_WEB_INTERACTIVE_ENABLE=1` and exposing port `7682`), you can instead point it to `http://vibecontainer:7682` for interactive access.

When exposing `vibecontainer` via Cloudflare Tunnel (or any remote access), the ttyd endpoints are no longer protected by the localhost-only bindings used in the quick start examples. For remote access, you should configure basic authentication by setting `TTYD_CREDENTIAL` (for example in your `.env` file or `docker-compose.yml`) and/or enforce strong access controls in Cloudflare Zero Trust before making these endpoints publicly reachable.

## Common Commands

| Command | Description |
|--------|-------------|
| `docker build -t vibecontainer .` | Build the Docker image |
| `docker run -d --name vibecontainer --env-file .env --cap-add NET_ADMIN --cap-add NET_RAW -e TMUX_WEB_ENABLE=1 -p 127.0.0.1:7681:7681 -v "$(pwd):/workspace" vibecontainer` | Run with strict firewall + read-only browser stream |
| `docker run -d --name vibecontainer --env-file .env --cap-add NET_ADMIN --cap-add NET_RAW -e TMUX_WEB_ENABLE=1 -e TMUX_WEB_INTERACTIVE_ENABLE=1 -p 127.0.0.1:7681:7681 -p 127.0.0.1:7682:7682 -v "$(pwd):/workspace" vibecontainer` | Run with strict firewall + read-only + interactive streams |
| `docker run -d --name vibecontainer --env-file .env -e FIREWALL_ENABLE=0 -e TMUX_WEB_ENABLE=1 -p 127.0.0.1:7681:7681 -v "$(pwd):/workspace" vibecontainer` | Run without firewall setup/capabilities |
| `docker run -d --name vibecontainer --env-file .env -e FIREWALL_ENABLE=0 -e TMUX_WEB_ENABLE=1 -e TMUX_WEB_INTERACTIVE_ENABLE=1 -p 127.0.0.1:7681:7681 -p 127.0.0.1:7682:7682 -v "$(pwd):/workspace" vibecontainer` | Interactive run without firewall setup/capabilities |
| `docker exec -it vibecontainer tmux attach -t vibe` | Attach to the tmux session |
| `docker build -t vibecontainer . && docker run -d --name vibecontainer --env-file .env --cap-add NET_ADMIN --cap-add NET_RAW -e TMUX_WEB_ENABLE=1 -p 127.0.0.1:7681:7681 -v "$(pwd):/workspace" vibecontainer && docker exec -it vibecontainer tmux attach -t vibe` | Build, run, and attach in one step |
| `docker rm -f vibecontainer` | Stop and remove the container |
| `docker logs -f vibecontainer` | Tail container logs |
| `docker exec -it vibecontainer bash -l` | Open a bash shell in the container |

## Browser Endpoints

| Endpoint | Default | Description |
|----------|---------|-------------|
| Read-only | [http://127.0.0.1:7681](http://127.0.0.1:7681) | View-only terminal stream |
| Interactive | [http://127.0.0.1:7682](http://127.0.0.1:7682) | Full terminal access (opt-in) |

## Workspace

The container uses `/workspace` as its working directory. Mount your project there:

```sh
docker run -d --name vibecontainer --env-file .env \
  --cap-add NET_ADMIN --cap-add NET_RAW \
  -e TMUX_WEB_ENABLE=1 \
  -p 127.0.0.1:7681:7681 \
  -v "$(pwd)/my-project:/workspace" \
  vibecontainer
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TMUX_SESSION_NAME` | `vibe` | Name of the tmux session |
| `TMUX_WEB_ENABLE` | `0` | Enable ttyd browser streaming |
| `TMUX_WEB_BIND_ADDRESS` | `0.0.0.0` | Bind address for ttyd inside the container |
| `TMUX_WEB_READONLY_PORT` | `7681` | Read-only stream port |
| `TMUX_WEB_INTERACTIVE_ENABLE` | `0` | Enable the interactive stream |
| `TMUX_WEB_INTERACTIVE_PORT` | `7682` | Interactive stream port |
| `TTYD_CREDENTIAL` | *(none)* | Basic auth for ttyd (`user:password`) |
| `FIREWALL_ENABLE` | `1` | `1` enables strict ufw policy, `0` skips firewall setup |

## Security

### Permissions

The container runs commands inside tmux as a non-root user (`dev`). Root is only used during startup to optionally configure firewall rules.

Note: `docker exec` defaults to root. Use `-u dev` for non-root shells:

```sh
docker exec -it -u dev vibecontainer bash -l
```

### Network Firewall

When `FIREWALL_ENABLE=1` (default), the container enables `ufw` at startup with restrictive defaults:

| Direction | Port | Protocol | Purpose |
|-----------|------|----------|---------|
| Outbound | 53 | TCP/UDP | DNS resolution |
| Outbound | 80 | TCP | HTTP |
| Outbound | 443 | TCP | HTTPS |
| Outbound | 22 | TCP | SSH (e.g., Git over SSH) |
| Inbound | `TMUX_WEB_READONLY_PORT` | TCP | ttyd read-only stream (when web enabled) |
| Inbound | `TMUX_WEB_INTERACTIVE_PORT` | TCP | ttyd interactive stream (when enabled) |
| Loopback | — | All | Localhost inter-process communication |

`FIREWALL_ENABLE=1` requires `NET_ADMIN` and `NET_RAW` capabilities.

For local development, you can explicitly bypass firewall setup with `FIREWALL_ENABLE=0`.

### Browser Streaming

The `ttyd` web terminal has no authentication by default. Security relies on:

1. localhost-only port binding in provided `docker run` examples
2. optional basic auth via `TTYD_CREDENTIAL`
3. interactive stream disabled unless explicitly enabled

## License

MIT

## Provider Extension Contract

`vibecontainer` owns tmux/web/firewall lifecycle. Provider repos should customize startup via container args or Docker `CMD` only.
Guidance:

1. Set provider command in Docker `CMD` (or pass args at runtime).
2. Keep base `ENTRYPOINT` unchanged unless you need non-standard behavior.
3. Reuse base env knobs (`TMUX_*`, `TTYD_CREDENTIAL`, `FIREWALL_ENABLE`) instead of duplicating lifecycle logic.

## Versioning and Pinning

Base images are intended to publish semver tags (`vX.Y.Z`) plus `latest`.
The publish workflow also emits a commit-`sha` tag for traceability.

Provider images should pin to base semver tags for reproducibility, instead of relying only on `latest`.

## CI Quality Gates

Pull requests and pushes to `main` run lint checks in GitHub Actions:
- `actionlint` for workflow validation
- `shellcheck` for `entrypoint.sh`
- `hadolint` for `Dockerfile`

## Provider Repos

This base image is intended to be consumed by:
- [claude-container](https://github.com/openhoo/claude-container)
- [codex-container](https://github.com/openhoo/codex-container)
- [gemini-container](https://github.com/openhoo/gemini-container)
