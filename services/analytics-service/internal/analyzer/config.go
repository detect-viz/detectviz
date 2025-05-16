package analyzer

import "analytics-service/internal/config"

// findRule 查找規則
func (a *AnomalyAnalyzer) findRule(metricName string) *config.Rule {
	for _, rule := range a.cfg.Rules {
		if rule.MetricName == metricName {
			return &rule
		}
	}
	return nil
}

// mergeConfig 合併配置，規則配置優先於全局配置
func (a *AnomalyAnalyzer) mergeConfig(model string, ruleConfig map[string]interface{}) map[string]interface{} {
	var defaultConfig map[string]interface{}

	// 獲取全局默認配置
	switch model {
	case "absolute_threshold":
		defaultConfig = a.cfg.AnomalyService.Models.AbsoluteThreshold
	case "percentage_threshold":
		defaultConfig = a.cfg.AnomalyService.Models.PercentageThreshold
	case "moving_average":
		defaultConfig = a.cfg.AnomalyService.Models.MovingAverage
	case "isolation_forest":
		defaultConfig = a.cfg.AnomalyService.Models.IsolationForest
	default:
		return ruleConfig
	}

	// 創建配置副本
	mergedConfig := make(map[string]interface{})

	// 複製默認配置
	for k, v := range defaultConfig {
		mergedConfig[k] = v
	}

	// 使用規則配置覆蓋默認配置，但忽略空值
	for k, v := range ruleConfig {
		// 忽略空字符串和 nil 值
		if v == "" || v == nil {
			continue
		}
		mergedConfig[k] = v
	}

	return mergedConfig
}
