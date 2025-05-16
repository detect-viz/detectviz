package factory

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"strconv"
	"sync"
	"time"
)

var (
	factoryPointBatch    []models.Point // 批次數據
	factoryPointMutex    sync.Mutex     // 保護批次數據的鎖
	factoryLastBatchTime time.Time      // 上次批次處理時間
)

func init() {
	factoryLastBatchTime = time.Now()
}

// AddPoints 添加新的數據點到批次中
func AddPoints(points []models.Point) error {
	if len(points) == 0 {
		return nil
	}

	factoryPointMutex.Lock()
	defer factoryPointMutex.Unlock()

	factoryPointBatch = append(factoryPointBatch, points...)

	// 判斷是否需要處理批次
	batchSize, _ := strconv.Atoi(global.Envs.FactoryConfig.BatchSize)
	if len(factoryPointBatch) >= batchSize {
		processFactoryBatch()
	}

	return nil
}

// processFactoryBatch 處理批次資料
func processFactoryBatch() {
	if len(factoryPointBatch) == 0 {
		return
	}

	formattedPoints := make([]models.Point, len(factoryPointBatch))
	copy(formattedPoints, factoryPointBatch)

	factoryPointBatch = nil
	factoryLastBatchTime = time.Now()

	FactoryMetricsToDC(formattedPoints)
}

// ProcessPendingBatch 供 crontab 調用的批次處理函數
func FlushDCBatch() error {
	factoryPointMutex.Lock()
	defer factoryPointMutex.Unlock()

	maxWaitTime, _ := strconv.Atoi(global.Envs.FactoryConfig.MaxWaitTime)
	if len(factoryPointBatch) > 0 && time.Since(factoryLastBatchTime) >= time.Duration(maxWaitTime)*time.Second {
		processFactoryBatch()
	}
	return nil
}
