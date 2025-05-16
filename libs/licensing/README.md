# 許可證模組 (Licensing Module)

## 架構說明

### 1. 核心介面
- `interfaces/license.go`: 定義基本許可證管理介面
```go
type LicenseManager interface {
    // Verify 驗證許可證
    Verify(ctx context.Context, token string) (bool, error)
    
    // GetFeatures 獲取許可功能列表
    GetFeatures(ctx context.Context) ([]string, error)
    
    // IsFeatureEnabled 檢查功能是否啟用
    IsFeatureEnabled(ctx context.Context, feature string) bool
}
```

### 2. 數據模型
- `models/license.go`: 許可證模型定義
```go
type License struct {
    ID                  string    `json:"id"`
    CustomerID          string    `json:"customer_id"`
    Type                string    `json:"type"`
    IssuedAt            time.Time `json:"issued_at"`
    ExpiresAt           time.Time `json:"expires_at"`
    IsEnterprise        bool      `json:"is_enterprise"`
    MaxUsers            int       `json:"max_users"`
    HasAdvancedFeatures bool      `json:"has_advanced_features"`
}
```

### 3. 管理器實現
- `manager/manager.go`: 提供許可證管理功能
- 支援功能開關管理
- 支援許可證驗證

### 4. 測試模擬
- `mock/mock.go`: 提供測試用模擬實現

## 使用示例

### 1. 基本使用
```go
import (
    "context"
    "detectviz/pkg/security/licensing/manager"
    "detectviz/pkg/security/licensing/models"
)

func main() {
    // 創建許可證
    license := &models.License{
        ID:                  "lic-001",
        Type:               "enterprise",
        IsEnterprise:       true,
        MaxUsers:           100,
        HasAdvancedFeatures: true,
        ExpiresAt:          time.Now().AddDate(1, 0, 0), // 一年後過期
    }

    // 初始化許可證管理器
    licenseManager := manager.NewManager(license)

    // 檢查功能是否啟用
    if licenseManager.IsFeatureEnabled(ctx, "enterprise") {
        // 使用企業版功能
    }

    // 獲取所有可用功能
    features, _ := licenseManager.GetFeatures(ctx)
    for _, feature := range features {
        // 處理每個功能
    }
}
```

### 2. 測試用例
```go
import (
    "context"
    "testing"
    "detectviz/pkg/security/licensing/mock"
)

func TestLicensing(t *testing.T) {
    // 創建模擬管理器
    mockManager := mock.NewMockManager()

    // 測試功能檢查
    if !mockManager.IsFeatureEnabled(context.Background(), "basic") {
        t.Error("Basic feature should be enabled")
    }

    // 測試企業版功能
    if mockManager.IsFeatureEnabled(context.Background(), "enterprise") {
        t.Error("Enterprise feature should be disabled in mock")
    }
}
```

## 注意事項

1. 許可證驗證
   - 定期檢查許可證有效性
   - 處理過期情況
   - 記錄驗證失敗

2. 功能管理
   - 明確定義功能列表
   - 適當的功能分級
   - 處理未授權訪問

3. 安全性
   - 安全存儲許可證信息
   - 加密敏感數據
   - 防止許可證被篡改

4. 錯誤處理
   - 優雅處理驗證失敗
   - 詳細的錯誤日誌
   - 用戶友好的錯誤信息

## 配置示例

```ini
[licensing]
# 許可證配置
license_path = /etc/detectviz/license.jwt
check_interval = 24h
grace_period = 72h

# 功能開關
enterprise_features = true
max_users = 100
advanced_analytics = true
```
