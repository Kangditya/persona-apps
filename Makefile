-include .env
export

.PHONY: install dev dev-web dev-storefront dev-operations dev-api dev-oidc infra-up infra-down \
	build lint typecheck test format format-check compose-check validate \
	db-validate db-status db-version db-migrate db-migrate-steps db-rollback \
	db-rollback-steps db-migration-create db-seed-list db-seed-status db-seed \
	db-seed-all db-seed-reference db-seed-development db-setup

DB_COMMAND := cd apps/api && go run ./cmd/db

install:
	pnpm install
	cd apps/api && go mod download

infra-up:
	docker compose -f infrastructure/compose.yaml up -d postgres

infra-down:
	docker compose -f infrastructure/compose.yaml down

dev:
	$(MAKE) db-setup db-seed-development
	$(MAKE) dev-oidc-ready
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

dev-oidc:
	cd apps/api && go run ./cmd/local-oidc

dev-oidc-ready:
	@if ! curl -fsS http://localhost:7071/.well-known/openid-configuration >/dev/null 2>&1; then \
		(setsid sh -c 'cd apps/api && exec go run ./cmd/local-oidc' >/tmp/persona-local-oidc.log 2>&1 &) ; \
	fi; \
	for attempt in $$(seq 1 30); do \
		if curl -fsS http://localhost:7071/.well-known/openid-configuration >/dev/null 2>&1; then exit 0; fi; \
		sleep 1; \
	done; \
	printf 'local OIDC issuer did not become ready on port 7071\n' >&2; exit 1

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
	cd apps/api && go run ./cmd/format -write

format-check:
	pnpm format:check
	cd apps/api && go run ./cmd/format -check

compose-check:
	docker compose -f infrastructure/compose.yaml config >/dev/null

validate: format-check lint typecheck test build compose-check

db-validate:
	$(DB_COMMAND) migrate validate

db-status:
	$(DB_COMMAND) migrate status

db-version:
	$(DB_COMMAND) migrate version

db-migrate:
	$(DB_COMMAND) migrate up

db-migrate-steps:
	@test -n "$(steps)" || (printf 'usage: make db-migrate-steps steps=N\n' >&2; exit 2)
	$(DB_COMMAND) migrate up --steps "$(steps)"

db-rollback:
	$(DB_COMMAND) migrate down --steps 1

db-rollback-steps:
	@test -n "$(steps)" || (printf 'usage: make db-rollback-steps steps=N\n' >&2; exit 2)
	$(DB_COMMAND) migrate down --steps "$(steps)"

db-migration-create:
	@test -n "$(name)" || (printf 'usage: make db-migration-create name=NAME\n' >&2; exit 2)
	$(DB_COMMAND) migrate create "$(name)"

db-seed-list:
	$(DB_COMMAND) seed list

db-seed-status:
	$(DB_COMMAND) seed status

db-seed:
	@test -n "$(name)" || (printf 'usage: make db-seed name=SEED_NAME\n' >&2; exit 2)
	$(DB_COMMAND) seed run "$(name)"

db-seed-all:
	$(DB_COMMAND) seed run --all

db-seed-reference:
	$(DB_COMMAND) seed run --group reference

db-seed-development:
	$(DB_COMMAND) seed run --group development

db-setup:
	$(DB_COMMAND) setup
