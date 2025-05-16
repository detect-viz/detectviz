package models

// ServiceConfig 服務配置模型
type ServiceConfig struct {
	Name       string                 `json:"name"`
	Version    string                 `json:"version"`
	ConfigPath string                 `json:"config_path"`
	Settings   map[string]interface{} `json:"settings"`
}

// ConfigFile 配置文件模型
type ConfigFile struct {
	Path     string
	Format   string // yaml/toml/ini
	Content  []byte
	Modified int64
}
