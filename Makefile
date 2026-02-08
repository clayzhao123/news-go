.PHONY: run run-go setup-py digest run-all test vet fmt

run: run-go

run-go:
	go run ./cmd/api

setup-py:
	python -m venv .venv
	. .venv/bin/activate && pip install -r requirements.txt

digest:
	python -m src.digest_job

# One-command local deploy (refresh digest then start Go UI/API)
run-all: digest run-go

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal
