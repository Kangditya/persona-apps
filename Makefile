-include .env
export

.PHONY: install dev dev-web dev-storefront dev-operations dev-api infra-up infra-down \
	build lint typecheck test format format-check compose-check validate

install:
	pnpm install
	cd apps/api && go mod download

infra-up:
	docker compose -f infrastructure/compose.yaml up -d postgres

infra-down:
	docker compose -f infrastructure/compose.yaml down

dev:
	$(MAKE) -j3 dev-storefront dev-operations dev-api

dev-web:
	pnpm dev

dev-storefront:
	pnpm --filter @persona-apps/storefront-web dev

dev-operations:
	pnpm --filter @persona-apps/operations-web dev

dev-api:
	@test -x "$$(command -v air 2>/dev/null)" || (printf 'air is required for API live reload. Install it with: go install github.com/air-verse/air@latest\n' >&2; exit 1)
	@set -eu; \
	requested_address="$${HTTP_ADDRESS:-:8080}"; \
	requested_port="$${HTTP_PORT:-$${requested_address##*:}}"; \
	port="$$requested_port"; \
	while ss -ltnH | grep -Eq ":$${port}[[:space:]]"; do \
		port=$$((port + 1)); \
	done; \
	address="$${requested_address%:*}:$${port}"; \
	if [ "$$address" != "$$requested_address" ]; then \
		printf 'API port %s is occupied; using %s instead.\n' "$$requested_port" "$$address"; \
	fi; \
	cd apps/api && HTTP_ADDRESS="$$address" air

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
