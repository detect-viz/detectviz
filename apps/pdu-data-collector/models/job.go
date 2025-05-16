package models

type JobDetail struct {
	Enable         string `json:"Enable"`
	Name           string `json:"Name"`
	CronExpression string `json:"CronExpression"`
	Description    string `json:"Description"`
	DataRange      string `json:"DataRange"`
	Delay          string `json:"Delay"`
	Aggregate      string `json:"Aggregate"`
}

type Jobs struct {
	FactorySyncGlobalEnv      JobDetail `json:"FactorySyncGlobalEnv"`
	FactorySyncGlobalPDU      JobDetail `json:"FactorySyncGlobalPDU"`
	FactorySyncGlobalRule     JobDetail `json:"FactorySyncGlobalRule"`
	FactorySyncGlobalJob      JobDetail `json:"FactorySyncGlobalJob"`
	FactorySyncGlobalLog      JobDetail `json:"FactorySyncGlobalLog"`
	FactoryRecoverDCData      JobDetail `json:"FactoryRecoverDCData"`
	FactoryFlushDCData        JobDetail `json:"FactoryFlushDCData"`
	DCSyncGlobalTag           JobDetail `json:"DCSyncGlobalTag"`
	DCAggregateInfluxDBHourly JobDetail `json:"DCAggregateInfluxDBHourly"`
	DCAggregateInfluxDBDaily  JobDetail `json:"DCAggregateInfluxDBDaily"`
	DCBackupInfluxDB          JobDetail `json:"DCBackupInfluxDB"`
	DCOutputSQLServerMinute   JobDetail `json:"DCOutputSQLServerMinute"`
	DCOutputSQLServerHour     JobDetail `json:"DCOutputSQLServerHour"`
	DCOutputSQLServerEvent    JobDetail `json:"DCOutputSQLServerEvent"`
	DCRecoverInfluxDBData     JobDetail `json:"DCRecoverInfluxDBData"`
	DCFlushInfluxDBData       JobDetail `json:"DCFlushInfluxDBData"`
}
