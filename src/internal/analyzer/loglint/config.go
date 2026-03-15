package loglint

import (
	"fmt"
	"strings"
)

// ParseConfig преобразует настройки golangci-lint в Config.
func ParseConfig(raw any) (Config, error) {
	cfg := DefaultConfig()
	if raw == nil {
		return cfg, nil
	}

	m, ok := toStringAnyMap(raw)
	if !ok {
		return Config{}, fmt.Errorf("unsupported config type: %T", raw)
	}

	if rulesRaw, ok := getValue(m, "rules"); ok {
		rulesMap, ok := toStringAnyMap(rulesRaw)
		if !ok {
			return Config{}, fmt.Errorf("rules must be map, got %T", rulesRaw)
		}
		if v, ok, err := boolValue(rulesMap, "lowercase"); err != nil {
			return Config{}, err
		} else if ok {
			cfg.Rules.Lowercase = v
		}
		if v, ok, err := boolValue(rulesMap, "english"); err != nil {
			return Config{}, err
		} else if ok {
			cfg.Rules.English = v
		}
		if v, ok, err := boolValue(rulesMap, "special-symbols", "special_symbols"); err != nil {
			return Config{}, err
		} else if ok {
			cfg.Rules.SpecialSymbols = v
		}
		if v, ok, err := boolValue(rulesMap, "sensitive-data", "sensitive_data"); err != nil {
			return Config{}, err
		} else if ok {
			cfg.Rules.SensitiveData = v
		}
	}

	if kwRaw, ok := getValue(m, "sensitive-keywords", "sensitive_keywords"); ok {
		keywords, err := stringSlice(kwRaw)
		if err != nil {
			return Config{}, fmt.Errorf("sensitive-keywords: %w", err)
		}
		cfg.SensitiveKeywords = keywords
	}

	if patRaw, ok := getValue(m, "sensitive-patterns", "sensitive_patterns"); ok {
		patterns, err := stringSlice(patRaw)
		if err != nil {
			return Config{}, fmt.Errorf("sensitive-patterns: %w", err)
		}
		cfg.SensitivePatterns = patterns
	}

	return cfg, nil
}

func getValue(m map[string]any, keys ...string) (any, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v, true
		}
	}
	return nil, false
}

func boolValue(m map[string]any, keys ...string) (bool, bool, error) {
	v, ok := getValue(m, keys...)
	if !ok {
		return false, false, nil
	}
	b, ok := v.(bool)
	if !ok {
		return false, false, fmt.Errorf("%s must be bool, got %T", keys[0], v)
	}
	return b, true, nil
}

func stringSlice(v any) ([]string, error) {
	raw, ok := v.([]any)
	if !ok {
		if s, ok := v.([]string); ok {
			return normalizeStringSlice(s), nil
		}
		return nil, fmt.Errorf("must be list, got %T", v)
	}

	result := make([]string, 0, len(raw))
	for _, item := range raw {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("list item must be string, got %T", item)
		}
		result = append(result, s)
	}
	return normalizeStringSlice(result), nil
}

func normalizeStringSlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		t := strings.TrimSpace(s)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func toStringAnyMap(v any) (map[string]any, bool) {
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	if m, ok := v.(map[any]any); ok {
		out := make(map[string]any, len(m))
		for k, val := range m {
			ks, ok := k.(string)
			if !ok {
				return nil, false
			}
			out[ks] = val
		}
		return out, true
	}
	return nil, false
}
