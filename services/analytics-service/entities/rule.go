package entities

type Rule struct {
	Name                string         `yaml:"name"`
	Metric              string         `yaml:"metric"`
	Model               string         `yaml:"model"`
	Config              map[string]any `yaml:"config"`
	Group               string         `yaml:"group"`
	DescriptionTemplate struct {
		Critical string `yaml:"critical"`
		Warning  string `yaml:"warning"`
	} `yaml:"description_template"`
}

type DataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type Data struct {
	Current []DataPoint `json:"current,omitempty"`
	History []DataPoint `json:"history,omitempty"`
}
