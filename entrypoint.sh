#!/bin/bash
set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
TMUX_SESSION_NAME="${TMUX_SESSION_NAME:-vibe}"
TMUX_WEB_ENABLE="${TMUX_WEB_ENABLE:-0}"
TMUX_WEB_BIND_ADDRESS="${TMUX_WEB_BIND_ADDRESS:-0.0.0.0}"
TMUX_WEB_READONLY_PORT="${TMUX_WEB_READONLY_PORT:-7681}"
TMUX_WEB_INTERACTIVE_ENABLE="${TMUX_WEB_INTERACTIVE_ENABLE:-0}"
TMUX_WEB_INTERACTIVE_PORT="${TMUX_WEB_INTERACTIVE_PORT:-7682}"
FIREWALL_ENABLE="${FIREWALL_ENABLE:-1}"

# ---------------------------------------------------------------------------
# Validation
# ---------------------------------------------------------------------------
validate_port() {
    local name="$1" value="$2"
    if ! [[ "$value" =~ ^[0-9]+$ ]] || [ "$value" -lt 1 ] || [ "$value" -gt 65535 ]; then
        echo "Error: $name must be an integer between 1 and 65535 (got: '$value')."
        exit 1
    fi
}

if [ "$TMUX_WEB_ENABLE" = "1" ]; then
    validate_port "TMUX_WEB_READONLY_PORT" "$TMUX_WEB_READONLY_PORT"
    if [ "$TMUX_WEB_INTERACTIVE_ENABLE" = "1" ]; then
        validate_port "TMUX_WEB_INTERACTIVE_PORT" "$TMUX_WEB_INTERACTIVE_PORT"
        if [ "$TMUX_WEB_READONLY_PORT" = "$TMUX_WEB_INTERACTIVE_PORT" ]; then
            echo "Error: TMUX_WEB_READONLY_PORT and TMUX_WEB_INTERACTIVE_PORT must differ."
            exit 1
        fi
    fi
fi

if [ "$FIREWALL_ENABLE" != "0" ] && [ "$FIREWALL_ENABLE" != "1" ]; then
    echo "Error: FIREWALL_ENABLE must be '0' or '1' (got: '$FIREWALL_ENABLE')."
    exit 1
fi

# ---------------------------------------------------------------------------
# Firewall — configure and enable ufw (restrict outbound traffic to web and SSH)
# ---------------------------------------------------------------------------
run_ufw() {
    if ! ufw "$@"; then
        echo "Error: failed to run 'ufw $*'."
        echo "This may indicate that ufw is not properly initialized or that the container"
        echo "lacks the required NET_ADMIN and NET_RAW capabilities to manage firewall rules."
        exit 1
    fi
}

if [ "$FIREWALL_ENABLE" = "1" ]; then
    # Ensure ufw is available before attempting firewall configuration.
    if ! command -v ufw >/dev/null 2>&1; then
        echo "Error: 'ufw' command not found. Firewall configuration cannot be applied."
        echo "Ensure the container image includes ufw and has NET_ADMIN and NET_RAW capabilities."
        exit 1
    fi

    run_ufw default deny incoming
    run_ufw default deny outgoing

    # Explicitly allow loopback (localhost) traffic so inter-process communication
    # within this container is not affected by the restrictive firewall policy. This
    # only applies to the container's own network namespace and does not affect host
    # loopback connectivity.
    run_ufw allow in on lo
    run_ufw allow out on lo

    run_ufw allow out 53/tcp     # DNS (TCP)
    run_ufw allow out 53/udp     # DNS (UDP)
    run_ufw allow out 80/tcp     # HTTP
    run_ufw allow out 443/tcp    # HTTPS
    run_ufw allow out 22/tcp     # SSH (e.g., Git over SSH)

    if [ "$TMUX_WEB_ENABLE" = "1" ]; then
        run_ufw allow in "$TMUX_WEB_READONLY_PORT/tcp"
        if [ "$TMUX_WEB_INTERACTIVE_ENABLE" = "1" ]; then
            run_ufw allow in "$TMUX_WEB_INTERACTIVE_PORT/tcp"
        fi
    fi

    run_ufw --force enable
else
    echo "Warning: FIREWALL_ENABLE=0, skipping ufw setup and capability checks."
fi

# ---------------------------------------------------------------------------
# Drop privileges — run remaining commands as non-root user
# ---------------------------------------------------------------------------
export HOME="/home/dev"

# ---------------------------------------------------------------------------
# Build command
# ---------------------------------------------------------------------------
quote_command_args() {
    local cmd
    if [ "$#" -eq 0 ]; then
        # Default to an interactive login shell when no command is provided.
        set -- bash -l
    fi

    local arg first=1
    cmd=""
    for arg in "$@"; do
        if [ "$first" -eq 1 ]; then
            cmd="$(printf '%q' "$arg")"
            first=0
        else
            cmd+=" $(printf '%q' "$arg")"
        fi
    done
    printf '%s' "$cmd"
}

# ---------------------------------------------------------------------------
# Start tmux session
# ---------------------------------------------------------------------------
if ! gosu dev tmux has-session -t "$TMUX_SESSION_NAME" 2>/dev/null; then
    gosu dev tmux -u new-session -d -s "$TMUX_SESSION_NAME" "$(quote_command_args "$@")"
fi

# ---------------------------------------------------------------------------
# No-web mode: poll for tmux session and exit when it ends
# ---------------------------------------------------------------------------
if [ "$TMUX_WEB_ENABLE" != "1" ]; then
    echo "Web streaming disabled. Attach with: docker exec -it <container> tmux attach -t $TMUX_SESSION_NAME"
    while gosu dev tmux has-session -t "$TMUX_SESSION_NAME" 2>/dev/null; do
        sleep 5
    done
    exit 0
fi

# ---------------------------------------------------------------------------
# Web mode: start ttyd instances
# ---------------------------------------------------------------------------
child_pids=()

cleanup() {
    gosu dev tmux kill-session -t "$TMUX_SESSION_NAME" 2>/dev/null || true
    local pid
    for pid in "${child_pids[@]}"; do
        kill "$pid" 2>/dev/null || true
    done
    if [ "${#child_pids[@]}" -gt 0 ]; then
        wait "${child_pids[@]}" 2>/dev/null || true
    fi
}

trap cleanup EXIT INT TERM

# Optional basic-auth for ttyd (format: "user:password")
ttyd_auth_args=()
if [ -n "${TTYD_CREDENTIAL:-}" ]; then
    ttyd_auth_args+=(-c "$TTYD_CREDENTIAL")
fi

gosu dev ttyd "${ttyd_auth_args[@]}" -i "$TMUX_WEB_BIND_ADDRESS" -p "$TMUX_WEB_READONLY_PORT" \
    tmux attach-session -r -t "$TMUX_SESSION_NAME" &
child_pids+=("$!")

if [ "$TMUX_WEB_INTERACTIVE_ENABLE" = "1" ]; then
    gosu dev ttyd "${ttyd_auth_args[@]}" -W -i "$TMUX_WEB_BIND_ADDRESS" -p "$TMUX_WEB_INTERACTIVE_PORT" \
        tmux attach-session -t "$TMUX_SESSION_NAME" &
    child_pids+=("$!")
fi

wait -n
