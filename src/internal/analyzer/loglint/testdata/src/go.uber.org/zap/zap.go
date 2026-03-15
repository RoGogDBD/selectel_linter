package zap

type Field struct{}

type Logger struct{}

type SugaredLogger struct{}

func String(key, val string) Field { return Field{} }

func (l *Logger) Debug(msg string, fields ...Field)  {}
func (l *Logger) Info(msg string, fields ...Field)   {}
func (l *Logger) Warn(msg string, fields ...Field)   {}
func (l *Logger) Error(msg string, fields ...Field)  {}
func (l *Logger) DPanic(msg string, fields ...Field) {}
func (l *Logger) Panic(msg string, fields ...Field)  {}
func (l *Logger) Fatal(msg string, fields ...Field)  {}

func (l *SugaredLogger) Debug(args ...any)                       {}
func (l *SugaredLogger) Info(args ...any)                        {}
func (l *SugaredLogger) Warn(args ...any)                        {}
func (l *SugaredLogger) Error(args ...any)                       {}
func (l *SugaredLogger) Debugf(template string, args ...any)     {}
func (l *SugaredLogger) Infof(template string, args ...any)      {}
func (l *SugaredLogger) Warnf(template string, args ...any)      {}
func (l *SugaredLogger) Errorf(template string, args ...any)     {}
func (l *SugaredLogger) Debugw(msg string, keysAndValues ...any) {}
func (l *SugaredLogger) Infow(msg string, keysAndValues ...any)  {}
func (l *SugaredLogger) Warnw(msg string, keysAndValues ...any)  {}
func (l *SugaredLogger) Errorw(msg string, keysAndValues ...any) {}
