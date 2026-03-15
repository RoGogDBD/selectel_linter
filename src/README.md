# loglint

`loglint` is a custom Go analyzer for log messages in `log/slog` and `go.uber.org/zap`.

Checks:

1. Message starts with lowercase letter.
2. Message contains only English text.
3. Message has no special symbols or emoji.
4. Message does not expose potentially sensitive data.

## Build checker

```bash
go build -o bin/loglint ./cmd/loglint
```

## Run checker

```bash
./bin/loglint ./...
```

## golangci-lint plugin entrypoint

Go plugin entrypoint is in `./plugin` and exports:

```go
func New(conf any) ([]*analysis.Analyzer, error)
```

## Test

```bash
go test ./...
```
