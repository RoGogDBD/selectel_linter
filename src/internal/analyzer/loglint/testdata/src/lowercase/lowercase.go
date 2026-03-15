package lowercase

import (
	"log/slog"

	"go.uber.org/zap"
)

func checkSlog(logger *slog.Logger) {
	slog.Info("Starting server")         // want "log message must start with lowercase letter"
	logger.Error("Failed to connect")    // want "log message must start with lowercase letter"
	slog.Info("starting server on 8080") // ok
	logger.Warn("connection failed")      // ok
}

func checkZap(logger *zap.Logger, sugar *zap.SugaredLogger) {
	logger.Info("Connection failed") // want "log message must start with lowercase letter"
	sugar.Infof("Value is %d", 1)    // want "log message must start with lowercase letter"
	logger.Info("connection failed") // ok
	sugar.Infof("value is %d", 1)    // ok
}
