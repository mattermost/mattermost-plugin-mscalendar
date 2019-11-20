package api

import (
	"context"
	"fmt"

	"github.com/mattermost/mattermost-plugin-msoffice/server/config"
)

var contextKey = config.Repository + "/" + fmt.Sprintf("%T", api{})

func Context(ctx context.Context, api API) context.Context {
	return context.WithValue(ctx, contextKey, api)
}

func FromContext(ctx context.Context) API {
	return ctx.Value(contextKey).(API)
}
