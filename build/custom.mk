# Include custom targets and environment variables here

# If there's no MM_RUDDER_PLUGINS_PROD, add DEV data
RUDDER_WRITE_KEY = 1d5bMvdrfWClLxgK1FvV3s4U1tg
ifdef MM_RUDDER_PLUGINS_PROD
	RUDDER_WRITE_KEY = $(MM_RUDDER_PLUGINS_PROD)
endif
LDFLAGS += -X "github.com/mattermost/mattermost-plugin-mscalendar/server/telemetry.rudderWriteKey=$(RUDDER_WRITE_KEY)"

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
	mockgen -destination server/jobs/mock_cluster/mock_cluster.go github.com/mattermost/mattermost-plugin-api/cluster JobPluginAPI
	mockgen -destination server/mscalendar/mock_mscalendar/mock_mscalendar.go github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar MSCalendar
	mockgen -destination server/mscalendar/mock_welcomer/mock_welcomer.go -package mock_welcomer github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar Welcomer
	mockgen -destination server/mscalendar/mock_plugin_api/mock_plugin_api.go -package mock_plugin_api github.com/mattermost/mattermost-plugin-mscalendar/server/mscalendar PluginAPI
	mockgen -destination server/remote/mock_remote/mock_remote.go github.com/mattermost/mattermost-plugin-mscalendar/server/remote Remote
	mockgen -destination server/remote/mock_remote/mock_client.go github.com/mattermost/mattermost-plugin-mscalendar/server/remote Client
	mockgen -destination server/utils/bot/mock_bot/mock_poster.go github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot Poster
	mockgen -destination server/utils/bot/mock_bot/mock_admin.go github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot Admin
	mockgen -destination server/utils/bot/mock_bot/mock_logger.go github.com/mattermost/mattermost-plugin-mscalendar/server/utils/bot Logger
	mockgen -destination server/store/mock_store/mock_store.go github.com/mattermost/mattermost-plugin-mscalendar/server/store Store
endif

clean_mock:
ifneq ($(HAS_SERVER),)
	rm -rf ./server/jobs/mock_cluster
	rm -rf ./server/mscalendar/mock_mscalendar
	rm -rf ./server/mscalendar/mock_welcomer
	rm -rf ./server/mscalendar/mock_plugin_api
	rm -rf ./server/remote/mock_remote
	rm -rf ./server/utils/bot/mock_bot
	rm -rf ./server/store/mock_store
endif
