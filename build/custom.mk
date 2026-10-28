# Include custom targets and environment variables here

E2E_DIR = e2e/playwright

$(E2E_DIR)/node_modules: $(E2E_DIR)/package.json $(E2E_DIR)/package-lock.json
	cd $(E2E_DIR) && npm ci && npx playwright install chromium
	touch $@

## Starts the disposable Mattermost server used by the E2E suite.
.PHONY: e2e-start
e2e-start:
	cd $(E2E_DIR) && docker compose up -d --wait

## Stops the E2E server and wipes its data.
.PHONY: e2e-stop
e2e-stop:
	cd $(E2E_DIR) && docker compose down -v

## Builds the plugin and runs the Playwright E2E suite against PW_BASE_URL (default http://localhost:8065).
.PHONY: e2e-test
e2e-test: dist $(E2E_DIR)/node_modules
	cd $(E2E_DIR) && npm run tsc && npm test

# If there's no MM_RUDDER_PLUGINS_PROD, add DEV data
RUDDER_WRITE_KEY = 1d5bMvdrfWClLxgK1FvV3s4U1tg
ifdef MM_RUDDER_PLUGINS_PROD
	RUDDER_WRITE_KEY = $(MM_RUDDER_PLUGINS_PROD)
endif
LDFLAGS += -X "$(REPOSITORY_URL)/server/telemetry.rudderWriteKey=$(RUDDER_WRITE_KEY)"

# Build info
BUILD_DATE = $(shell date -u)
BUILD_HASH = $(shell git rev-parse HEAD)
BUILD_HASH_SHORT = $(shell git rev-parse --short HEAD)
LDFLAGS += -X "main.BuildDate=$(BUILD_DATE)"
LDFLAGS += -X "main.BuildHash=$(BUILD_HASH)"
LDFLAGS += -X "main.BuildHashShort=$(BUILD_HASH_SHORT)"

GO_BUILD_FLAGS = -ldflags '$(LDFLAGS)'

# Generates mock golang interfaces for testing
mock:
ifneq ($(HAS_SERVER),)
	go install github.com/golang/mock/mockgen@v1.6.0
	mockgen -destination calendar/jobs/mock_cluster/mock_cluster.go github.com/mattermost/mattermost/server/public/pluginapi/cluster JobPluginAPI
	mockgen -destination calendar/engine/mock_engine/mock_engine.go $(REPOSITORY_URL)/calendar/engine Engine
	mockgen -destination calendar/engine/mock_welcomer/mock_welcomer.go -package mock_welcomer $(REPOSITORY_URL)/calendar/engine Welcomer
	mockgen -destination calendar/engine/mock_plugin_api/mock_plugin_api.go -package mock_plugin_api $(REPOSITORY_URL)/calendar/engine PluginAPI
	mockgen -destination calendar/remote/mock_remote/mock_remote.go $(REPOSITORY_URL)/calendar/remote Remote
	mockgen -destination calendar/remote/mock_remote/mock_client.go $(REPOSITORY_URL)/calendar/remote Client
	mockgen -destination calendar/utils/bot/mock_bot/mock_poster.go $(REPOSITORY_URL)/calendar/utils/bot Poster
	mockgen -destination calendar/utils/bot/mock_bot/mock_admin.go $(REPOSITORY_URL)/calendar/utils/bot Admin
	mockgen -destination calendar/utils/bot/mock_bot/mock_logger.go $(REPOSITORY_URL)/calendar/utils/bot Logger
	mockgen -destination calendar/store/mock_store/mock_store.go $(REPOSITORY_URL)/calendar/store Store
endif

clean_mock:
ifneq ($(HAS_SERVER),)
	rm -rf ./calendar/jobs/mock_cluster
	rm -rf ./calendar/engine/mock_engine
	rm -rf ./calendar/engine/mock_welcomer
	rm -rf ./calendar/engine/mock_plugin_api
	rm -rf ./calendar/remote/mock_remote
	rm -rf ./calendar/utils/bot/mock_bot
	rm -rf ./calendar/store/mock_store
endif
