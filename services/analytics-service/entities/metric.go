package entities

// MetricData 指標數據結構
type MetricData struct {
	Timestamp int64   `json:"timestamp"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit,omitempty"`
	// 移動平均模型特有字段
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
	// Prophet 預測模型特有字段
	Predict bool `json:"predict,omitempty"`
	// 異常檢測結果
	Severity string `json:"severity,omitempty"`
	Anomaly  bool   `json:"anomaly,omitempty"`
}

// MetricResponse Python API 回應結構
type MetricResponse struct {
	Data []MetricData `json:"data"`
}

// MetricRequest 請求數據結構
type MetricRequest struct {
	Data struct {
		Current []MetricData `json:"current"`
		History []MetricData `json:"history"`
	} `json:"data"`
}
