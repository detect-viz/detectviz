package factory

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services"
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func RunFactoryCrontab() {

	// 確保目錄存在
	services.IfNotExistCreateDir(global.Envs.FactoryRecoveryDir.DCMetrics)
	services.IfNotExistCreateDir(global.Envs.FactoryRecoveryDir.DCEvents)

	// 更新 PDU 名單
	job := global.Jobs.FactorySyncGlobalPDU
	if job.Enable != "false" && job.CronExpression != "" {
		runUpdateGlobalPDUList(job, global.EnvConfig.Factory.InitGlobalData.PDU)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 更新環境配置
	job = global.Jobs.FactorySyncGlobalEnv
	if job.Enable != "false" && job.CronExpression != "" {
		runUpdateGlobalEnv(job, global.EnvConfig.Factory.InitGlobalData.Env)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 恢復數據
	job = global.Jobs.FactoryRecoverDCData
	if job.Enable != "false" && job.CronExpression != "" {
		runFactoryRecoverData(job)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	// 刷新數據
	job = global.Jobs.FactoryFlushDCData
	if job.Enable != "false" && job.CronExpression != "" {
		runFlushDCData(job)
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}
}

func runFlushDCData(job models.JobDetail) {

	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				FlushDCBatch()
			})))

	if err != nil {
		global.Logger.Error(fmt.Sprintf("Cron job [%s] 設置失敗: %v", job.Name, err),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	} else {
		global.Logger.Info(fmt.Sprintf("Cron job [%s] 設置成功, CronID [%v]", job.Name, cronID),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}
}

func runUpdateGlobalPDUList(job models.JobDetail, initData models.InitData) {

	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				UpdateConfig(initData)
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

func runUpdateGlobalEnv(job models.JobDetail, initData models.InitData) {

	cronID, err := global.Crontab.AddJob(
		strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
			Then(cron.FuncJob(func() {
				UpdateConfig(initData)
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

func runFactoryRecoverData(job models.JobDetail) {

	if job.Enable != "false" && job.CronExpression != "" {

		cronID, err := global.Crontab.AddJob(
			strings.Join([]string{global.Envs.GlobalConfig.CrontabTimezone, job.CronExpression}, " "),
			cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).
				Then(cron.FuncJob(func() {
					// Run the job logic
					RecoveryFactoryData(global.Envs.FactoryRecoveryDir.DCMetrics)
					RecoveryFactoryData(global.Envs.FactoryRecoveryDir.DCEvents)
				})))

		if err != nil {
			global.Logger.Error(fmt.Sprintf("Cron job [%s] 設置失敗: %v", job.Name, err),
				zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
		} else {
			global.Logger.Info(fmt.Sprintf("Cron job [%s] 設置成功, CronID [%v]", job.Name, cronID),
				zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
		}
	} else {
		global.Logger.Warn(fmt.Sprintf("Cron job [%s] 未啟動，因為沒有設置 crontab 週期或未啟用", job.Name),
			zap.Any(global.Logs.LoadCrontab.Name, global.Logs.LoadCrontab))
	}

	global.Crontab.Start()
}
