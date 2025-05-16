package models

import "time"

type Event struct {
	Name        string `json:"Name"`
	Code        string `json:"Code"`
	Category    string `json:"Category"`
	Level       string `json:"Level"`
	Description string `json:"Description"`
}

type Logs struct {
	OutputInfluxDB    Event `json:"OutputInfluxDB"`
	AggregateInfluxDB Event `json:"AggregateInfluxDB"`
	ConnectInfluxDB   Event `json:"ConnectInfluxDB"`
	BackupInfluxDB    Event `json:"BackupInfluxDB"`
	OutputSqlServer   Event `json:"OutputSqlServer"`
	ConnectSqlServer  Event `json:"ConnectSqlServer"`
	ConnectMysql      Event `json:"ConnectMysql"`
	LoadMigration     Event `json:"LoadMigration"`
	LoadEnvConfig     Event `json:"LoadEnvConfig"`
	LoadCrontab       Event `json:"LoadCrontab"`
	MatchTag          Event `json:"MatchTag"`
}

type StepLog struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  float64   `json:"duration"`
	Point     int       `json:"point"`
}
