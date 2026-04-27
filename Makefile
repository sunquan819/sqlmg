.PHONY: dev dev-backend dev-frontend build clean

PORT ?= 9180

dev: dev-backend dev-frontend

dev-backend:
	go run ./cmd/sqlmg

dev-frontend:
	cd web && npm run dev -- --port 5180

build-frontend:
	cd web && npm install && npm run build

build: build-frontend
	go build -ldflags="-s -w" -o sqlmg.exe ./cmd/sqlmg

clean:
	rm -rf web/dist web/.svelte-kit web/node_modules sqlmg.exe test_sample.db
