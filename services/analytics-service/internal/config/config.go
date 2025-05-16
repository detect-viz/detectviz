package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Panel 定義面板配置
type Panel struct {
	Title          string   `yaml:"title"`
	Description    string   `yaml:"description"`
	PromptTemplate string   `yaml:"prompt_template"`
	Rules          []string `yaml:"rules"`
}

// ProphetSettings 定義 Prophet 設置
type ProphetSettings struct {
	Horizon int64  `yaml:"horizon"`
	Period  string `yaml:"period"`
}

// PredictionPanel 定義預測面板配置
type PredictionPanel struct {
	Title           string          `yaml:"title"`
	Description     string          `yaml:"description"`
	Metrics         []string        `yaml:"metrics"`
	ProphetSettings ProphetSettings `yaml:"prophet_settings"`
}

// AnomalyService 異常檢測服務配置
type AnomalyService struct {
	Mode    string
	APIHost string
	CLIPath string
	Timeout int          `yaml:"timeout"`
	Retry   RetryConfig  `yaml:"retry"`
	Models  ModelsConfig `yaml:"models"`
}

// Config 存儲應用程序配置
type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Profiles         map[string]Profile         `yaml:"profiles"`
	Rules            map[string]Rule            `yaml:"rules"`
	Panels           map[string]Panel           `yaml:"panels"`
	PredictionPanels map[string]PredictionPanel `yaml:"prediction_panels"`

	LLMAPI struct {
		Mode        string      `yaml:"mode"` // api 或 local
		Timeout     int         `yaml:"timeout"`
		MaxTokens   int         `yaml:"max_tokens"`
		Temperature float64     `yaml:"temperature"`
		Retry       RetryConfig `yaml:"retry"` // 共用的重試配置

		API struct {
			URL          string `yaml:"url"`
			Model        string `yaml:"model"`
			APIKey       string `yaml:"api_key"`
			Organization string `yaml:"organization,omitempty"`
		} `yaml:"api"`

		Local struct {
			Host          string  `yaml:"host"` // 本地 LLM 服務端點
			ModelPath     string  `yaml:"model_path"`
			Device        string  `yaml:"device"`
			TopP          float64 `yaml:"top_p"`
			TopK          int     `yaml:"top_k"`
			RepeatPenalty float64 `yaml:"repeat_penalty"`
		} `yaml:"local"`
	} `yaml:"llm_api"`

	AnomalyService AnomalyService `yaml:"anomaly_service"`
}

// Profile 監控配置
type Profile struct {
	Title            string   `yaml:"title"`
	Description      string   `yaml:"description"`
	Prompt           string   `yaml:"prompt"`
	Panels           []string `yaml:"panels"`
	PredictionPanels []string `yaml:"prediction_panels"`
}

// Rule 定義異常檢測規則
type Rule struct {
	MetricName string         `yaml:"metric_name"`
	Model      string         `yaml:"model"`
	Unit       string         `yaml:"unit"`
	Config     map[string]any `yaml:"config"`
	Templates  RuleTemplates  `yaml:"templates"`
}

// RuleTemplates 規則模板
type RuleTemplates struct {
	Desc   map[string]string `yaml:"desc"`
	Prompt map[string]string `yaml:"prompt"`
}

// DescriptionTemplates 描述模板
type DescriptionTemplates struct {
	Critical string `yaml:"critical,omitempty"`
	Warning  string `yaml:"warning,omitempty"`
	Normal   string `yaml:"normal,omitempty"`
	Anomaly  string `yaml:"anomaly,omitempty"` // 用於 isolation_forest
}

// PromptTemplates 提示詞模板
type PromptTemplates struct {
	Critical string `yaml:"critical,omitempty"`
	Warning  string `yaml:"warning,omitempty"`
	Anomaly  string `yaml:"anomaly,omitempty"` // 用於 isolation_forest
}

// ModelsConfig 模型配置
type ModelsConfig struct {
	AbsoluteThreshold   map[string]interface{} `yaml:"absolute_threshold"`
	PercentageThreshold map[string]interface{} `yaml:"percentage_threshold"`
	MovingAverage       map[string]interface{} `yaml:"moving_average"`
	IsolationForest     map[string]interface{} `yaml:"isolation_forest"`
	Prophet             struct {
		SeasonalityMode       string  `yaml:"seasonality_mode"`
		ChangepointPriorScale float64 `yaml:"changepoint_prior_scale"`
		IntervalWidth         float64 `yaml:"interval_width"`
		UncertaintySamples    int     `yaml:"uncertainty_samples"`
		Horizon               string  `yaml:"horizon"`
		Period                string  `yaml:"period"`
		HolidaysPriorScale    float64 `yaml:"holidays_prior_scale"`
		WeeklySeasonality     bool    `yaml:"weekly_seasonality"`
		DailySeasonality      bool    `yaml:"daily_seasonality"`
		YearlySeasonality     bool    `yaml:"yearly_seasonality"`
	} `yaml:"prophet"`
}

// RetryConfig 重試配置
type RetryConfig struct {
	MaxAttempts     int `yaml:"max_attempts"`
	InitialInterval int `yaml:"initial_interval"`
	MaxInterval     int `yaml:"max_interval"`
}

// Load 加載配置
func Load() (*Config, error) {
	// 讀取 settings.yaml
	data, err := os.ReadFile("config/settings.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 讀取 panels.yaml
	panelsData, err := os.ReadFile("config/panels.yaml")
	if err != nil {
		return nil, err
	}

	var panelsConfig struct {
		Panels map[string]Panel `yaml:"panels"`
	}
	if err := yaml.Unmarshal(panelsData, &panelsConfig); err != nil {
		return nil, err
	}

	cfg.Panels = panelsConfig.Panels

	// 讀取 profiles.yaml
	profilesData, err := os.ReadFile("config/profiles.yaml")
	if err != nil {
		return nil, err
	}

	var profilesConfig struct {
		Profiles map[string]Profile `yaml:"profiles"`
	}
	if err := yaml.Unmarshal(profilesData, &profilesConfig); err != nil {
		return nil, err
	}

	cfg.Profiles = profilesConfig.Profiles

	// 讀取 rules.yaml
	rulesData, err := os.ReadFile("config/rules.yaml")
	if err != nil {
		return nil, err
	}

	var rulesConfig struct {
		Rules map[string]Rule `yaml:"rules"`
	}
	if err := yaml.Unmarshal(rulesData, &rulesConfig); err != nil {
		return nil, err
	}

	cfg.Rules = rulesConfig.Rules

	return &cfg, nil
}

