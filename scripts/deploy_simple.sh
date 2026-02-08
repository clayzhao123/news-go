#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [ ! -d .venv ]; then
  python -m venv .venv
fi

# shellcheck disable=SC1091
source .venv/bin/activate
pip install -r requirements.txt

python -m src.digest_job
cp -n .env.example .env || true

echo "[ok] digest generated -> data/daily_digest.json"
echo "[ok] starting go server at http://localhost:8080"

go run ./cmd/api
