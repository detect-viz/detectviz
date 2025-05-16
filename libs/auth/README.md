# 認證模組 (Authentication Module)

本模組整合了 `libs/auth` 和 `libs/shared-lib/auth` 的功能，提供統一的認證和授權解決方案。

## 功能概述

### 1. 基本認證
- 用戶認證 (`Authenticate`)
- 令牌驗證 (`ValidateToken`)
- 令牌刷新 (`RefreshToken`)

### 2. 使用者管理
- 獲取使用者資訊 (`GetUser`)
- 列出使用者 (`ListUsers`)

### 3. 角色管理
- 獲取角色 (`GetRole`)
- 列出角色 (`ListRoles`)

### 4. 權限管理
- 權限檢查 (`HasPermission`)
- 獲取權限列表 (`GetPermissions`)

### 5. Keycloak 整合
- SSO 單點登入
- 使用者、角色、資源管理
- 令牌驗證和刷新

## 使用方法

### 初始化認證提供者

```go
import (
    "github.com/detect-viz/shared-lib/auth"
    "github.com/detect-viz/shared-lib/auth/keycloak"
    "github.com/detect-viz/shared-lib/models"
)

func initAuth() (auth.AuthProvider, error) {
    // 配置 Keycloak
    config := &models.KeycloakConfig{
        URL:          "https://keycloak.example.com",
        Realm:        "detectviz",
        ClientID:     "web-client",
        ClientSecret: "client-secret",
    }
    
    // 創建 Keycloak 提供者
    provider, err := keycloak.NewProvider(config)
    if err != nil {
        return nil, err
    }
    
    return provider, nil
}
```

### 使用者認證

```go
// 使用者登入
user, err := provider.Authenticate(ctx, keycloak.UserCredentials{
    Username: "user@example.com",
    Password: "password",
})

// 令牌驗證
user, err := provider.ValidateToken(ctx, token)
```

### 權限檢查

```go
// 檢查使用者是否有特定權限
allowed, err := provider.HasPermission(ctx, userID, "view", "dashboards")
if err != nil {
    // 處理錯誤
}
if allowed {
    // 允許操作
} else {
    // 拒絕操作
}
```

### 中間件使用

```go
import (
    "github.com/detect-viz/shared-lib/auth/middleware"
)

// 創建認證中間件
authMiddleware := middleware.NewAuthMiddleware(provider)

// 在 HTTP 處理器中使用
http.Handle("/api/secured", authMiddleware.Handler(yourHandler))
```

## 架構說明

### 主要接口

- `auth.Authenticator`: 基本認證功能
- `auth.AuthProvider`: 完整認證和權限管理
- `auth.SSOClient`: SSO 客戶端操作

### 實現

- `keycloak.Provider`: Keycloak 認證提供者
- `keycloak.Client`: Keycloak API 客戶端

## 擴展說明

要增加新的認證提供者，需要實現以下接口：

1. `auth.Authenticator` 接口: 基本認證功能
2. `auth.AuthProvider` 接口: 完整認證和權限管理

例如：

```go
type MyAuthProvider struct {
    // 配置和依賴
}

// 實現 Authenticator 接口
func (p *MyAuthProvider) Verify(ctx context.Context, token string) (bool, error) {
    // 實現邏輯
}

func (p *MyAuthProvider) GetUserInfo(ctx context.Context, token string) (map[string]interface{}, error) {
    // 實現邏輯
}

// 實現 AuthProvider 接口
func (p *MyAuthProvider) Authenticate(ctx context.Context, credentials interface{}) (*auth.User, error) {
    // 實現邏輯
}

// ... 實現其他方法
``` 