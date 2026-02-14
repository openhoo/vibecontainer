FROM debian:13-slim

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG TTYD_VERSION=1.7.7

# Package versions are intentionally unpinned to consume Debian security updates
# on image rebuilds while retaining a slim, provider-agnostic base image.
# hadolint ignore=DL3008,DL3005
RUN apt-get update && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends \
    bash \
    curl \
    ca-certificates \
    git \
    openssh-client \
    jq \
    ripgrep \
    fd-find \
    tree \
    less \
    build-essential \
    procps \
    tmux \
    locales \
    ufw \
    gosu \
    && rm -rf /var/lib/apt/lists/* \
    && sed -i '/en_US.UTF-8/s/^# //' /etc/locale.gen && locale-gen \
    && ln -sf /usr/bin/fdfind /usr/local/bin/fd

RUN set -eux; \
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
    rm -f /tmp/ttyd /tmp/SHA256SUMS

ENV LANG=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8

ARG USERNAME=dev
ARG USER_UID=1000
ARG USER_GID=1000
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
