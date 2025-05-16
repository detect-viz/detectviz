package mock

import (
	"context"
	"detectviz/pkg/security/licensing/interfaces"
)

type MockManager struct {
	features map[string]bool
}

func NewMockManager() interfaces.LicenseManager {
	return &MockManager{
		features: map[string]bool{
			"basic":      true,
			"advanced":   true,
			"enterprise": false,
		},
	}
}

func (m *MockManager) Verify(ctx context.Context, token string) (bool, error) {
	return true, nil
}

func (m *MockManager) GetFeatures(ctx context.Context) ([]string, error) {
	var features []string
	for feature := range m.features {
		features = append(features, feature)
	}
	return features, nil
}

func (m *MockManager) IsFeatureEnabled(ctx context.Context, feature string) bool {
	enabled, exists := m.features[feature]
	return exists && enabled
}
