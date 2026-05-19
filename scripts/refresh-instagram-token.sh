#!/usr/bin/env sh
set -eu

if [ ! -f .env ]; then
  echo "Missing .env file" >&2
  exit 1
fi

set -a
. ./.env
set +a

if [ -z "${ADMIN_TOKEN:-}" ]; then
  echo "ADMIN_TOKEN is not set in .env" >&2
  exit 1
fi

if [ -z "${HOST:-}" ]; then
  echo "HOST is not set in .env" >&2
  exit 1
fi

url="https://${HOST}/instagram/refresh-access-token"

curl -fsS \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  "${url}"

printf "\n"
