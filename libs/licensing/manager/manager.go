package manager

import (
	"context"
	"detectviz/pkg/security/licensing/interfaces"
	"detectviz/pkg/security/licensing/models"
)

type Manager struct {
	features map[string]bool
	license  *models.License
}

func NewManager(license *models.License) interfaces.LicenseManager {
	m := &Manager{
		features: make(map[string]bool),
		license:  license,
	}
	m.initFeatures()
	return m
}

func (m *Manager) Verify(ctx context.Context, token string) (bool, error) {
	// 實現許可證驗證邏輯
	return true, nil
}

func (m *Manager) GetFeatures(ctx context.Context) ([]string, error) {
	var features []string
	for feature := range m.features {
		features = append(features, feature)
	}
	return features, nil
}

func (m *Manager) IsFeatureEnabled(ctx context.Context, feature string) bool {
	enabled, exists := m.features[feature]
	return exists && enabled
}

func (m *Manager) initFeatures() {
	// 初始化功能列表
	m.features = map[string]bool{
		"basic":      true,
		"advanced":   m.license.HasAdvancedFeatures,
		"enterprise": m.license.IsEnterprise,
	}
}
