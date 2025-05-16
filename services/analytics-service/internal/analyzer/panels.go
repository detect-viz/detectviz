package analyzer

import (
	"analytics-service/entities"
	"analytics-service/internal/config"
	"context"
	"fmt"
)

// ProcessNormalPanels 處理一般面板的異常檢測
func (a *AnomalyAnalyzer) ProcessNormalPanels(ctx context.Context, current, history []entities.MetricData) ([]entities.MetricData, error) {
	var results []entities.MetricData

	// 按照規則分組處理
	metricGroups := make(map[string][]entities.MetricData)
	historyGroups := make(map[string][]entities.MetricData)

	// 處理當前和歷史數據
	for _, metric := range current {
		rule := a.findRule(metric.Metric)
		if rule == nil {
			continue
		}
		metricGroups[rule.Model] = append(metricGroups[rule.Model], metric)
	}

	for _, metric := range history {
		rule := a.findRule(metric.Metric)
		if rule == nil {
			continue
		}
		historyGroups[rule.Model] = append(historyGroups[rule.Model], metric)
	}

	// 根據不同模型處理數據
	for model, modelMetrics := range metricGroups {
		var modelResults []entities.MetricData
		var err error

		// 獲取規則配置
		rule := a.findRule(modelMetrics[0].Metric)
		if rule == nil {
			continue
		}

		// 合併配置，規則配置優先於全局配置
		modelConfig := a.mergeConfig(model, rule.Config)

		// 將歷史數據添加到請求中
		requestData := map[string]interface{}{
			"data": map[string]interface{}{
				"current": modelMetrics,
				"history": historyGroups[model],
			},
			"config": modelConfig,
		}

		modelResults, err = a.anomalyClient.CallDetector(ctx, model, requestData)
		if err != nil {
			return nil, fmt.Errorf("process %s model failed: %v", model, err)
		}
		results = append(results, modelResults...)
	}

	return results, nil
}

// processPredictionPanels 處理預測性面板
func (a *AnomalyAnalyzer) processPredictionPanels(ctx context.Context, history []entities.MetricData) ([]entities.MetricData, error) {
	var results []entities.MetricData

	// 遍歷所有預測性面板
	for panelID, panel := range a.cfg.PredictionPanels {
		// 收集該面板需要的指標數據
		var panelMetrics []entities.MetricData
		for _, metric := range panel.Metrics {
			// 從歷史數據中找出對應的數據
			for _, data := range history {
				if data.Metric == metric {
					panelMetrics = append(panelMetrics, data)
				}
			}
		}

		// 使用 Prophet 進行預測
		predictResults, err := a.processProphet(ctx, panelMetrics, panel.ProphetSettings)
		if err != nil {
			return nil, fmt.Errorf("process prediction panel %s failed: %v", panelID, err)
		}

		results = append(results, predictResults...)
	}

	return results, nil
}

// processProphet Prophet 預測處理
func (a *AnomalyAnalyzer) processProphet(ctx context.Context, metrics []entities.MetricData, settings config.ProphetSettings) ([]entities.MetricData, error) {
	requestData := map[string]interface{}{
		"metrics": metrics,
		"params": map[string]interface{}{
			"horizon": settings.Horizon,
			"period":  settings.Period,
		},
	}
	return a.anomalyClient.CallDetector(ctx, "prophet", requestData)
}
