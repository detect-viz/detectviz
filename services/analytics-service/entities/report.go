package entities

// MetricReport 指標報告數據結構
type MetricReport struct {
	Timestamp   int64   `json:"timestamp"`
	Value       float64 `json:"value"`
	Severity    string  `json:"severity"`
	Anomaly     bool    `json:"anomaly"`
	Min         float64 `json:"min,omitempty"`
	Max         float64 `json:"max,omitempty"`
	Description string  `json:"description"`
	Threshold   float64 `json:"threshold"`
}

// PanelReport 面板報告結構
type PanelReport struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Metrics     []MetricReport `json:"metrics"`
}

// Report 完整報告結構
type Report struct {
	Timestamp   int64         `json:"timestamp"`
	ReportID    string        `json:"report_id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Content     string        `json:"content"`
	Panels      []PanelReport `json:"panels"`
}

type ReportResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Summary string `json:"summary"`
}
