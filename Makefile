.PHONY: dev-web test-web build-web check-web

dev-web:
	cd web && pnpm dev

test-web:
	cd web && pnpm test:run

build-web:
	cd web && pnpm build

check-web:
	cd web && pnpm check
