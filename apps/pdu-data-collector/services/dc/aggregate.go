package dc

import (
	"bimap-zbox/databases"
	"bimap-zbox/global"
	"bimap-zbox/models"
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	hourlyTag = "hourly"
	dailyTag  = "daily"
	aggTag    = "interval"
)

// * 從 raw 取 min/max/first/last/mean 寫到 agg
func AggregateHourlyData(job models.JobDetail) {
	startTime := time.Now() // 開始時間
	defer func() {
		// 結束時間
		duration := time.Since(startTime)
		fields := map[string]interface{}{
			"duration": int(duration.Seconds()),
		}
		tags := map[string]string{
			"name": job.Name,
		}
		global.Logger.Info(job.Name, zap.Any("fields", fields), zap.Any("tags", tags))
		databases.WriteLogToInfluxDB(job.Name, tags, fields)
	}()

	queryAPI := databases.NewInfluxDBClient(time.Second).QueryAPI(global.EnvConfig.DC.InfluxDB.Org)
	var start, end int64

	// 轉換為開始時間戳和結束時間戳
	hourAgo := time.Now().Add(-time.Hour)
	start = time.Date(hourAgo.Year(), hourAgo.Month(), hourAgo.Day(), hourAgo.Hour(), 0, 0, 0, hourAgo.Location()).Unix()
	end = time.Date(hourAgo.Year(), hourAgo.Month(), hourAgo.Day(), hourAgo.Hour(), 59, 59, 0, hourAgo.Location()).Unix()

	aggList := []string{"max", "min", "first", "last", "mean"}

	var wg sync.WaitGroup

	for _, agg := range aggList {
		wg.Add(1)
		go func(agg string) {
			defer wg.Done()

			// Flux 查詢
			fluxQuery := fmt.Sprintf(`
			from(bucket: "%s")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "%s")
			|> filter(fn: (r) => r["_value"] >= 0)
			|> set(key: "%s", value: "%s")
			|> %s()
			|> map(fn: (r) => ({ r with _field: r._field + "_%s" }))
			|> map(fn: (r) => ({r with _time: r._stop}))
			|> to(bucket: "%s")
			`,
				global.Envs.InfluxDBBucket.Raw,
				start,
				end,
				global.Envs.PDUSchema.Measurement,
				aggTag,
				hourlyTag,
				agg,
				agg,
				global.Envs.InfluxDBBucket.Aggregate,
			)

			// Debug 輸出 Flux 查詢
			if global.EnvConfig.Global.Log.Level == "debug" {
				global.Logger.Debug(fluxQuery)
			}

			// 執行查詢並檢查錯誤
			if _, err := queryAPI.Query(context.Background(), fluxQuery); err != nil {
				global.Logger.Error(
					job.Name,
					zap.String("message", err.Error()),
				)
			}
		}(agg)
	}

	// 等待所有 goroutines 完成
	wg.Wait()
}

// * 從 raw 取 min/max/first/last/mean 寫到 agg
func AggregateDailyData(job models.JobDetail) {
	startTime := time.Now() // 開始時間
	defer func() {
		// 結束時間
		duration := time.Since(startTime)
		fields := map[string]interface{}{
			"duration": int(duration.Seconds()),
		}
		tags := map[string]string{
			"name": job.Name,
		}
		global.Logger.Info(job.Name, zap.Any("fields", fields), zap.Any("tags", tags))
		databases.WriteLogToInfluxDB(job.Name, tags, fields)
	}()

	queryAPI := databases.NewInfluxDBClient(time.Second).QueryAPI(global.EnvConfig.DC.InfluxDB.Org)
	var start, end int64

	yesterday := time.Now().AddDate(0, 0, -1)
	start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location()).Unix()
	end = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, yesterday.Location()).Unix()

	aggList := []string{"max", "min", "first", "last", "mean"}

	var wg sync.WaitGroup

	for _, agg := range aggList {
		wg.Add(1)
		go func(agg string) {
			defer wg.Done()
			// Flux 查詢
			fluxQuery := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => r["_measurement"] == "%s")
			|> filter(fn: (r) => r["interval"] == "%s")
			|> filter(fn: (r) => r["_field"] =~ /_%s/)
			|> %s()
			|> map(fn: (r) => ({r with _time: r._stop}))
			|> set(key: "%s", value: "%s")
			|> to(bucket: "%s")
		`,
				global.Envs.InfluxDBBucket.Aggregate,
				start,
				end,
				global.Envs.PDUSchema.Measurement,
				hourlyTag,
				agg,
				agg,
				aggTag,
				dailyTag,
				global.Envs.InfluxDBBucket.Aggregate,
			)

			// Debug 輸出 Flux 查詢
			if global.EnvConfig.Global.Log.Level == "debug" {
				global.Logger.Debug(fluxQuery)
			}

			// 執行查詢並檢查錯誤
			if _, err := queryAPI.Query(context.Background(), fluxQuery); err != nil {
				global.Logger.Error(
					job.Name,
					zap.String("message", err.Error()),
				)
			}
		}(agg)
	}

	// 等待所有 goroutines 完成
	wg.Wait()
}
