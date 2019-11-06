package grpc_playground_validator

import (
    `context`
    `testing`
)

func TestContextWithLocale(t *testing.T) {
    want := "en"
    ctx := NewContextWithLocale(context.Background(), want)
    if ctx == nil {
        t.Errorf("want not nil, got nil")
    }

    if got := localeFromContext(ctx); want != got {
        t.Errorf("want %s, got %s", want, got)
    }
    if got := localeFromContext(context.Background()); got != "" {
        t.Errorf("want empty string, got %s", got)
    }
}
