package generator

import (
	"fmt"

	"github.com/tylerbutler/commit-config-gen/internal/config"
	"gopkg.in/yaml.v3"
)

func init() {
	Register(&ChangieGenerator{})
}

// ChangieGenerator generates .changie.yaml.
type ChangieGenerator struct{}

func (g *ChangieGenerator) Name() string     { return "changie" }
func (g *ChangieGenerator) FileName() string { return ".changie.yaml" }

func (g *ChangieGenerator) Generate(cfg *config.Config, existing []byte) ([]byte, error) {
	kinds := buildChangieKinds(cfg)

	if existing != nil {
		return mergeChangie(existing, kinds)
	}
	return freshChangie(kinds)
}

type changieKind struct {
	Label string `yaml:"label"`
	Auto  string `yaml:"auto,omitempty"`
}

func buildChangieKinds(cfg *config.Config) []changieKind {
	var kinds []changieKind
	for _, name := range cfg.TypeNames() {
		t := cfg.Types[name]
		if t.ChangelogGroup == nil {
			continue
		}
		kind := changieKind{Label: *t.ChangelogGroup}
		if t.Bump != "" && t.Bump != "none" {
			kind.Auto = t.Bump
		}
		kinds = append(kinds, kind)
	}
	return kinds
}

func freshChangie(kinds []changieKind) ([]byte, error) {
	doc := map[string]any{
		"kinds": kinds,
	}
	data, err := yaml.Marshal(doc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func mergeChangie(existing []byte, kinds []changieKind) ([]byte, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(existing, &root); err != nil {
		return nil, fmt.Errorf("parsing existing .changie.yaml: %w", err)
	}

	// root is a Document node; its first child is the mapping
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return freshChangie(kinds)
	}

	mapping := root.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return freshChangie(kinds)
	}

	// Find and replace the "kinds" key
	found := false
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == "kinds" {
			// Marshal kinds to a YAML node
			var kindsNode yaml.Node
			kindsBytes, err := yaml.Marshal(kinds)
			if err != nil {
				return nil, err
			}
			if err := yaml.Unmarshal(kindsBytes, &kindsNode); err != nil {
				return nil, err
			}
			// kindsNode is Document -> the sequence
			if kindsNode.Kind == yaml.DocumentNode && len(kindsNode.Content) > 0 {
				mapping.Content[i+1] = kindsNode.Content[0]
			}
			found = true
			break
		}
	}

	if !found {
		// Append kinds key-value pair
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "kinds", Tag: "!!str"}
		var kindsNode yaml.Node
		kindsBytes, err := yaml.Marshal(kinds)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(kindsBytes, &kindsNode); err != nil {
			return nil, err
		}
		var valNode *yaml.Node
		if kindsNode.Kind == yaml.DocumentNode && len(kindsNode.Content) > 0 {
			valNode = kindsNode.Content[0]
		} else {
			valNode = &kindsNode
		}
		mapping.Content = append(mapping.Content, keyNode, valNode)
	}

	data, err := yaml.Marshal(&root)
	if err != nil {
		return nil, fmt.Errorf("encoding .changie.yaml: %w", err)
	}
	return data, nil
}
