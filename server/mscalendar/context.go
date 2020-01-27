package mscalendar

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost-plugin-mscalendar/server/config"
)

var apiContextKey = config.Repository + "/" + fmt.Sprintf("%T", mscalendar{})
var notificationHandlerContextKey = config.Repository + "/" + fmt.Sprintf("%T", notificationHandler{})

func Context(ctx context.Context, mscalendar MSCalendar, h NotificationHandler) context.Context {
	ctx = context.WithValue(ctx, apiContextKey, mscalendar)
	ctx = context.WithValue(ctx, notificationHandlerContextKey, h)
	return ctx
}

func FromContext(ctx context.Context) MSCalendar {
	return ctx.Value(apiContextKey).(MSCalendar)
}

func NotificationHandlerFromContext(ctx context.Context) NotificationHandler {
	return ctx.Value(notificationHandlerContextKey).(NotificationHandler)
}
