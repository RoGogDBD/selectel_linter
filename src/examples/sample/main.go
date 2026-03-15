package sample

import (
	"fmt"
	"log/slog"
)

func run(password string) {
	slog.Info("Starting server")
	slog.Info(fmt.Sprintf("password: %s", password))
	slog.Info("request finished")
}
