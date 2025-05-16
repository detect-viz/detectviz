package analyzer

import (
	"analytics-service/entities"
	"analytics-service/internal/config"
	"context"
	"fmt"
	"time"
)

// AnomalyAnalyzer 異常分析器
type AnomalyAnalyzer struct {
	cfg           *config.Config
	anomalyClient AnomalyServiceClient
}

// NewAnomalyAnalyzer 創建新的分析器
func NewAnomalyAnalyzer(cfg *config.Config) *AnomalyAnalyzer {
	return &AnomalyAnalyzer{
		cfg:           cfg,
		anomalyClient: NewAnomalyClient(cfg.AnomalyService),
	}
}

// Analyze 進行異常檢測
func (a *AnomalyAnalyzer) Analyze(ctx context.Context, current, history []entities.MetricData) (*entities.DetectionResult, error) {
	// 1. 一般面板的異常檢測
	normalResults, err := a.ProcessNormalPanels(ctx, current, history)
	if err != nil {
		return nil, err
	}

	// 2. 預測面板的趨勢分析
	predictionResults, err := a.processPredictionPanels(ctx, history)
	if err != nil {
		return nil, err
	}

	// 3. 返回檢測結果
	return &entities.DetectionResult{
		Anomalies:   normalResults,
		Predictions: predictionResults,
		Timestamp:   time.Now().Unix(),
	}, nil
}

// GetRequiredMetrics 獲取指定 profile 需要的 metrics 列表
func (a *AnomalyAnalyzer) GetRequiredMetrics(profileID string) (*entities.MetricsResponse, error) {
	profile, ok := a.cfg.Profiles[profileID]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", profileID)
	}

	metricsMap := make(map[string]string) // metric_name -> unit
	needHistory := false

	// 處理一般面板
	for _, panelID := range profile.Panels {
		panel, ok := a.cfg.Panels[panelID]
		if !ok {
			continue
		}
		for _, ruleID := range panel.Rules {
			rule, ok := a.cfg.Rules[ruleID]
			if !ok {
				continue
			}
			metricsMap[rule.MetricName] = rule.Unit

			// 檢查是否需要歷史數據
			switch rule.Model {
			case "absolute_threshold":
				// 只需要當前數據
			case "prophet":
				needHistory = true
			case "percentage_threshold", "moving_average", "isolation_forest":
				needHistory = true
			}
		}
	}

	// 處理預測面板
	for _, panelID := range profile.PredictionPanels {
		panel, ok := a.cfg.PredictionPanels[panelID]
		if !ok {
			continue
		}
		for _, metricName := range panel.Metrics {
			if rule, ok := a.findRuleByMetric(metricName); ok {
				metricsMap[rule.MetricName] = rule.Unit
				needHistory = true
			}
		}
	}

	metrics := make([]entities.MetricInfo, 0, len(metricsMap))
	for name, unit := range metricsMap {
		metrics = append(metrics, entities.MetricInfo{
			MetricName: name,
			Unit:       unit,
		})
	}

	return &entities.MetricsResponse{
		Metrics: metrics,
		Current: true,
		History: needHistory,
	}, nil
}

// findRuleByMetric 通過 metric 名稱查找規則
func (a *AnomalyAnalyzer) findRuleByMetric(metricName string) (*config.Rule, bool) {
	for _, rule := range a.cfg.Rules {
		if rule.MetricName == metricName {
			return &rule, true
		}
	}
	return nil, false
}
