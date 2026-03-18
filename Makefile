.PHONY: dev-web dev-server test-web test-server build-web build-server

dev-web:
	cd web && pnpm dev

dev-server:
	cd server && go run ./cmd/api

test-web:
	cd web && pnpm test:run

test-server:
	cd server && go test ./...

build-web:
	cd web && pnpm build

build-server:
	cd server && go build ./cmd/api

