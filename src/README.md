# loglint

`loglint` is a production-oriented custom linter for log messages in Go.

Supported loggers:
- `log/slog`
- `go.uber.org/zap` (`Logger` and `SugaredLogger`)

Go version:
- `1.25+`

## What it checks

1. Message starts with lowercase letter.
2. Message contains only English letters.
3. Message has no special symbols or emoji.
4. Message does not leak sensitive data.

Sensitive data detection supports:
- keyword-based matching (`password`, `token`, etc.)
- dynamic concatenation (`"token=" + token`)
- formatted strings (`fmt.Sprintf("password: %s", value)`)
- structured args (`slog.Info("...", "password", value)`, `zap.String("token", value)`)

## Repository Layout

- `internal/analyzer/loglint` — analyzer implementation and tests.
- `cmd/loglint` — standalone multichecker binary.
- `plugin` — Go plugin entrypoint (`-buildmode=plugin`).
- `moduleplugin` — golangci-lint module plugin entrypoint.
- `examples` — ready-to-use configs and demo code.

## Build

```bash
go build -o bin/loglint ./cmd/loglint
go build -buildmode=plugin -o bin/loglint.so ./plugin
```

Or via `Makefile`:

```bash
make build-all
```

## Run as standalone checker

```bash
./bin/loglint ./...
```

## Integrate with golangci-lint (Go Plugin)

1. Build plugin:
```bash
go build -buildmode=plugin -o bin/loglint.so ./plugin
```

2. Copy config from `examples/go-plugin/.golangci.yml`.

3. Run:
```bash
golangci-lint run
```

## Integrate with golangci-lint (Module Plugin)

1. Copy `examples/module-plugin/.custom-gcl.yml`.
2. Build custom binary:
```bash
golangci-lint custom
```

Alternative:
```bash
make custom-gcl
```

3. Copy `examples/module-plugin/.golangci.yml`.
4. Run custom binary:
```bash
./custom-gcl run
```

## Configuration

All settings are provided via `.golangci.yml` under:

```yaml
linters:
  settings:
    custom:
      loglint:
        settings:
          rules:
            lowercase: true
            english: true
            special-symbols: true
            sensitive-data: true
          sensitive-keywords: ["password", "token"]
          sensitive-patterns: ["(?i)ssn\\s*[:=]"]
```

Settings:
- `rules.lowercase` (bool)
- `rules.english` (bool)
- `rules.special-symbols` (bool)
- `rules.sensitive-data` (bool)
- `sensitive-keywords` ([]string): overrides default keywords when set.
- `sensitive-patterns` ([]string): custom regex patterns for sensitive data.

## Auto-fix support

`loglint` provides `SuggestedFix` for lowercase rule when message is a string literal.

Use:
```bash
golangci-lint run --fix
```

## Tests

```bash
go test ./...
go vet ./...
```

Or via `Makefile`:

```bash
make lint
```

Testdata covers:
- each rule separately
- context/logattrs APIs
- custom configuration behavior

## CI

GitHub Actions workflow is provided in:
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`

## Usage Example

Example code with violations:
- `examples/sample/main.go`
