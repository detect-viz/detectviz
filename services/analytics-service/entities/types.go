package entities

// DetectionResult 檢測結果
type DetectionResult struct {
	Anomalies   []MetricData `json:"anomalies"`   // 異常檢測結果
	Predictions []MetricData `json:"predictions"` // 預測分析結果
	Timestamp   int64        `json:"timestamp"`   // 檢測時間戳
}
