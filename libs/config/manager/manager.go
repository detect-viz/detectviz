package manager

import (
	"context"
	"detectviz/pkg/server/config/interfaces"
	"detectviz/pkg/server/config/loader"
	"fmt"
	"sync"
)

type Manager struct {
	store   *sync.Map
	loader  loader.Loader
	watcher *watcher.Watcher
	confDir string
}

func NewManager(opts ...Option) *Manager {
	m := &Manager{
		store:   &sync.Map{},
		confDir: "conf",
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Manager) Load(ctx context.Context, path string) error {
	config, err := m.loader.Load(path)
	if err != nil {
		return err
	}

	m.store.Store(path, config)
	return nil
}

func (m *Manager) Get(ctx context.Context, key string) (interface{}, error) {
	if value, ok := m.store.Load(key); ok {
		return value, nil
	}
	return nil, fmt.Errorf("config not found: %s", key)
}

func (m *Manager) Watch(ctx context.Context) (<-chan interfaces.ConfigChange, error) {
	return m.watcher.Watch(ctx)
}
