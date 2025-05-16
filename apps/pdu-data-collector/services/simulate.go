package services

import (
	"bimap-zbox/global"
	"bimap-zbox/models"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// 發送模擬資料到工廠 endpoint 的函數，並附帶 API-KEY
func sendTelegrafDataToFactory(endpoint, apiKey string, data models.MetricsData) error {

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 構建 POST 請求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// 設置標頭，包括 API-KEY
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-KEY", apiKey)

	// 發送請求
	client := &http.Client{Timeout: time.Duration(global.EnvConfig.Simulate.Timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("非預期的狀態碼: %d", resp.StatusCode)
	}

	return nil
}

// 模擬生成 PDU 的 Metrics，根據 protocol 決定 fields 和 tags
func generatePDUMetrics(pduKey string, pduInfo map[string]string) models.Point {
	fields := make(map[string]float64)

	schema := global.Envs.PDUSchema

	// 模擬生成 fields 資料
	fields[schema.TotalCurrentField] = rand.Float64() * 100

	for i := 1; i <= 3; i++ {
		fields[fmt.Sprintf("current_L%d-1", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("current_L%d-2", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("current_L%d-3", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("energy_L%d-1", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("energy_L%d-2", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("energy_L%d-3", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("voltage_L%d", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("watt_L%d-1", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("watt_L%d-2", i)] = rand.Float64() * 100
		fields[fmt.Sprintf("watt_L%d-3", i)] = rand.Float64() * 100
	}

	// 根據 protocol 不同生成特有欄位
	if pduInfo["protocol"] == "snmp" {
		fields["current_L1"] = rand.Float64() * 100
		fields["current_L2"] = rand.Float64() * 100
		fields["current_L3"] = rand.Float64() * 100
		fields["watt_L1"] = rand.Float64() * 100
		fields["watt_L2"] = rand.Float64() * 100
		fields["watt_L3"] = rand.Float64() * 100
		fields[schema.TotalEnergyField] = rand.Float64() * 100
		fields[schema.TotalWattField] = rand.Float64() * 100
	}

	// 模擬 tags 資料
	tags := map[string]string{
		"pdu_key":  pduKey,
		"host":     "zbox",
		"model":    pduInfo["model"],
		"protocol": pduInfo["protocol"],
	}

	if pduInfo["protocol"] == "snmp" {
		tags["serial_number"] = "Y5R22700013WJ"
		tags["version"] = "01.12.12h"
		tags["source"] = pduInfo["ip"]
		tags["ip"] = pduInfo["ip"]
	} else if pduInfo["protocol"] == "modbus" {
		tags["ip"] = pduInfo["wifi_client_ip"]
		tags["port"] = pduInfo["wifi_client_port"]
	}

	// 返回 metrics 資料
	return models.Point{
		Name:   "pdu",
		Fields: fields,
		Tags:   tags,
		Time:   time.Now().Unix(),
	}
}

// 每 X 秒定時生成並發送 Y 台 PDU 模擬數據，限制併發數
func SimulateAndSendPDUData() {
	pduList := global.PDUList

	interval := time.Duration(global.EnvConfig.Simulate.Interval) * time.Second
	endpoint := global.EnvConfig.Simulate.URL
	for range time.Tick(interval) {
		var wg sync.WaitGroup
		sem := make(chan struct{}, global.EnvConfig.Simulate.MaxConcurrentRequests) // 控制併發量

		for pduKey, pduInfo := range *pduList {
			wg.Add(1)
			sem <- struct{}{} // 每次執行前佔用一個併發位

			go func(pduKey string, pduInfo map[string]string) {
				defer wg.Done()
				defer func() { <-sem }() // 任務完成後釋放併發位

				// 生成模擬數據
				telegrafData := models.MetricsData{}
				metric := generatePDUMetrics(pduKey, pduInfo)

				telegrafData.Metrics = append(telegrafData.Metrics, metric)

				// 發送模擬數據
				err := sendTelegrafDataToFactory(endpoint, global.Envs.GlobalConfig.LicenseKey, telegrafData)
				if err != nil {
					global.Logger.Error("模擬 Telegraf 發送資料時出錯", zap.Error(err))

				}
			}(pduKey, pduInfo)
		}

		wg.Wait()
		log.Println("已完成當前批次的所有 PDU 模擬數據發送")
	}
}
