package a

import (
	"log/slog"

	"go.uber.org/zap"
)

func checkSlog(logger *slog.Logger, password string, token string) {
	slog.Info("Starting server")              // want "log message must start with lowercase letter"
	slog.Info("запуск сервера")               // want "log message must contain only english text"
	slog.Info("server started!!!")            // want "log message must not contain special symbols or emoji"
	slog.Info("password: " + password)        // want "log message may contain sensitive data"
	logger.Error("token: " + token)           // want "log message may contain sensitive data"
	slog.Info("token validated")              // ok
	slog.Info("starting server on port 8080") // ok
	slog.Warn("something went wrong")         // ok
}

func checkZap(logger *zap.Logger, sugar *zap.SugaredLogger, apiKey string) {
	logger.Info("Connection failed")    // want "log message must start with lowercase letter"
	logger.Info("connection failed!!!") // want "log message must not contain special symbols or emoji"
	sugar.Infow("api_key=" + apiKey)    // want "log message may contain sensitive data"
	sugar.Infof("ошибка подключения")   // want "log message must contain only english text"
	logger.Warn("connection failed")    // ok
}
