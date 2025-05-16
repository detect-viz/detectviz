package plugins

import (
	"context"

	"detectviz/pkg/storage"
)

type PluginManager struct {
	store    storage.Store
	registry plugin.Registry
	loader   plugin.Loader
}

func (pm *PluginManager) Load(ctx context.Context) error {
	// 插件加載邏輯
}

func (pm *PluginManager) Start(ctx context.Context) error {
	// 插件啟動邏輯
}

type PluginManager interface {
	// 插件生命週期
	Load(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// 插件管理
	Register(descriptor PluginDescriptor) error
	Get(pluginID string) (Plugin, bool)
	List() []Plugin
}
