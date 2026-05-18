#!/bin/bash
set -e

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

# Só carrega .env se o token não estiver definido
if [[ -z "$JIRA_API_TOKEN" && -f "$SCRIPT_DIR/.env" ]]; then
    set -a
    source "$SCRIPT_DIR/.env"
    set +a
fi

exec "$SCRIPT_DIR/bin/mcp"
