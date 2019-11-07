package grpc_playground_validator

import (
	"context"
)

var localeKey struct{}

// NewContextWithLocale creates new context with locale
func NewContextWithLocale(parent context.Context, value string) context.Context {
	return context.WithValue(parent, localeKey, value)
}

func localeFromContext(ctx context.Context) string {
	val := ctx.Value(localeKey)
	if locale, ok := val.(string); ok {
		return locale
	}
	return ""
}
