#!/bin/bash

set -euo pipefail

LOG_FILE="backend.log"
PID_FILE="backend.pid"

if [[ -f "$PID_FILE" ]]; then
    existing_pid=$(cat "$PID_FILE")
    if kill -0 "$existing_pid" 2>/dev/null; then
        echo "Backend already running (pid $existing_pid)"
        exit 0
    fi
    rm -f "$PID_FILE"
fi

nohup go run . > "$LOG_FILE" 2>&1 &
echo $! > "$PID_FILE"

echo "Backend started with pid $(cat "$PID_FILE")"
echo "Logs: $LOG_FILE"
