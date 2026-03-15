package sensitive

import (
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
	sugar.Infow("api_key=" + apiKey)   // want "log message may contain sensitive data"
	sugar.Debugf("secret: %s", secret) // want "log message may contain sensitive data"
	logger.Info("request completed")   // ok
	sugar.Infof("request id %d", 1)    // ok
}
