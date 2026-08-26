#!/usr/bin/env bash
set -euo pipefail

port="${REDIS_PORT:-6379}"
host="127.0.0.1"

if ! command -v redis-server >/dev/null 2>&1 || ! command -v redis-cli >/dev/null 2>&1; then
  echo "redis-server and redis-cli are required to bootstrap Redis" >&2
  exit 1
fi

redis_ready() {
  redis-cli -h "$host" -p "$port" ping 2>/dev/null | grep -qx PONG
}

wait_for_redis() {
  for _ in $(seq 1 30); do
    if redis_ready; then
      return 0
    fi
    sleep 1
  done
  return 1
}

if redis_ready; then
  echo "Redis is already ready on ${host}:${port}"
  exit 0
fi

if command -v systemctl >/dev/null 2>&1; then
  for unit in redis-server.service redis.service; do
    if systemctl list-unit-files --type=service --no-legend "$unit" 2>/dev/null | grep -q .; then
      echo "Starting systemd unit ${unit}"
      if sudo systemctl start "$unit" && wait_for_redis; then
        echo "Redis is ready on ${host}:${port}"
        exit 0
      fi
      echo "systemd unit ${unit} did not become ready; trying the next bootstrap method" >&2
    fi
  done
fi

if command -v service >/dev/null 2>&1; then
  echo "Trying redis-server through the service command"
  if sudo service redis-server start && wait_for_redis; then
    echo "Redis is ready on ${host}:${port}"
    exit 0
  fi
  echo "service redis-server did not become ready; trying redis-server directly" >&2
fi

run_dir="${RUNNER_TEMP:-/tmp}/law-oa-ci-redis"
mkdir -p "$run_dir"
if ! redis-server \
      --daemonize yes \
      --bind "$host" \
      --port "$port" \
      --dir "$run_dir" \
      --pidfile "${run_dir}/redis.pid" \
      --dbfilename dump.rdb \
      --save "" \
      --appendonly no; then
  echo "redis-server failed to start directly" >&2
  exit 1
fi

if wait_for_redis; then
  echo "Redis is ready on ${host}:${port}"
  exit 0
fi

echo "Redis did not become ready on ${host}:${port}" >&2
exit 1
