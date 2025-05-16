package auth

import (
	"context"
)

// SSOClient 定義 SSO 客戶端介面
type SSOClient interface {
	GetAdminAcecessToken(ctx context.Context) (string, error)
}

// Authenticator 基本認證介面
type Authenticator interface {
	// Verify 驗證 token
	Verify(ctx context.Context, token string) (bool, error)

	// GetUserInfo 獲取用戶信息
	GetUserInfo(ctx context.Context, token string) (map[string]interface{}, error)
}

// AuthProvider 認證提供者介面
type AuthProvider interface {
	// 基本認證
	Authenticate(ctx context.Context, credentials interface{}) (*User, error)
	ValidateToken(ctx context.Context, token string) (*User, error)
	RefreshToken(ctx context.Context, token string) (string, error)

	// 用戶管理
	GetUser(ctx context.Context, userID string) (*User, error)
	ListUsers(ctx context.Context, filter UserFilter) ([]*User, error)

	// 角色管理
	GetRole(ctx context.Context, roleID string) (*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)

	// 權限管理
	HasPermission(ctx context.Context, userID string, action string, resource string) (bool, error)
	GetPermissions(ctx context.Context, userID string) ([]*Permission, error)
}

// User 用戶信息
type User struct {
	ID          string
	Username    string
	Email       string
	Roles       []string
	Permissions []*Permission
	Metadata    map[string]interface{}
}

// Role 角色定義
type Role struct {
	ID          string
	Name        string
	Description string
	Permissions []*Permission
}

// Permission 權限定義
type Permission struct {
	Action   string // 操作: view, edit, admin
	Resource string // 資源: dashboards, users, settings
	Scope    string // 範圍: global, org, folder, dashboard
}

// UserFilter 用戶過濾條件
type UserFilter struct {
	Role     string
	OrgID    string
	Username string
}
