package generator

import (
	"fmt"
	"sort"
	"sync"

	"github.com/tylerbutler/commit-config-gen/internal/config"
)

// Generator defines the interface for all config file generators.
// existing is nil when no file exists (generate fresh), otherwise the current
// file content to merge into.
type Generator interface {
	Name() string
	FileName() string
	Generate(cfg *config.Config, existing []byte) ([]byte, error)
}

var (
	mu       sync.Mutex
	registry = map[string]Generator{}
	regOrder []string
)

// Register adds a generator to the registry. Called from init() in each generator file.
func Register(g Generator) {
	mu.Lock()
	defer mu.Unlock()
	name := g.Name()
	if _, ok := registry[name]; ok {
		panic(fmt.Sprintf("generator already registered: %s", name))
	}
	registry[name] = g
	regOrder = append(regOrder, name)
}

// Get returns a generator by name, or an error if not found.
func Get(name string) (Generator, error) {
	mu.Lock()
	defer mu.Unlock()
	g, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown generator: %s", name)
	}
	return g, nil
}

// All returns all registered generators in a stable order.
func All() []Generator {
	mu.Lock()
	defer mu.Unlock()
	sorted := make([]string, len(regOrder))
	copy(sorted, regOrder)
	sort.Strings(sorted)
	gens := make([]Generator, len(sorted))
	for i, name := range sorted {
		gens[i] = registry[name]
	}
	return gens
}

// Names returns all registered generator names in sorted order.
func Names() []string {
	mu.Lock()
	defer mu.Unlock()
	sorted := make([]string, len(regOrder))
	copy(sorted, regOrder)
	sort.Strings(sorted)
	return sorted
}
