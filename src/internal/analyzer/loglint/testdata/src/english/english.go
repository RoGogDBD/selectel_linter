package english

import (
	"log/slog"

	"go.uber.org/zap"
)

func checkSlog(logger *slog.Logger) {
	slog.Info("запуск сервера")       // want "log message must contain only english text"
	logger.Warn("ошибка подключения") // want "log message must contain only english text"
	slog.Info("starting server")      // ok
	logger.Error("connection failed") // ok
}

func checkZap(logger *zap.Logger, sugar *zap.SugaredLogger) {
	logger.Info("ошибка")          // want "log message must contain only english text"
	sugar.Info("привет мир")       // want "log message must contain only english text"
	logger.Info("request failed")  // ok
	sugar.Info("request complete") // ok
}
