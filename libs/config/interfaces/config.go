package interfaces

import "context"

// ConfigManager 配置管理介面
type ConfigManager interface {
	// Load 加載配置
	Load(ctx context.Context, path string) error

	// Get 獲取配置值
	Get(ctx context.Context, key string) (interface{}, error)

	// Watch 監聽配置變更
	Watch(ctx context.Context) (<-chan ConfigChange, error)
}

// ConfigChange 配置變更事件
type ConfigChange struct {
	Type     string      // 變更類型: file/service
	Service  string      // 服務名稱
	Path     string      // 配置路徑
	OldValue interface{} // 舊值
	NewValue interface{} // 新值
}
