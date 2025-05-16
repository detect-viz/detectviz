# 配置服務 (Configuration Service)

## 架構說明

### 1. 核心介面
- `interfaces/config.go`: 定義基本配置管理介面
```go
type ConfigManager interface {
    // Load 加載配置
    Load(ctx context.Context, path string) error
    
    // Get 獲取配置值
    Get(ctx context.Context, key string) (interface{}, error)
    
    // Watch 監聽配置變更
    Watch(ctx context.Context) (<-chan ConfigChange, error)
}
```

### 2. 數據模型
- `models/config.go`: 配置相關模型
```go
type ServiceConfig struct {
    Name       string                 `json:"name"`
    Version    string                 `json:"version"`
    ConfigPath string                 `json:"config_path"`
    Settings   map[string]interface{} `json:"settings"`
}
```

### 3. 配置加載器
- `loader/loader.go`: 支援多種格式的配置加載
  - YAML 格式
  - TOML 格式
  - INI 格式

### 4. 配置管理器
- `manager/manager.go`: 提供配置管理功能
  - 配置加載
  - 配置緩存
  - 變更監聽

## 使用示例

### 1. 基本使用
```go
import (
    "context"
    "detectviz/pkg/server/config/manager"
    "detectviz/pkg/server/config/loader"
)

func main() {
    // 初始化配置管理器
    configManager := manager.NewManager(
        manager.WithLoader(loader.NewYAMLLoader()),
        manager.WithConfDir("conf"),
    )
    
    // 加載配置
    err := configManager.Load(ctx, "conf/default.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // 獲取配置
    dbConfig, err := configManager.Get(ctx, "database.mysql")
    if err != nil {
        log.Error(err)
    }
}
```

### 2. 監聽配置變更
```go
// 監聽配置變更
changes, err := configManager.Watch(ctx)
if err != nil {
    log.Fatal(err)
}

go func() {
    for change := range changes {
        switch change.Type {
        case "file":
            log.Info("配置文件變更",
                "path", change.Path,
                "old", change.OldValue,
                "new", change.NewValue)
        case "service":
            log.Info("服務配置變更",
                "service", change.Service,
                "path", change.Path)
        }
    }
}()
```

### 3. 使用不同格式
```go
// YAML 配置
yamlLoader := loader.NewYAMLLoader()
config, err := yamlLoader.Load("conf/config.yaml")

// TOML 配置
tomlLoader := loader.NewTOMLLoader()
config, err := tomlLoader.Load("conf/config.toml")

// INI 配置
iniLoader := loader.NewINILoader()
config, err := iniLoader.Load("conf/config.ini")
```

## 配置文件示例

### YAML 格式
```yaml
server:
  host: localhost
  port: 8080
  
database:
  mysql:
    host: localhost
    port: 3306
    user: root
    password: secret
```

### TOML 格式
```toml
[server]
host = "localhost"
port = 8080

[database.mysql]
host = "localhost"
port = 3306
user = "root"
password = "secret"
```

### INI 格式
```ini
[server]
host = localhost
port = 8080

[database.mysql]
host = localhost
port = 3306
user = root
password = secret
```

## 注意事項

1. 配置加載
   - 支援多種格式
   - 統一的配置結構
   - 配置驗證

2. 配置管理
   - 配置緩存
   - 熱重載
   - 變更通知

3. 安全性
   - 敏感配置加密
   - 權限控制
   - 審計日誌

4. 錯誤處理
   - 配置格式錯誤
   - 文件訪問錯誤
   - 配置缺失
