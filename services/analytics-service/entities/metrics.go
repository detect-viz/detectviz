package entities

type MetricInfo struct {
	MetricName string `json:"metric_name"`
	Unit       string `json:"unit"`
}

type MetricsResponse struct {
	Metrics []MetricInfo `json:"metrics"`
	Current bool         `json:"current"`
	History bool         `json:"history"`
}
