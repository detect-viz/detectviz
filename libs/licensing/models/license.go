package models

import "time"

// License 許可證模型
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
