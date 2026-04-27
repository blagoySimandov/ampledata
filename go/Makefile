run:
	@docker compose -f docker-compose.dev.yml up --build

dev:
	@trap 'kill 0' SIGINT; \
	(cd web/ampledata-fe && bun run build:watch) & \
	go run ./cmd/server/

test:
	@go test ./...

test-live:
	@AI_LIVE=1 go test ./internal/services/... -run TestLive -v -count=1
