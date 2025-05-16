package dc

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services"

	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func RunDCCrontab() {

	// 確保目錄存在
	services.IfNotExistCreateDir(global.Envs.DCRecoveryDir.InfluxDBRaw)
	services.IfNotExistCreateDir(global.Envs.DCRecoveryDir.InfluxDBEvent)
	services.IfNotExistCreateDir(global.Envs.DCRecoveryDir.SQLServerMinute)
	services.IfNotExistCreateDir(global.Envs.DCRecoveryDir.SQLServerHour)
	services.IfNotExistCreateDir(global.Envs.DCRecoveryDir.SQLServerEvent)

	// 每10分鐘匯總數據到 InfluxDB
	job := global.Jobs.DCAggregateInfluxDBHourly
	if job.Enable != "false" && job.CronExpression != "" {
		runAggregateHourlyData(job)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 每24小時匯總數據到 InfluxDB
	job = global.Jobs.DCAggregateInfluxDBDaily
	if job.Enable != "false" && job.CronExpression != "" {
		runAggregateDailyData(job)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 每10分鐘輸出數據到 SQL Server
	job = global.Jobs.DCOutputSQLServerMinute
	if job.Enable != "false" && job.CronExpression != "" {
		runDCMinuteOutput(job)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 每3小時輸出數據到 SQL Server
	job = global.Jobs.DCOutputSQLServerHour
	if job.Enable != "false" && job.CronExpression != "" {
		runDCHourOutput(job)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 重傳失敗的 JSON 檔案
	job = global.Jobs.DCRecoverInfluxDBData
	if job.Enable != "false" && job.CronExpression != "" {
		runDCRecoverData(job)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 備份 InfluxDB
	job = global.Jobs.DCBackupInfluxDB
	if job.Enable != "false" && job.CronExpression != "" {
		runBackupInfluxDB(job)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 刷新 DC 數據
	job = global.Jobs.DCFlushInfluxDBData
	if job.Enable != "false" && job.CronExpression != "" {
		runFlushInfluxDBData(job)
	}
}

func runFlushInfluxDBData(job models.JobDetail) {
	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				FlushInfluxDBBatch()
			})))

	if err != nil {
		global.Logger.Error(fmt.Sprintf("Cron job [%s] 設置失敗: %v", job.Name, err),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	} else {
		global.Logger.Info(fmt.Sprintf("Cron job [%s] 設置成功, CronID [%v]", job.Name, cronID),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}
}

func runAggregateHourlyData(job models.JobDetail) {
	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				AggregateHourlyData(job)
			})))

	if err != nil {
		global.Logger.Error(fmt.Sprintf("Cron job [%s] 設置失敗: %v", job.Name, err),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	} else {
		global.Logger.Info(fmt.Sprintf("Cron job [%s] 設置成功, CronID [%v]", job.Name, cronID),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}
}

func runAggregateDailyData(job models.JobDetail) {
	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				AggregateDailyData(job)
			})))

	if err != nil {
		global.Logger.Error(fmt.Sprintf("Cron job [%s] 設置失敗: %v", job.Name, err),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	} else {
		global.Logger.Info(fmt.Sprintf("Cron job [%s] 設置成功, CronID [%v]", job.Name, cronID),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}
}

func runDCMinuteOutput(job models.JobDetail) {
	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				OutputDcim(job)
			})))

	if err != nil {
		global.Logger.Error(fmt.Sprintf("Cron job [%s] 設置失敗: %v", job.Name, err),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	} else {
		global.Logger.Info(fmt.Sprintf("Cron job [%s] 設置成功, CronID [%v]", job.Name, cronID),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

}

func runDCHourOutput(job models.JobDetail) {
	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				OutputDcim(job)
			})))

	if err != nil {
		global.Logger.Error(fmt.Sprintf("Cron job [%s] 設置失敗: %v", job.Name, err),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	} else {
		global.Logger.Info(fmt.Sprintf("Cron job [%s] 設置成功, CronID [%v]", job.Name, cronID),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}
}

func runDCRecoverData(job models.JobDetail) {

	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				RecoveryInfluxData(global.Envs.DCRecoveryDir.InfluxDBRaw, global.Envs.InfluxDBBucket.Raw)
				RecoveryInfluxData(global.Envs.DCRecoveryDir.InfluxDBEvent, global.Envs.InfluxDBBucket.Event)
				RecoverySqlServerData(global.Envs.DCRecoveryDir.SQLServerMinute, global.Envs.SQLServerTable.Minute)
				RecoverySqlServerData(global.Envs.DCRecoveryDir.SQLServerHour, global.Envs.SQLServerTable.Hour)
				RecoverySqlServerData(global.Envs.DCRecoveryDir.SQLServerEvent, global.Envs.SQLServerTable.Event)

			})))

	if err != nil {
		global.Logger.Error(fmt.Sprintf("Cron job [%s] 設置失敗: %v", job.Name, err),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	} else {
		global.Logger.Info(fmt.Sprintf("Cron job [%s] 設置成功, CronID [%v]", job.Name, cronID),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	global.Crontab.Start()
}

func runBackupInfluxDB(job models.JobDetail) {
	if job.Enable != "false" && job.CronExpression != "" {
		cronID, err := global.Crontab.AddJob(
			strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
			cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
				Then(cron.FuncJob(func() {
					err := RunBackupInfluxDB()
					if err != nil {
						global.Logger.Error(err.Error(),
							zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
					}
				})))
		if err != nil {
			global.Logger.Error(err.Error(),
				zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
		} else {
			global.Logger.Info(fmt.Sprintf("CronID [%v] Set [%s]", cronID, job.Name),
				zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
		}
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	global.Crontab.Start()

}
