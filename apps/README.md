# Viot 機房監控管理平台

Viot 是 Detectviz 平台下的一個應用（App），專為工廠、數據中心等機房環境設計，具備收值部署、異常偵測、模組組合、視覺化整合等功能，採用 Go monorepo 架構，各應用相互獨立但可透過 API 組合，建構中控 + 多場域的模組化監控平台。。

---

## 架構簡介

本平台由兩大角色構成：

- **Fab 資料中心**（21 套部署）：
  - 使用 Telegraf 收集電力/設備數據（如電流、電壓、功率）
  - 儲存至 InfluxDB
  - 並透過 analytics-service 進行異常判定

- **DC 中控中心**（1 套部署）：
	- 整合 21 個 Fab 的狀態與事件
	- 提供視覺化儀表板管理（連接 Grafana）
	- 儲存場域異常與操作事件至 MySQL
	- 提供 API 給 Fab 回傳事件（例如 zap hook 回報）


## Fab 回傳機制
每個 Fab 模組操作（如：部署、收值啟動、異常觸發）都會透過 HTTP `POST /api/event-log` 將事件資訊送至中控。
中控將事件紀錄寫入 MySQL 資料庫，作為告警呈現與稽核分析依據。

範例 JSON：

```json
{
  "fab_code": "F12P7DC1R3",
  "level": "error",
  "type": "collector",
  "message": "收值失敗，port 161 拒絕連線",
  "timestamp": 1746640000
}
```


---

## 引用模組（來自 Detectviz 平台）

### 來自 `detectviz/services/` 的微服務：

| 模組 | 功能 |
|------|------|
| `collector-service` | 掃描設備、產生 Telegraf 設定、部署與監控收值狀態 |
| `analytics-service` | 包含 Prophet、MovingAvg、Threshold 等多種異常分析模型 |
| `alert-service` | 處理告警事件、異常紀錄與等級判定 |
| `notifier-service` | 推送通知至 LINE / Email 等通道 |
| `automation-service` | 可自動觸發 shell 修復腳本等操作（選用） |

### 來自 `detectviz/libs/` 的共用函式庫：

| 套件 | 功能說明 |
|------|----------|
| `libs/transform` | 將原始欄位標準化為統一格式（如電流 current, 電壓 voltage） |
| `libs/alert` | 判斷告警等級、去重處理、狀態標記 |
| `libs/config` | 載入 config.toml, .env，並提供模組組合邏輯 |
| `libs/db` | 提供對 InfluxDB/MySQL/Redis 的封裝操作 |

---

## 📁 目錄結構說明（應用分工）

```
apps/
├── viot-fab/     # Fab 場域端應用：負責設備掃描、部署、收值與異常分析
├── viot-dc/      # 中控端應用：負責集中告警、事件日誌、儀表板整合與管理
├── website/      # 管理介面 Web UI：與 viot-dc 對接，呈現模組控制與場域狀態
```

---

## 📦 適用場景

- 多場域（如分廠）電力監控整合管理
- 機房級設備監控自動部署與告警聯動
- 以 configuration 為中心，模組自由組合
- 提供中控與場域雙層監控結構設計

---

本 App 為 Detectviz 平台應用方案之一，若需擴展其他場景（如 IoT 製程監控、品質統計 SPC），可獨立建立對應 App 組合模組。