package models

type EnvironmentModel struct {
	Simulate SimulateConfig `mapstructure:"simulate"`
	Global   GlobalConfig   `mapstructure:"global"`
	Factory  FactoryConfig  `mapstructure:"factory"`
	DC       DCConfig       `mapstructure:"dc"`
	Jobs     Jobs           `mapstructure:"jobs"`
}

type Envs struct {
	GlobalConfig struct {
		LicenseKey      string `json:"LicenseKey"`
		CrontabTimezone string `json:"CrontabTimezone"`
	} `json:"GlobalConfig"`
	FactoryConfig struct {
		AllowTags       string `json:"AllowTags"`
		SettingConfPath string `json:"SettingConfPath"`
		BatchSize       string `json:"BatchSize"`
		MaxWaitTime     string `json:"MaxWaitTime"`
	} `json:"FactoryConfig"`
	DCConfig struct {
		BatchSize   string `json:"BatchSize"`
		MaxWaitTime string `json:"MaxWaitTime"`
	} `json:"DCConfig"`
	InfluxDBBucket struct {
		Raw       string `json:"Raw"`
		Aggregate string `json:"Aggregate"`
		Event     string `json:"Event"`
		Log       string `json:"Log"`
	} `json:"InfluxDBBucket"`
	InfluxDBOptions struct {
		SetBatchSize          string `json:"SetBatchSize"`
		SetLogLevel           string `json:"SetLogLevel"`
		SetUseGzip            string `json:"SetUseGzip"`
		SetFlushInterval      string `json:"SetFlushInterval"`
		SetMaxRetries         string `json:"SetMaxRetries"`
		SetMaxRetryTime       string `json:"SetMaxRetryTime"`
		SetRetryInterval      string `json:"SetRetryInterval"`
		SetHTTPRequestTimeout string `json:"SetHTTPRequestTimeout"`
		SetApplicationName    string `json:"SetApplicationName"`
	} `json:"InfluxDBOptions"`
	SQLServerTable struct {
		Minute string `json:"Minute"`
		Hour   string `json:"Hour"`
		Event  string `json:"Event"`
	} `json:"SQLServerTable"`
	PDUSchema struct {
		Measurement        string `json:"Measurement"`
		TotalCurrentField  string `json:"TotalCurrent"`
		TotalEnergyField   string `json:"TotalEnergy"`
		TotalWattField     string `json:"TotalWatt"`
		BranchCurrentField string `json:"BranchCurrent"`
		BranchEnergyField  string `json:"BranchEnergy"`
		BranchWattField    string `json:"BranchWatt"`
		PhaseCurrentField  string `json:"PhaseCurrent"`
		PhaseWattField     string `json:"PhaseWatt"`
		PhaseVoltageField  string `json:"PhaseVoltage"`
		PhaseEnergyField   string `json:"PhaseEnergy"`
		BranchTag          string `json:"BranchTag"`
		PhaseTag           string `json:"PhaseTag"`
	} `json:"PDUSchema"`
	ColumnName struct {
		Factory        string `json:"Factory"`
		Phase          string `json:"Phase"`
		Datacenter     string `json:"Datacenter"`
		Room           string `json:"Room"`
		Rack           string `json:"Rack"`
		Side           string `json:"Side"`
		Model          string `json:"Model"`
		Protocol       string `json:"Protocol"`
		Manufacturer   string `json:"Manufacturer"`
		IP             string `json:"IP"`
		Name           string `json:"Name"`
		Panel          string `json:"Panel"`
		PLCIP          string `json:"PLCIP"`
		WifiClientIP   string `json:"WifiClientIP"`
		WifiClientPort string `json:"WifiClientPort"`
	} `json:"ColumnName"`
	DeviceType struct {
		PDU        string `json:"PDU"`
		Panel      string `json:"Panel"`
		WifiClient string `json:"WifiClient"`
		PLC        string `json:"PLC"`
		Switch     string `json:"Switch"`
		AP         string `json:"AP"`
	} `json:"DeviceType"`
	DCRecoveryDir struct {
		InfluxDBRaw     string `json:"InfluxDBRaw"`
		InfluxDBEvent   string `json:"InfluxDBEvent"`
		SQLServerMinute string `json:"SQLServerMinute"`
		SQLServerHour   string `json:"SQLServerHour"`
		SQLServerEvent  string `json:"SQLServerEvent"`
	} `json:"DCRecoveryDir"`
	FactoryRecoveryDir struct {
		DCMetrics string `json:"DCMetrics"`
		DCEvents  string `json:"DCEvents"`
	} `json:"FactoryRecoveryDir"`
}

type GlobalConfig struct {
	Mode             string   `mapstructure:"mode"`
	MaxConnsPerHost  int      `mapstructure:"max_conns_per_host"`
	ServerMode       string   `mapstructure:"server_mode"`
	CORSAllowHeaders []string `mapstructure:"cors_allow_headers"`
	Log              struct {
		Level   string `mapstructure:"level"`
		Path    string `mapstructure:"path"`
		MaxSize int    `mapstructure:"maxsize"`
		MaxAge  int    `mapstructure:"maxage"`
	} `mapstructure:"log"`
	HttpTimeout int `mapstructure:"http_timeout"`
}

type FactoryConfig struct {
	GroupName      string        `mapstructure:"group_name"`
	Port           string        `mapstructure:"port"`
	DCEndpoint     string        `mapstructure:"dc_endpoint"`
	DeviceScale    []DeviceScale `mapstructure:"device_scale"`
	InitGlobalData struct {
		PDU InitData `mapstructure:"pdu"`
		Env InitData `mapstructure:"env"`
		Job InitData `mapstructure:"job"`
		Log InitData `mapstructure:"log"`
	} `mapstructure:"init_global_data"`
}

type InitData struct {
	Name string `mapstructure:"name"`
	File string `mapstructure:"file"`
}

type DCConfig struct {
	Port      string          `mapstructure:"port"`
	InfluxDB  InfluxDBConfig  `mapstructure:"influxdb"`
	SQLServer SQLServerConfig `mapstructure:"sqlserver"`
	Mysql     MysqlConfig     `mapstructure:"mysql"`
}

type DeviceScale struct {
	Manufacturer string  `mapstructure:"manufacturer"`
	Current      float64 `mapstructure:"current"`
	Voltage      float64 `mapstructure:"voltage"`
	Watt         float64 `mapstructure:"watt"`
	Energy       float64 `mapstructure:"energy"`
}

type SimulateConfig struct {
	Enable                bool   `mapstructure:"enable"`
	URL                   string `mapstructure:"url"`
	MaxConcurrentRequests int    `mapstructure:"max_concurrent_requests"`
	Interval              int    `mapstructure:"interval"`
	Timeout               int    `mapstructure:"timeout"`
}

type InfluxDBConfig struct {
	URL        string `mapstructure:"url"`
	Org        string `mapstructure:"org"`
	Token      string `mapstructure:"token"`
	InfluxExec string `mapstructure:"influx_exec"`
	BackupPath string `mapstructure:"backup_path"`
}

type SQLServerConfig struct {
	Host      string `mapstructure:"host"`
	Port      string `mapstructure:"port"`
	DBName    string `mapstructure:"dbname"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	BatchSize int    `mapstructure:"batch_size"`
}

type MysqlConfig struct {
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	DBName      string `mapstructure:"dbname"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	MaxIdle     int    `mapstructure:"max_idle"`
	MaxLifeTime string `mapstructure:"max_life_time"`
	MaxOpenConn int    `mapstructure:"max_open_conn"`
	MigratePath string `mapstructure:"migrate_path"`
}
