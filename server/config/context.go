package config

import "fmt"

import "context"

var contextKey = Repository + "/" + fmt.Sprintf("%T", Config{})

func Context(ctx context.Context, conf *Config) context.Context {
	return context.WithValue(ctx, contextKey, conf)
}

func FromContext(ctx context.Context) *Config {
	return ctx.Value(contextKey).(*Config)
}
