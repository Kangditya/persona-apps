-include .env
export

.PHONY: install dev dev-web dev-api infra-up infra-down \
	build lint typecheck test format format-check compose-check validate

install:
	pnpm install
	cd apps/api && go mod download

infra-up:
	docker compose -f infrastructure/compose.yaml up -d postgres

infra-down:
	docker compose -f infrastructure/compose.yaml down

dev:
	$(MAKE) -j2 dev-web dev-api

dev-web:
	pnpm dev

dev-api:
	cd apps/api && go run ./cmd/api

build:
	pnpm build
	cd apps/api && go build ./...

lint:
	pnpm lint
	cd apps/api && go vet ./...

typecheck:
	pnpm typecheck

test:
	pnpm test
	cd apps/api && go test ./...

format:
	pnpm format
	cd apps/api && gofmt -w .

format-check:
	pnpm format:check
	test -z "$$(cd apps/api && gofmt -l .)"

compose-check:
	docker compose -f infrastructure/compose.yaml config >/dev/null

validate: format-check lint typecheck test build compose-check
