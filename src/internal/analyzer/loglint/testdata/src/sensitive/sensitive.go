package sensitive

import (
	"fmt"
	"log/slog"

	"go.uber.org/zap"
)

func checkSlog(logger *slog.Logger, password, token string) {
	slog.Info("password: " + password) // want "log message may contain sensitive data"
	logger.Error("token=" + token)     // want "log message may contain sensitive data"
	slog.Info("token validated")       // ok
	logger.Info("request completed")   // ok
}

func checkZap(logger *zap.Logger, sugar *zap.SugaredLogger, apiKey, secret string) {
	sugar.Infow("api_key=" + apiKey)                       // want "log message may contain sensitive data"
	sugar.Debug("secret=" + secret)                        // want "log message may contain sensitive data"
	logger.Info("request completed")                       // ok
	sugar.Info("request id one")                           // ok
	slog.Info("user login", "password", secret)            // want "log message may contain sensitive data"
	logger.Info("user login", zap.String("token", secret)) // want "log message may contain sensitive data"
	slog.Info(fmt.Sprintf("password: %s", secret))         // want "log message may contain sensitive data"
}
