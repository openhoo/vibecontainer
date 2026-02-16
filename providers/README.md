# Provider Flavors

This directory contains flavor-specific assets for the unified `vibecontainer`
repository.

File-based auth env vars (`*_FILE`) are strict by design: if set, the path must
exist and be readable or container startup fails.

Current flavors:
- `claude`: Claude Code runtime defaults and config files
- `codex`: Codex CLI runtime defaults

Flavor images are built from the root `Dockerfile` targets:
- `--target claude`
- `--target codex`

Published tags:
- base image: `latest`, `vX.Y.Z`, `sha-<commit>`
- claude image: `claude`, `claude-vX.Y.Z`, `claude-code-<version>`
- codex image: `codex`, `codex-vX.Y.Z`, `codex-cli-<version>`
