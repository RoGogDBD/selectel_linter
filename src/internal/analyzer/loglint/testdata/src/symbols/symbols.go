package symbols

import (
	"log/slog"

	"go.uber.org/zap"
)

func checkSlog(logger *slog.Logger) {
	slog.Info("server started!!!")    // want "log message must not contain special symbols or emoji"
	logger.Warn("request failed...")  // want "log message must not contain special symbols or emoji"
	slog.Info("server started")       // ok
	logger.Warn("request failed 503") // ok
}

func checkZap(logger *zap.Logger, sugar *zap.SugaredLogger) {
	logger.Info("connection failed!") // want "log message must not contain special symbols or emoji"
	sugar.Info("value one 🚀")         // want "log message must not contain special symbols or emoji"
	logger.Info("connection failed")  // ok
	sugar.Info("value one")           // ok
	sugar.Infof("value %d", 10)       // ok
}
