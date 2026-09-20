.PHONY: help build-server run-server dev-webapp build-webapp test-server test-public

help: ## Show this help message
	@echo "Mattermost Accelerated Build Commands:"
	@echo "  make build-server   - Fast compiles Mattermost Go server (32-core parallel, stripped symbols)"
	@echo "  make run-server     - Fast runs Mattermost Go server directly"
	@echo "  make dev-webapp     - Starts Webapp Webpack Dev Server with HMR (port 9005)"
	@echo "  make build-webapp   - Builds Webapp production bundle"
	@echo "  make test-server    - Runs fast unit tests for server"
	@echo "  make test-public    - Runs unit tests for public models"

build-server:
	@cd server && $(MAKE) build-server-fast

run-server:
	@cd server && $(MAKE) run-server-fast

dev-webapp:
	@cd webapp && npm run dev-server

build-webapp:
	@cd webapp && npm run build

test-public:
	@cd server/public && go test -short -v ./model

test-server:
	@cd server && go test -short -p 32 ./channels/app/...
