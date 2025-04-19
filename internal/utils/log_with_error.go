package utils

import (
	"context"
	"log/slog"
)

func LogWithError(ctx context.Context, msg string, err error) {
	if err != nil {
		slog.ErrorContext(ctx, msg, slog.String("error", err.Error()))
	} else {
		slog.ErrorContext(ctx, msg)
	}
}

func LogInfo(ctx context.Context, msg string, fields ...any) {
	slog.InfoContext(ctx, msg, fields...)
}
