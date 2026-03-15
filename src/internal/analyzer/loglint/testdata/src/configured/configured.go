package configured

import "log/slog"

func checkConfigured(ssn, token, password string) {
	slog.Info("Starting server!!!") // ok (lowercase + symbols checks disabled)
	slog.Info("ssn: " + ssn)        // want "log message may contain sensitive data"
	slog.Info("token=" + token)     // want "log message may contain sensitive data"
	slog.Info("password: " + password)
}
