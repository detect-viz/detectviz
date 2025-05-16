package dc

import (
	"bimap-zbox/databases"
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bimap-zbox/services"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

func OutputDcim(job models.JobDetail) {
	var points int
	startTime := time.Now() // 開始時間
	defer func() {
		// 結束時間
		duration := time.Since(startTime)
		fields := map[string]interface{}{
			"duration": int(duration.Seconds()),
			"points":   points,
		}
		tags := map[string]string{
			"name": job.Name,
		}
		global.Logger.Info(job.Name, zap.Any("fields", fields), zap.Any("tags", tags))
		databases.WriteLogToInfluxDB(job.Name, tags, fields)

	}()

	bucket := global.Envs.InfluxDBBucket.Raw
	dataRange, _ := strconv.Atoi(job.DataRange)
	start := time.Now().Unix() + int64(dataRange)
	stop := time.Now().Unix()
	measurement := global.Envs.PDUSchema.Measurement
	fields := []string{
		global.Envs.PDUSchema.BranchCurrentField,
		global.Envs.PDUSchema.BranchEnergyField,
		global.Envs.PDUSchema.BranchWattField,
		global.Envs.PDUSchema.PhaseVoltageField,
	}

	// 延遲時間，以秒為單位
	delay, _ := strconv.Atoi(job.Delay)
	time.Sleep(time.Duration(delay) * time.Second)

	var sqlData []map[string]interface{}
	var err error
	var table, recoveryDir string

	switch job.Name {
	case global.Jobs.DCOutputSQLServerMinute.Name:
		sqlData, err = queryInfluxDBForMinuteData(start, stop, bucket, measurement, fields)
		table = global.Envs.SQLServerTable.Minute
		recoveryDir = global.Envs.DCRecoveryDir.SQLServerMinute
	case global.Jobs.DCOutputSQLServerHour.Name:
		aggregate, _ := strconv.Atoi(job.Aggregate)
		sqlData, err = queryInfluxDBForHourData(start, stop, bucket, measurement, fields, aggregate)
		table = global.Envs.SQLServerTable.Hour
		recoveryDir = global.Envs.DCRecoveryDir.SQLServerHour
	}

	if err != nil {
		global.Logger.Error("查詢 InfluxDB 資料失敗", zap.Error(err))
		return
	}
	points = len(sqlData)
	err = databases.WriteToSqlServer(table, sqlData)
	//err = errors.New("錯誤測試")
	if err != nil {
		global.Logger.Error("批量寫入 SqlServer 失敗", zap.Error(err))

		// 如果發送失敗，將數據保存至 JSON 檔案
		filename := filepath.Join(recoveryDir, time.Now().Format(time.RFC3339Nano)+".json")
		saveErr := services.SaveDataToFile(filename, sqlData)
		if saveErr != nil {
			global.Logger.Error(
				"保存 DC 資料至 SqlServer JSON 檔案失敗",
				zap.Error(saveErr),
			)
		} else {
			global.Logger.Info(
				"批量寫入 SqlServer 失敗，已保存至 JSON 檔案",
				zap.String("filename", filename),
			)
		}
	}
}

func queryInfluxDBForMinuteData(start, stop int64, bucket, measurement string, fields []string) ([]map[string]interface{}, error) {
	startTime := time.Now() // 開始時間
	defer func() {
		// 結束時間
		duration := time.Since(startTime)
		global.Logger.Info(
			fmt.Sprintf("queryInfluxDBForMinuteData 執行時間: %v", duration),
		)
	}()
	// 使用 InfluxDB 客戶端初始化
	client := global.InfluxDB
	defer client.Close()
	queryAPI := client.QueryAPI(global.EnvConfig.DC.InfluxDB.Org)

	fmt.Println(start, stop)

	// 構建查詢語句
	query := fmt.Sprintf(`
	from(bucket: "%s")
	  |> range(start: %v, stop: %v)
	  |> filter(fn: (r) => r["_measurement"] == "%s")
	  |> filter(fn: (r) => %s)
	  |> last()
	  |> keep(columns: ["_field", "_value", "%s", "%s", "%s"])
	`, bucket, start, stop, measurement, buildFieldFilter(fields),
		global.Envs.ColumnName.Name,
		global.Envs.PDUSchema.BranchTag,
		global.Envs.PDUSchema.PhaseTag,
	)
	fmt.Println(query)
	// 執行查詢
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	// 資料轉換為 []map[string]interface{}
	var sqlData []map[string]interface{}
	for result.Next() {
		record := result.Record()
		data := make(map[string]interface{})
		data["PDU"] = record.ValueByKey(global.Envs.ColumnName.Name)
		data["CreateTime"] = time.Unix(stop, 0) // 使用查詢的 stop 作為 CreateTime

		schema := global.Envs.PDUSchema
		// 設定 DataType
		switch record.Field() {
		case schema.BranchCurrentField, schema.PhaseCurrentField:
			data["DataType"] = "C"
		case schema.BranchEnergyField, schema.PhaseEnergyField:
			data["DataType"] = "E"
		case schema.BranchWattField, schema.PhaseWattField:
			data["DataType"] = "P"
		case schema.PhaseVoltageField:
			data["DataType"] = "V"
		}

		// 設定 Bank
		if record.Field() == schema.PhaseVoltageField {
			data["Bank"] = record.ValueByKey(schema.PhaseTag)
		} else {
			data["Bank"] = record.ValueByKey(schema.BranchTag)
		}

		data["Value"] = fmt.Sprintf("%.2f", record.Value().(float64))

		// 加入到結果集
		sqlData = append(sqlData, data)
	}

	if result.Err() != nil {
		return nil, result.Err()
	}
	return sqlData, nil
}

func queryInfluxDBForHourData(start, stop int64, bucket, measurement string, fields []string, aggregateWindow int) ([]map[string]interface{}, error) {
	schema := global.Envs.PDUSchema
	startTime := time.Now() // 開始時間
	defer func() {
		// 結束時間
		duration := time.Since(startTime)
		global.Logger.Info(
			fmt.Sprintf("queryInfluxDBForHourData 執行時間: %v", duration),
		)
	}()
	// 使用 InfluxDB 客戶端初始化
	client := global.InfluxDB
	defer client.Close()
	queryAPI := client.QueryAPI(global.EnvConfig.DC.InfluxDB.Org)

	// 構建查詢語句
	query := fmt.Sprintf(`
	from(bucket: "%s")
	  |> range(start: %v, stop: %v)
	  |> filter(fn: (r) => r["_measurement"] == "%s")
	  |> filter(fn: (r) => %s)
	  |> aggregateWindow(every: %vs, fn: last, createEmpty: false)
	  |> keep(columns: ["_field", "_value","%s","%s","%s"])
	  |> group(columns: ["_field","%s","%s","%s"], mode:"by")  
	`, bucket, start, stop, measurement, buildFieldFilter(fields), aggregateWindow,
		global.Envs.ColumnName.Name,
		schema.BranchTag,
		schema.PhaseTag,
		global.Envs.ColumnName.Name,
		schema.BranchTag,
		schema.PhaseTag,
	)
	global.Logger.Debug(query)
	//fmt.Println(query)
	// 執行查詢
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	// 資料轉換為 []map[string]interface{}
	var sqlData []map[string]interface{}
	groupValues := []string{}
	var currentGroup string
	var data map[string]interface{}

	for result.Next() {
		record := result.Record()
		value := record.Value()

		// 建立唯一分組鍵（例如 "name_field_branch_phase"）
		groupKey := fmt.Sprintf("%s_%s_%s_%s", record.ValueByKey(global.Envs.ColumnName.Name), record.Field(), record.ValueByKey(global.Envs.PDUSchema.BranchTag), record.ValueByKey(global.Envs.PDUSchema.PhaseTag))

		// 當前組第一次出現或發生變更時，將前一組結果添加到 sqlData
		if currentGroup != "" && currentGroup != groupKey {
			if len(groupValues) > 0 {
				data["Value"] = strings.Join(groupValues, ",")
				sqlData = append(sqlData, data)
				groupValues = []string{}
			}
		}

		// 當前組更新，設置 data 基本資料
		if currentGroup != groupKey {
			data = make(map[string]interface{})
			data["PDU"] = record.ValueByKey("name")
			data["CreateTime"] = time.Unix(stop, 0) // 使用查詢的 stop 作為 CreateTime

			// 設定 DataType
			switch record.Field() {
			case schema.BranchCurrentField, schema.PhaseCurrentField:
				data["DataType"] = "C"
			case schema.BranchEnergyField, schema.PhaseEnergyField:
				data["DataType"] = "E"
			case schema.BranchWattField, schema.PhaseWattField:
				data["DataType"] = "P"
			case schema.PhaseVoltageField:
				data["DataType"] = "V"
			}

			// 設定 Bank
			if record.Field() == schema.PhaseVoltageField {
				data["Bank"] = record.ValueByKey(global.Envs.PDUSchema.PhaseTag)
			} else {
				data["Bank"] = record.ValueByKey(global.Envs.PDUSchema.BranchTag)
			}

			currentGroup = groupKey
		}

		// 累積 _value 值
		groupValues = append(groupValues, fmt.Sprintf("%.2f", value))
	}

	// 處理最後一個分組的數據
	if len(groupValues) > 0 {
		data["Value"] = strings.Join(groupValues, ",")
		sqlData = append(sqlData, data)
	}

	if result.Err() != nil {
		return nil, result.Err()
	}
	return sqlData, nil
}

// 輔助函式，建立 fields 過濾語句
func buildFieldFilter(fields []string) string {
	filter := ""
	for i, field := range fields {
		if i > 0 {
			filter += " or "
		}
		filter += fmt.Sprintf(`r["_field"] == "%s"`, field)
	}
	return filter
}
