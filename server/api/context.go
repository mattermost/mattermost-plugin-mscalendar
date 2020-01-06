package api

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

var apiContextKey = config.Repository + "/" + fmt.Sprintf("%T", api{})
var notificationHandlerContextKey = config.Repository + "/" + fmt.Sprintf("%T", notificationHandler{})

func Context(ctx context.Context, api API, h NotificationHandler) context.Context {
	ctx = context.WithValue(ctx, apiContextKey, api)
	ctx = context.WithValue(ctx, notificationHandlerContextKey, h)
	return ctx
}

func FromContext(ctx context.Context) API {
	return ctx.Value(apiContextKey).(API)
}

func NotificationHandlerFromContext(ctx context.Context) NotificationHandler {
	return ctx.Value(notificationHandlerContextKey).(NotificationHandler)
}
