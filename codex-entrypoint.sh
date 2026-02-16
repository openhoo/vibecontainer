#!/bin/bash
set -euo pipefail

# shellcheck disable=SC2034
PROVIDER_NAME="codex"
PROVIDER_DEFAULT_CMD=("codex")

CODEX_AUTH_PATH="/home/dev/.codex/auth.json"

write_codex_auth_json_from_payload() {
    local payload="$1"
    local tmpfile

    if ! command -v node >/dev/null 2>&1; then
        echo "Error: node is required to configure Codex OAuth auth mode."
        exit 1
    fi

    tmpfile="$(mktemp)"
    if ! CODEX_AUTH_JSON_PAYLOAD="$payload" node -e '
const payload = process.env.CODEX_AUTH_JSON_PAYLOAD || "";
let parsed;

try {
  parsed = JSON.parse(payload);
} catch (error) {
  console.error("Error: CODEX auth JSON is not valid JSON.");
  process.exit(1);
}

if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
  console.error("Error: CODEX auth JSON must be a JSON object.");
  process.exit(1);
}

if (
  Object.prototype.hasOwnProperty.call(parsed, "auth_mode")
  && !["apikey", "chatgpt", "chatgptAuthTokens"].includes(parsed.auth_mode)
) {
  console.error("Error: CODEX auth JSON auth_mode must be one of: apikey, chatgpt, chatgptAuthTokens.");
  process.exit(1);
}

const hasApiKeyShape = typeof parsed.OPENAI_API_KEY === "string" && parsed.OPENAI_API_KEY.trim().length > 0;
const tokens = parsed.tokens;
const hasTokenShape = tokens
  && typeof tokens === "object"
  && !Array.isArray(tokens)
  && typeof tokens.id_token === "string"
  && typeof tokens.access_token === "string"
  && typeof tokens.refresh_token === "string";

if (!hasApiKeyShape && !hasTokenShape) {
  console.error("Error: CODEX auth JSON must include either OPENAI_API_KEY (non-empty string) or tokens.id_token/tokens.access_token/tokens.refresh_token (strings).");
  process.exit(1);
}

process.stdout.write(JSON.stringify(parsed, null, 2));
' > "$tmpfile"; then
        rm -f "$tmpfile"
        exit 1
    fi

    mkdir -p /home/dev/.codex
    chown dev:dev /home/dev/.codex
    install -m 0600 -o dev -g dev "$tmpfile" "$CODEX_AUTH_PATH"
    rm -f "$tmpfile"
}

provider_auth_setup() {
    local cmd_name="$1"

    if ! provider_command_basename_is "$cmd_name" "codex"; then
        return
    fi

    provider_read_secret_from_file_env "OPENAI_API_KEY" "OPENAI_API_KEY_FILE"
    provider_read_secret_from_file_env "CODEX_API_KEY" "CODEX_API_KEY_FILE"
    provider_read_secret_from_file_env "CODEX_AUTH_JSON" "CODEX_AUTH_JSON_FILE"

    # Auth precedence:
    # 1) CODEX_AUTH_JSON* 2) API keys.
    if [ -n "${CODEX_AUTH_JSON:-}" ]; then
        write_codex_auth_json_from_payload "${CODEX_AUTH_JSON}"
        unset OPENAI_API_KEY CODEX_API_KEY
    fi
}

# shellcheck source=provider-entrypoint-base.sh
source /usr/local/bin/provider-entrypoint-base.sh

provider_entrypoint_main "$@"
