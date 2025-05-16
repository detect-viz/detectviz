package databases

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"go.uber.org/zap"
)

func LoadInfluxDB() {
	client := NewInfluxDBClient(time.Second)
	defer client.Close()
	ctx := context.Background()

	d, err := client.Health(ctx)
	if err != nil {
		global.Logger.Error(err.Error(),
			zap.Any(global.Logs.ConnectInfluxDB.Name, global.Logs.ConnectInfluxDB))
		global.InfluxDB = client
		return
	}
	global.InfluxDB = client
	global.Logger.Info(fmt.Sprintf("influxdb connection success: %v", *d.Message),
		zap.Any(global.Logs.ConnectInfluxDB.Name, global.Logs.ConnectInfluxDB))

}

func NewInfluxDBClient(precision time.Duration) influxdb2.Client {
	db := global.EnvConfig.DC.InfluxDB
	env := global.Envs.InfluxDBOptions

	batchSize, _ := strconv.Atoi(env.SetBatchSize)
	logLevel, _ := strconv.Atoi(env.SetLogLevel)
	flushInterval, _ := strconv.Atoi(env.SetFlushInterval)
	maxRetries, _ := strconv.Atoi(env.SetMaxRetries)
	maxRetryTime, _ := strconv.Atoi(env.SetMaxRetryTime)
	retryInterval, _ := strconv.Atoi(env.SetRetryInterval)
	httpRequestTimeout, _ := strconv.Atoi(env.SetHTTPRequestTimeout)

	influxdb := influxdb2.NewClientWithOptions(db.URL, db.Token,
		influxdb2.DefaultOptions().
			//*** 指定最粗時間戳的精度
			SetPrecision(precision).
			//*** 最佳批量大小是 5000 行
			SetBatchSize(uint(batchSize)).
			//*** 0=error, 1=warning, 2=info, 3=debug, nil 禁用
			SetLogLevel(uint(logLevel)).
			//*** 壓縮傳輸可提高5倍速
			SetUseGZip(env.SetUseGzip == "true").
			//*** 如果緩衝區3s尚未寫入，則刷新緩衝區
			SetFlushInterval(uint(flushInterval)).
			//*** 失敗寫入的最大重試次數
			SetMaxRetries(uint(maxRetries)).
			//*** 失敗寫入的最大重試時間
			SetMaxRetryTime(uint(maxRetryTime)).
			//*** 重試之間的最大延遲
			SetRetryInterval(uint(retryInterval)).
			//*** 長時間寫資料設定
			SetHTTPRequestTimeout(uint(httpRequestTimeout)).
			SetApplicationName(env.SetApplicationName))

	return influxdb
}

// * 寫入 InfluxDB
func WriteToInfluxDB(bucket string, data []models.Point) error {
	startTime := time.Now() // 開始時間

	points := []string{}

	db := global.EnvConfig.DC.InfluxDB
	client := NewInfluxDBClient(time.Second)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(db.Org, bucket)

	for _, line := range data {
		points = append(points, toLineProtocol(line))
	}

	if err := writeAPI.WriteRecord(context.Background(), points...); err != nil {
		global.Logger.Error(fmt.Sprintf("WriteToInfluxDB Error: %v", err),
			zap.Any(global.Logs.OutputInfluxDB.Name, global.Logs.OutputInfluxDB))
		return err
	}
	duration := time.Since(startTime)
	global.Logger.Info(fmt.Sprintf("批量寫入 InfluxDB 成功，執行時間: %v", duration), zap.Int("points", len(points)))

	return nil
}

func toLineProtocol(p models.Point) string {
	var builder strings.Builder

	builder.WriteString(p.Name)

	var tagStrings []string
	for key, value := range p.Tags {
		tagStrings = append(tagStrings, fmt.Sprintf("%s=%s", key, value))
	}
	builder.WriteString("," + strings.Join(tagStrings, ","))

	var fieldStrings []string
	for key, value := range p.Fields {
		fieldStrings = append(fieldStrings, fmt.Sprintf("%s=%v", key, value))
	}
	builder.WriteString(" " + strings.Join(fieldStrings, ","))

	builder.WriteString(" " + fmt.Sprintf("%d", p.Time))

	return builder.String()
}

func WriteLogToInfluxDB(name string, tags map[string]string, fields map[string]interface{}) {
	client := global.InfluxDB
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(global.EnvConfig.DC.InfluxDB.Org, global.Envs.InfluxDBBucket.Log)

	// 寫入到 InfluxDB
	point := influxdb2.NewPoint(name, tags, fields, time.Now())
	err := writeAPI.WritePoint(context.Background(), point)
	if err != nil {
		fmt.Printf("failed to write log to influxdb: %v\n", err)
	}
}
