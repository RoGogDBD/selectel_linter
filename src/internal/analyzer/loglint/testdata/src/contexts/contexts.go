package contexts

import (
	"context"
	"log/slog"
)

func checkContext(logger *slog.Logger, token string) {
	ctx := context.Background()
	logger.InfoContext(ctx, "request ok")                                         // ok
	logger.ErrorContext(ctx, "token: "+token)                                     // want "log message may contain sensitive data"
	slog.Log(ctx, slog.LevelInfo, "auth event", "password", token)                // want "log message may contain sensitive data"
	slog.LogAttrs(ctx, slog.LevelInfo, "auth event", slog.String("token", token)) // want "log message may contain sensitive data"
}
