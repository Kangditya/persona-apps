.PHONY: install dev dev-web dev-api infra-up infra-down \
	build lint typecheck test format validate

install:
	pnpm install
	cd apps/api && go mod tidy

infra-up:
	docker compose -f infrastructure/compose.yaml up -d postgres

infra-down:
	docker compose -f infrastructure/compose.yaml down

dev:
	pnpm dev

dev-web:
	pnpm --parallel \
		--filter @brand-commerce/operations-web \
		--filter @brand-commerce/storefront-web \
		dev

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

validate: lint typecheck test build
	docker compose -f infrastructure/compose.yaml config >/dev/null
