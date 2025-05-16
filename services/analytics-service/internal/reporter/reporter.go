package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"analytics-service/entities"
	"analytics-service/internal/config"
	"context"
)

// ReportGenerator 報告生成器
type ReportGenerator struct {
	cfg *config.Config
}

// NewReportGenerator 創建新的報告生成器
func NewReportGenerator(cfg *config.Config) *ReportGenerator {
	return &ReportGenerator{
		cfg: cfg,
	}
}

// GenerateReport 生成報告
func (r *ReportGenerator) GenerateReport(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	// 準備請求
	client := &http.Client{
		Timeout: time.Duration(r.cfg.LLMAPI.Timeout) * time.Second,
	}

	// 獲取 profile ID 和配置
	profileID := data["profile_id"].(string)
	profile, ok := r.cfg.Profiles[profileID]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", profileID)
	}

	// 構建請求體
	requestBody := map[string]interface{}{
		"model":       r.cfg.LLMAPI.API.Model,
		"temperature": r.cfg.LLMAPI.Temperature,
		"max_tokens":  r.cfg.LLMAPI.MaxTokens,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": profile.Prompt,
			},
			{
				"role":    "user",
				"content": data["prompt"].(string),
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	// 重試機制
	var response struct {
		Report string `json:"report"`
	}
	var lastErr error

	for attempt := 0; attempt < r.cfg.LLMAPI.Retry.MaxAttempts; attempt++ {
		response, lastErr = r.callLLMAPI(client, jsonData)
		if lastErr == nil {
			break
		}
		time.Sleep(time.Duration(r.cfg.LLMAPI.Retry.InitialInterval) * time.Second)
	}
	if lastErr != nil {
		return nil, lastErr
	}

	// 構建面板報告
	var panelReports []entities.PanelReport
	panelsData := data["panels"].(map[string]map[string]interface{})

	for panelID, panel := range panelsData {
		metrics := panel["anomalies"].([]entities.MetricData)
		var metricReports []entities.MetricReport

		for _, metric := range metrics {
			rule := r.findRule(metric.Metric)
			if rule == nil {
				continue
			}

			// 構建指標報告
			report := entities.MetricReport{
				Timestamp:   metric.Timestamp,
				Value:       metric.Value,
				Severity:    getSeverity(metric.Value, rule),
				Anomaly:     true,
				Description: getDescription(metric, rule),
			}

			// 如果是移動平均模型，添加最大最小值
			if rule.Model == "moving_average" {
				report.Min = metric.Min
				report.Max = metric.Max
			}

			metricReports = append(metricReports, report)
		}

		panelReport := entities.PanelReport{
			ID:          panelID,
			Title:       panel["title"].(string),
			Description: panel["description"].(string),
			Metrics:     metricReports,
		}
		panelReports = append(panelReports, panelReport)
	}

	// 構建最終報告
	finalReport := entities.Report{
		Timestamp:   time.Now().Unix(),
		ReportID:    data["report_id"].(string),
		Title:       data["title"].(string),
		Description: data["description"].(string),
		Content:     response.Report,
		Panels:      panelReports,
	}

	data["report"] = finalReport
	return data, nil
}

func (r *ReportGenerator) callLLMAPI(client *http.Client, jsonData []byte) (struct {
	Report string `json:"report"`
}, error) {
	var response struct {
		Report string `json:"report"`
	}

	resp, err := client.Post(
		fmt.Sprintf("%s/generate", r.cfg.LLMAPI.API.URL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return response, fmt.Errorf("call llm api failed: %v", err)
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("llm api returned status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("decode response failed: %v", err)
	}

	return response, nil
}

// 輔助函數
func (r *ReportGenerator) findRule(metricName string) *config.Rule {
	for _, rule := range r.cfg.Rules {
		if rule.MetricName == metricName {
			ruleCopy := rule // 創建副本以避免迴圈變數捕獲
			return &ruleCopy
		}
	}
	return nil
}

func getSeverity(value float64, rule *config.Rule) string {
	// 對於異常檢測模型，使用二元分類
	switch rule.Model {
	case "isolation_forest", "prophet":
		if value > 0 {
			return "anomaly"
		}
		return "normal"
	}

	// 從規則配置中獲取閾值
	thresholds, ok := rule.Config["thresholds"].(map[string]interface{})
	if !ok {
		return "normal"
	}

	// 獲取臨界值和警告值
	critical, hasCritical := thresholds["critical"].(float64)
	warning, hasWarning := thresholds["warning"].(float64)

	if hasCritical && value > critical {
		return "critical"
	}
	if hasWarning && value > warning {
		return "warning"
	}
	return "normal"
}

func getDescription(metric entities.MetricData, rule *config.Rule) string {
	var template string
	switch metric.Severity {
	case "critical":
		template = rule.Templates.Desc["critical"]
	case "warning":
		template = rule.Templates.Desc["warning"]
	case "anomaly":
		template = rule.Templates.Desc["anomaly"]
	default:
		template = rule.Templates.Desc["normal"]
	}

	description := strings.ReplaceAll(template, "{value}", fmt.Sprintf("%.2f", metric.Value))
	if metric.Min != 0 || metric.Max != 0 {
		description = strings.ReplaceAll(description, "{min}", fmt.Sprintf("%.2f", metric.Min))
		description = strings.ReplaceAll(description, "{max}", fmt.Sprintf("%.2f", metric.Max))
	}
	description = strings.ReplaceAll(description, "{unit}", rule.Unit)
	return description
}
