FROM debian:13-slim AS base

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG TTYD_VERSION=1.7.7
ARG USERNAME=dev
ARG USER_UID=1000
ARG USER_GID=1000

# Package versions are intentionally unpinned to consume Debian security updates
# on image rebuilds while retaining a slim, provider-agnostic base image.
# hadolint ignore=DL3008,DL3005
RUN apt-get update && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    git \
    openssh-client \
    tmux \
    ufw \
    gosu \
    && rm -rf /var/lib/apt/lists/*

# hadolint ignore=DL3008
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends curl; \
    arch="$(dpkg --print-architecture)"; \
    case "$arch" in \
        amd64) ttyd_asset="ttyd.x86_64" ;; \
        arm64) ttyd_asset="ttyd.aarch64" ;; \
        *) echo "Unsupported architecture: $arch" >&2; exit 1 ;; \
    esac; \
    base_url="https://github.com/tsl0922/ttyd/releases/download/${TTYD_VERSION}"; \
    curl -fsSL -o /tmp/ttyd "${base_url}/${ttyd_asset}"; \
    curl -fsSL -o /tmp/SHA256SUMS "${base_url}/SHA256SUMS"; \
    expected_checksum="$(awk -v asset="$ttyd_asset" '$2 == asset { print $1 }' /tmp/SHA256SUMS)"; \
    test -n "$expected_checksum"; \
    printf '%s  %s\n' "$expected_checksum" "/tmp/ttyd" | sha256sum -c -; \
    install -m 0755 /tmp/ttyd /usr/local/bin/ttyd; \
    rm -f /tmp/ttyd /tmp/SHA256SUMS; \
    apt-get purge -y --auto-remove curl; \
    rm -rf /var/lib/apt/lists/*

ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

RUN groupadd -g "$USER_GID" "$USERNAME" && \
    useradd -l -m -u "$USER_UID" -g "$USER_GID" -s /bin/bash "$USERNAME"

RUN mkdir -p /workspace && chown "$USERNAME:$USERNAME" /workspace

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER "$USERNAME"

ENV SHELL=/bin/bash

COPY --chown=$USERNAME:$USERNAME .tmux.conf /home/$USERNAME/.tmux.conf

# Root at startup is required for optional ufw setup before privileges are
# dropped to the non-root dev user by entrypoint.sh.
# hadolint ignore=DL3002
USER root

WORKDIR /workspace

EXPOSE 7681 7682

ENTRYPOINT ["entrypoint.sh"]
CMD ["bash", "-l"]

FROM base AS claude

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG USERNAME=dev
# renovate: datasource=github-releases depName=anthropics/claude-code
ARG CLAUDE_CODE_VERSION=2.1.49

LABEL org.opencontainers.image.claude-code.version="${CLAUDE_CODE_VERSION}"

# hadolint ignore=DL3008
RUN command -v curl >/dev/null 2>&1 || { \
        apt-get update && apt-get install -y --no-install-recommends curl ca-certificates \
        && rm -rf /var/lib/apt/lists/*; \
    }

USER "$USERNAME"

RUN set -eux; \
    curl -fsSL -o /tmp/claude-install.sh https://claude.ai/install.sh; \
    bash /tmp/claude-install.sh "${CLAUDE_CODE_VERSION}"; \
    rm -f /tmp/claude-install.sh; \
    rm -rf "/home/$USERNAME/.cache"/*

ENV PATH="/home/$USERNAME/.local/bin:$PATH"

RUN mkdir -p /home/$USERNAME/.claude
COPY --chown=$USERNAME:$USERNAME providers/claude/.claude.json /home/$USERNAME/.claude.json
COPY --chown=$USERNAME:$USERNAME providers/claude/settings.json /home/$USERNAME/.claude/settings.json

# Root at startup is required for optional ufw setup before privileges are
# dropped to the non-root dev user by entrypoint.sh.
# hadolint ignore=DL3002
USER root

COPY provider-entrypoint-base.sh /usr/local/bin/provider-entrypoint-base.sh
COPY claude-entrypoint.sh /usr/local/bin/claude-entrypoint.sh
RUN chmod +x /usr/local/bin/provider-entrypoint-base.sh /usr/local/bin/claude-entrypoint.sh

ENV DISABLE_TELEMETRY=1
ENV CLAUDE_CODE_DISABLE_FEEDBACK_SURVEY=1
ENV CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1

ENTRYPOINT ["claude-entrypoint.sh"]
CMD []

FROM base AS codex

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# renovate: datasource=npm depName=@openai/codex
ARG CODEX_VERSION=0.104.0

LABEL org.opencontainers.image.codex.version="${CODEX_VERSION}"

# hadolint ignore=DL3008
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends nodejs npm; \
    npm install -g "@openai/codex@${CODEX_VERSION}"; \
    rm -rf /var/lib/apt/lists/* /root/.npm

COPY provider-entrypoint-base.sh /usr/local/bin/provider-entrypoint-base.sh
COPY codex-entrypoint.sh /usr/local/bin/codex-entrypoint.sh
RUN chmod +x /usr/local/bin/provider-entrypoint-base.sh /usr/local/bin/codex-entrypoint.sh

ENTRYPOINT ["codex-entrypoint.sh"]
CMD []

FROM base AS final
