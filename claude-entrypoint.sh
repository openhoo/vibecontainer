#!/bin/bash
set -euo pipefail

# shellcheck disable=SC2034
PROVIDER_NAME="claude"
PROVIDER_DEFAULT_CMD=("claude" "--dangerously-skip-permissions")

provider_auth_setup() {
    local cmd_name="$1"
    if ! provider_command_basename_is "$cmd_name" "claude"; then
        return
    fi

    provider_read_secret_from_file_env "CLAUDE_CODE_OAUTH_TOKEN" "CLAUDE_CODE_OAUTH_TOKEN_FILE"
    provider_read_secret_from_file_env "ANTHROPIC_API_KEY" "ANTHROPIC_API_KEY_FILE"

    if [ -z "${CLAUDE_CODE_OAUTH_TOKEN:-}" ] && [ -z "${ANTHROPIC_API_KEY:-}" ]; then
        echo "Error: Claude authentication is not configured."
        echo "Set one of: CLAUDE_CODE_OAUTH_TOKEN or ANTHROPIC_API_KEY"
        echo "You can also use file vars: CLAUDE_CODE_OAUTH_TOKEN_FILE or ANTHROPIC_API_KEY_FILE"
        exit 1
    fi
}

# shellcheck source=provider-entrypoint-base.sh
source /usr/local/bin/provider-entrypoint-base.sh

provider_entrypoint_main "$@"
