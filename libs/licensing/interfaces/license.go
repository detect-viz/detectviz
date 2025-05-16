package interfaces

import "context"

// LicenseManager 許可證管理介面
type LicenseManager interface {
	// Verify 驗證許可證
	Verify(ctx context.Context, token string) (bool, error)

	// GetFeatures 獲取許可功能列表
	GetFeatures(ctx context.Context) ([]string, error)

	// IsFeatureEnabled 檢查功能是否啟用
	IsFeatureEnabled(ctx context.Context, feature string) bool
}
