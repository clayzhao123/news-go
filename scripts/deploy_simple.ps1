$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

if (!(Test-Path .venv)) {
  python -m venv .venv
}

& .\.venv\Scripts\python -m pip install -r requirements.txt
& .\.venv\Scripts\python -m src.digest_job

if (!(Test-Path .env)) {
  Copy-Item .env.example .env
}

Write-Host "[ok] digest generated -> data/daily_digest.json"
Write-Host "[ok] starting go server at http://localhost:8080"

go run ./cmd/api
