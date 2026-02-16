#!/bin/bash
set -euo pipefail

provider_read_secret_from_file_env() {
    local value_var="$1"
    local file_var="$2"
    local file_path=""

    if [ "${!file_var+x}" = "x" ]; then
        file_path="${!file_var}"
        if [ -z "$file_path" ]; then
            echo "Error: ${file_var} is set but empty."
            exit 1
        fi

        if [ ! -e "$file_path" ]; then
            echo "Error: ${file_var} points to '${file_path}', but that path does not exist."
            exit 1
        fi

        if [ ! -f "$file_path" ]; then
            echo "Error: ${file_var} points to '${file_path}', but that path is not a file."
            exit 1
        fi

        if [ ! -r "$file_path" ]; then
            echo "Error: ${file_var} points to '${file_path}', but that file is not readable."
            exit 1
        fi

        printf -v "$value_var" '%s' "$(< "$file_path")"
        # shellcheck disable=SC2163
        export "$value_var"
    fi
}

provider_command_basename_is() {
    local cmd_path="$1"
    local expected="$2"
    local cmd_basename

    cmd_basename="$(basename -- "$cmd_path")"
    [ "$cmd_basename" = "$expected" ]
}

ensure_provider_defaults() {
    if [ -z "${PROVIDER_NAME:-}" ]; then
        echo "Error: PROVIDER_NAME is required."
        exit 1
    fi

    if [ "${#PROVIDER_DEFAULT_CMD[@]}" -eq 0 ]; then
        echo "Error: PROVIDER_DEFAULT_CMD must contain at least one command argument."
        exit 1
    fi
}

provider_entrypoint_main() {
    local cmd=()

    ensure_provider_defaults

    if [ "$#" -gt 0 ]; then
        cmd=("$@")
    else
        cmd=("${PROVIDER_DEFAULT_CMD[@]}")
    fi

    if command -v provider_auth_setup >/dev/null 2>&1; then
        provider_auth_setup "${cmd[0]}"
    fi

    exec /usr/local/bin/entrypoint.sh "${cmd[@]}"
}
