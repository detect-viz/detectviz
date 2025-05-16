package loader

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/go-ini/ini"
	"gopkg.in/yaml.v3"
)

// Loader 配置加載器介面
type Loader interface {
	Load(path string) (interface{}, error)
	Format() string
}

// YAML 加載器
type YAMLLoader struct{}

func NewYAMLLoader() Loader {
	return &YAMLLoader{}
}

func (l *YAMLLoader) Load(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func (l *YAMLLoader) Format() string {
	return "yaml"
}

// TOML 加載器
type TOMLLoader struct{}

func NewTOMLLoader() Loader {
	return &TOMLLoader{}
}

func (l *TOMLLoader) Load(path string) (interface{}, error) {
	var config interface{}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func (l *TOMLLoader) Format() string {
	return "toml"
}

// INI 加載器
type INILoader struct{}

func NewINILoader() Loader {
	return &INILoader{}
}

func (l *INILoader) Load(path string) (interface{}, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	// 轉換為 map
	config := make(map[string]interface{})
	for _, section := range cfg.Sections() {
		sectionMap := make(map[string]string)
		for _, key := range section.Keys() {
			sectionMap[key.Name()] = key.Value()
		}
		config[section.Name()] = sectionMap
	}

	return config, nil
}

func (l *INILoader) Format() string {
	return "ini"
}

// 工廠函數
func NewLoader(format string) (Loader, error) {
	switch format {
	case "yaml", "yml":
		return NewYAMLLoader(), nil
	case "toml":
		return NewTOMLLoader(), nil
	case "ini":
		return NewINILoader(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// 加載配置
func Load(path string) (interface{}, error) {
	config, err := AutoLoad(path)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	return config, nil
}
