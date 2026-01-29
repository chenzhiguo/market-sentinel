package llm

import (
	"fmt"
	"sync"
)

// Factory defines a function that creates a Provider
type Factory func(config map[string]string) (Provider, error)

var (
	providersMu sync.RWMutex
	providers   = make(map[string]Factory)
)

// Register registers a provider factory
func Register(name string, factory Factory) {
	providersMu.Lock()
	defer providersMu.Unlock()
	if factory == nil {
		panic("llm: Register factory is nil")
	}
	if _, dup := providers[name]; dup {
		panic("llm: Register called twice for provider " + name)
	}
	providers[name] = factory
}

// NewProvider creates a provider instance by name
func NewProvider(name string, config map[string]string) (Provider, error) {
	providersMu.RLock()
	factory, ok := providers[name]
	providersMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("llm: unknown provider %q (forgot to import?)", name)
	}

	return factory(config)
}
