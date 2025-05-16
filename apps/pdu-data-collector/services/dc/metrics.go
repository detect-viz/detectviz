package dc

import (
	"strconv"
	"sync"
	"time"

	"bimap-zbox/global"
	"bimap-zbox/models"
)

var (
	dcPointBatch    []models.Point
	dcPointMutex    sync.Mutex
	dcLastBatchTime time.Time
)

func init() {
	dcLastBatchTime = time.Now() // 初始化時間
}

// AddPoints 添加新的數據點到批次中
func AddPoints(points []models.Point) error {
	if len(points) == 0 {
		return nil
	}

	dcPointMutex.Lock()
	defer dcPointMutex.Unlock()

	dcPointBatch = append(dcPointBatch, points...)

	// 判斷是否需要處理批次
	batchSize, _ := strconv.Atoi(global.Envs.DCConfig.BatchSize)
	if len(dcPointBatch) >= batchSize {
		processDCBatch()
	}

	return nil
}

// processDCBatch 處理批次資料並寫入 InfluxDB
// 此函數已在 AddPoints 中上鎖，不需要重複加鎖
func processDCBatch() {
	if len(dcPointBatch) == 0 {
		return
	}

	pointsToProcess := make([]models.Point, len(dcPointBatch))
	copy(pointsToProcess, dcPointBatch)

	// 先嘗試寫入資料
	PointsToInfluxDB(pointsToProcess)

	// 寫入成功後才清空批次
	dcPointBatch = []models.Point{}
	dcLastBatchTime = time.Now()

}

// ProcessPendingBatch 供 crontab 調用的批次處理函數
func FlushInfluxDBBatch() error {
	dcPointMutex.Lock()
	defer dcPointMutex.Unlock()

	maxWaitTime, _ := strconv.Atoi(global.Envs.DCConfig.MaxWaitTime)
	if len(dcPointBatch) > 0 && time.Since(dcLastBatchTime) >= time.Duration(maxWaitTime)*time.Second {
		processDCBatch()

	}
	return nil
}
