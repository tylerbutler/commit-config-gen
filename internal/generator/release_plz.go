package generator

import (
	"bytes"
	"fmt"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/tylerbutler/commit-config-gen/internal/config"
)

func init() {
	Register(&ReleasePlzGenerator{})
}

// ReleasePlzGenerator generates release-plz.toml.
type ReleasePlzGenerator struct{}

func (g *ReleasePlzGenerator) Name() string     { return "release-plz" }
func (g *ReleasePlzGenerator) FileName() string { return "release-plz.toml" }

func (g *ReleasePlzGenerator) Generate(cfg *config.Config, existing []byte) ([]byte, error) {
	parsers := buildReleasePlzParsers(cfg)

	if existing != nil {
		return mergeReleasePlz(existing, parsers)
	}
	return freshReleasePlz(parsers)
}

type releasePlzParser struct {
	Message string `toml:"message"`
	Group   string `toml:"group,omitempty"`
	Skip    bool   `toml:"skip,omitempty"`
}

func buildReleasePlzParsers(cfg *config.Config) []releasePlzParser {
	var parsers []releasePlzParser

	for _, scope := range cfg.ExcludedScopes {
		parsers = append(parsers, releasePlzParser{
			Message: fmt.Sprintf(`^[a-z]+\(%s\)`, scope),
			Skip:    true,
		})
	}

	for _, name := range cfg.TypeNames() {
		t := cfg.Types[name]
		if t.ChangelogGroup != nil {
			parsers = append(parsers, releasePlzParser{
				Message: "^" + name,
				Group:   *t.ChangelogGroup,
			})
		}
	}

	parsers = append(parsers, releasePlzParser{
		Message: ".*",
		Skip:    true,
	})

	return parsers
}

func freshReleasePlz(parsers []releasePlzParser) ([]byte, error) {
	doc := map[string]any{
		"changelog": map[string]any{
			"commit_parsers": tomlParsers(parsers),
		},
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func mergeReleasePlz(existing []byte, parsers []releasePlzParser) ([]byte, error) {
	var doc map[string]any
	if err := toml.Unmarshal(existing, &doc); err != nil {
		return nil, fmt.Errorf("parsing existing release-plz.toml: %w", err)
	}

	changelog, ok := doc["changelog"].(map[string]any)
	if !ok {
		changelog = map[string]any{}
		doc["changelog"] = changelog
	}

	changelog["commit_parsers"] = tomlParsers(parsers)

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("encoding release-plz.toml: %w", err)
	}
	return buf.Bytes(), nil
}

func tomlParsers(parsers []releasePlzParser) []any {
	var result []any
	for _, p := range parsers {
		m := map[string]any{"message": p.Message}
		if p.Group != "" {
			m["group"] = p.Group
		}
		if p.Skip {
			m["skip"] = p.Skip
		}
		result = append(result, m)
	}
	return result
}
