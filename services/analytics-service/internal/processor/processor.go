package processor

import (
	"analytics-service/entities"
	"analytics-service/internal/config"
	"fmt"
	"strings"
)

// ModelProcessor 處理模型數據
type ModelProcessor struct {
	cfg *config.Config
}

// NewModelProcessor 創建新的處理器
func NewModelProcessor(cfg *config.Config) *ModelProcessor {
	return &ModelProcessor{
		cfg: cfg,
	}
}

// ProcessMetrics 處理異常數據並組裝 prompt
func (p *ModelProcessor) ProcessMetrics(profileID string, anomalies []entities.MetricData) (map[string]interface{}, error) {
	// 查找配置
	profile, ok := p.cfg.Profiles[profileID]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", profileID)
	}

	// 根據 panel 整理異常數據
	panelResults := make(map[string]map[string]interface{})

	// 建立指標到異常數據的映射
	metricAnomalies := make(map[string]entities.MetricData)
	for _, anomaly := range anomalies {
		metricAnomalies[anomaly.Metric] = anomaly
	}

	// 遍歷配置中的面板
	for _, panelID := range profile.Panels {
		panel, ok := p.cfg.Panels[panelID]
		if !ok {
			continue
		}

		var panelAnomalies []entities.MetricData
		// 檢查面板中的每個規則
		for _, ruleName := range panel.Rules {
			if rule := p.findRule(ruleName); rule != nil {
				if anomaly, exists := metricAnomalies[rule.MetricName]; exists {
					panelAnomalies = append(panelAnomalies, anomaly)
				}
			}
		}

		// 如果面板有異常數據，添加到結果中
		if len(panelAnomalies) > 0 {
			panelResults[panelID] = map[string]interface{}{
				"title":       panel.Title,
				"description": panel.Description,
				"anomalies":   panelAnomalies,
				"prompt":      panel.PromptTemplate,
			}
		}
	}

	// 組裝結果
	result := map[string]interface{}{
		"profile_id":     profileID,
		"title":          profile.Title,
		"description":    profile.Description,
		"panels":         panelResults,
		"prompt":         profile.Prompt,
		"prompt_details": p.buildPrompt(panelResults),
	}

	return result, nil
}

// buildPrompt 根據分組構建 prompt
func (p *ModelProcessor) buildPrompt(panels map[string]map[string]interface{}) string {
	var prompt strings.Builder

	for _, panel := range panels {
		// 添加面板標題和描述
		prompt.WriteString("\n")
		prompt.WriteString(fmt.Sprintf("【%s】\n", panel["title"].(string)))
		prompt.WriteString(fmt.Sprintf("%s\n", panel["description"].(string)))

		// 添加面板提示詞
		prompt.WriteString(panel["prompt"].(string))
		prompt.WriteString("\n\n異常指標：\n")

		// 添加異常數據描述
		for _, anomaly := range panel["anomalies"].([]entities.MetricData) {
			rule := p.findRule(anomaly.Metric)
			if rule == nil {
				continue
			}

			// 根據嚴重程度選擇模板
			var template string
			switch anomaly.Severity {
			case "critical":
				template = rule.Templates.Desc["critical"]
			case "warning":
				template = rule.Templates.Desc["warning"]
			default:
				if rule.Model == "isolation_forest" {
					template = rule.Templates.Desc["anomaly"]
				} else {
					template = rule.Templates.Desc["normal"]
				}
			}

			// 替換模板變量
			description := strings.ReplaceAll(template, "{value}", fmt.Sprintf("%.2f", anomaly.Value))
			if anomaly.Min != 0 || anomaly.Max != 0 {
				description = strings.ReplaceAll(description, "{min}", fmt.Sprintf("%.2f", anomaly.Min))
				description = strings.ReplaceAll(description, "{max}", fmt.Sprintf("%.2f", anomaly.Max))
			}
			description = strings.ReplaceAll(description, "{unit}", rule.Unit)

			prompt.WriteString("- ")
			prompt.WriteString(description)
			prompt.WriteString("\n")
		}
		prompt.WriteString("\n")
	}

	return prompt.String()
}

// findRule 查找規則
func (p *ModelProcessor) findRule(ruleName string) *config.Rule {
	rule, ok := p.cfg.Rules[ruleName]
	if !ok {
		return nil
	}
	return &rule
}
