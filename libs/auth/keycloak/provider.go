package keycloak

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/detect-viz/shared-lib/auth"
	"github.com/detect-viz/shared-lib/models"
)

// Provider Keycloak 認證提供者實現
type Provider struct {
	client      KeycloakClient
	auditLogger *log.Logger
	realm       string
}

// NewProvider 創建 Keycloak 提供者
func NewProvider(config *models.KeycloakConfig) (*Provider, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}

	p := &Provider{
		client:      client,
		auditLogger: log.New(os.Stdout, "AUDIT: ", log.Ldate|log.Ltime),
		realm:       config.Realm,
	}

	return p, nil
}

// UserCredentials 用戶憑證
type UserCredentials struct {
	Username string
	Password string
}

// Authenticate 用戶認證
func (p *Provider) Authenticate(ctx context.Context, credentials interface{}) (*auth.User, error) {
	creds, ok := credentials.(UserCredentials)
	if !ok {
		return nil, fmt.Errorf("invalid credentials type")
	}

	token, err := p.client.GetAccessTokenByUsernamePassword(ctx, creds.Username, creds.Password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	userInfo, err := p.client.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	roles, err := p.client.GetUserRoles(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return &auth.User{
		ID:       userInfo["sub"].(string),
		Username: userInfo["preferred_username"].(string),
		Email:    userInfo["email"].(string),
		Roles:    roles,
		Metadata: userInfo,
	}, nil
}

// ValidateToken 驗證 token
func (p *Provider) ValidateToken(ctx context.Context, token string) (*auth.User, error) {
	// 這裡我們可以使用 Gocloak 的 token 驗證功能
	userInfo, err := p.client.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	roles, err := p.client.GetUserRoles(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return &auth.User{
		ID:       userInfo["sub"].(string),
		Username: userInfo["preferred_username"].(string),
		Email:    userInfo["email"].(string),
		Roles:    roles,
		Metadata: userInfo,
	}, nil
}

// RefreshToken 刷新 token
func (p *Provider) RefreshToken(ctx context.Context, token string) (string, error) {
	// 實現 token 刷新邏輯
	return "", fmt.Errorf("refresh token not implemented")
}

// GetUser 獲取用戶信息
func (p *Provider) GetUser(ctx context.Context, userID string) (*auth.User, error) {
	// 使用已有的 JWT 令牌
	jwt := p.client.GetJWT()
	if jwt == nil {
		return nil, fmt.Errorf("no valid admin token")
	}

	// 使用 gocloak 獲取用戶資訊
	users, err := p.client.GetUsers(ctx, jwt.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range users {
		if user.ID != nil && *user.ID == userID {
			return &auth.User{
				ID:       *user.ID,
				Username: *user.Username,
				Email:    *user.Email,
			}, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// ListUsers 列出用戶
func (p *Provider) ListUsers(ctx context.Context, filter auth.UserFilter) ([]*auth.User, error) {
	// 需要實現列出用戶的邏輯
	return nil, fmt.Errorf("list users not implemented")
}

// GetRole 獲取角色
func (p *Provider) GetRole(ctx context.Context, roleID string) (*auth.Role, error) {
	// 需要實現獲取角色的邏輯
	return nil, fmt.Errorf("get role not implemented")
}

// ListRoles 列出角色
func (p *Provider) ListRoles(ctx context.Context) ([]*auth.Role, error) {
	// 需要實現列出角色的邏輯
	return nil, fmt.Errorf("list roles not implemented")
}

// HasPermission 檢查權限
func (p *Provider) HasPermission(ctx context.Context, userID string, action string, resource string) (bool, error) {
	permissions, err := p.GetPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// 檢查權限邏輯
	for _, perm := range permissions {
		if perm.Action == action && perm.Resource == resource {
			p.auditLogger.Printf("user=%s action=%s resource=%s result=allowed", userID, action, resource)
			return true, nil
		}
	}

	p.auditLogger.Printf("user=%s action=%s resource=%s result=denied", userID, action, resource)
	return false, nil
}

// GetPermissions 獲取用戶權限
func (p *Provider) GetPermissions(ctx context.Context, userID string) ([]*auth.Permission, error) {
	// 使用已有的 JWT 令牌
	jwt := p.client.GetJWT()
	if jwt == nil {
		return nil, fmt.Errorf("no valid admin token")
	}

	// 使用 gocloak 獲取用戶角色
	// 這裡需要根據實際情況實現
	var permissions []*auth.Permission

	// 暫時返回空權限列表
	return permissions, nil
}

// 實現 Authenticator 接口
// Verify 驗證 token
func (p *Provider) Verify(ctx context.Context, token string) (bool, error) {
	_, err := p.ValidateToken(ctx, token)
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetUserInfo 獲取用戶信息
func (p *Provider) GetUserInfo(ctx context.Context, token string) (map[string]interface{}, error) {
	return p.client.GetUserInfo(ctx, token)
}
