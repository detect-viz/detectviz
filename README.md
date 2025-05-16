# Detectviz

Detectviz 是一套模組化、可擴展的多場域部署監控平台，採用 Go monorepo 架構，整合電力監控、自動部署、資料收集、異常偵測、告警通知與報表輸出，支援 Data Center（DC）與現場 Fab 雙模態，並能根據不同場景需求組裝微服務，快速交付完整解決方案。

⸻

## 專案目標

## 📎 關於 viot 應用

`viot` 是 Detectviz 生態系下的一個應用（app），專注於電力監控與自動部署任務。它運用 Detectviz 提供的微服務與共用函式庫組合成具體可交付的專案方案，例如 `apps/viot-fab` 和 `apps/viot-dc`。

請注意，Detectviz 是整體產品平台的核心，若 `viot` 中有與特定商業客戶或應用場景強綁的邏輯（如報價單整合、特殊 UI 流程等），應當**獨立撰寫於 `apps/viot-*` 中實作**，而不應混入核心 microservices 或 libs，確保平台可長期維護與擴展。
	•	統一多場域異質設備的監控、自動化部署與告警流程
	•	提供 CLI + Web UI 雙介面，支援現場/中控操作
	•	以模組化微服務構成，可動態擴充 / 熱插拔功能
	•	實踐以「配置驅動任務」、「模式驅動場景」的 SaaS 架構平台

⸻

## 架構總覽：Go Monorepo

本專案遵循 go-monorepo 模式，依據功能責任劃分為：
	•	apps/：具體應用方案（如 viot-dc, viot-fab, website）
	•	services/：可獨立部署的微服務（如 alert, analytics, notifier）
	•	libs/：跨模組共用的核心邏輯（如 alert 判斷、欄位轉換、config 管理）
	•	conf/：全域設定與環境參數（.env, secrets, schema）
	•	orchestrator/：平台調度總控，可讀取 pages.yaml 管理任務模組組合

⸻

🔧 開發流程（建議順序）

階段	模組	說明
①	libs/	建立共用邏輯（alert, transform, config）
②	services/collector-service	整合掃描器 / 部署器（原 viot 功能）
③	services/alert-service, analytics-service	整合異常偵測邏輯、SPC 分析
④	notifier-service, automation-service	事件通知、shell 自動修復
⑤	apps/website/	提供 HTMX 前端管理 UI
⑥	apps/viot-fab, viot-dc	場域特定部署方案，透過 --mode 控制
⑦	Makefile, scripts/	建構 / 測試 / CI/CD 自動化流程


⸻

## 模組分工一覽

類別	目錄	說明
解決方案	apps/viot-*	實際商業部署組合應用
管理介面	apps/website/	Web 任務控制中心
核心微服務	services/	單一職責服務，可獨立部署
共用函式庫	libs/	alert 判斷、config loader、欄位轉換等
編排調度	orchestrator/	任務對應、模組組合、中控執行主體


⸻

🚀 啟動與測試方式

# 啟動 viot-fab 應用（支援 --mode=fab/dc）
cd apps/viot-fab
go run main.go --mode=fab

# 啟動單一微服務（如 collector）
cd services/collector-service
go run main.go

# 使用 docker-compose 啟動全套服務
make up


⸻

📂 目錄結構總覽

```bash
detectviz/
├── apps/
│   ├── viot-fab/
│   ├── viot-dc/
│   └── website/
├── services/
│   ├── accesscontrol-service
│   ├── alert-service
│   ├── analytics-service
│   ├── anomaly-service
│   ├── automation-service
│   ├── collector-service
│   ├── healthcheck-service
│   ├── llm-service
│   ├── notifier-service
│   ├── report-service
│   └── visual-service
├── libs/
│   ├── alert/
│   ├── transform/
│   ├── config/
│   └── db/
├── orchestrator/
│   └── config/pages.yaml
├── conf/
│   ├── .env.local
│   ├── secrets.toml
│   └── migrations/
├── scripts/
├── Makefile
├── docker-compose.yml
└── README.md
```

---

## 📁 根目錄說明（功能導向）

| 目錄 | 說明 |
|------|------|
| `apps/` | 具體應用實作（如 viot-fab / viot-dc），是由各微服務組裝而成的場景解決方案 |
| `services/` | 各微服務模組，具有單一職責，如資料收集、異常分析、通知等 |
| `libs/` | 共用邏輯與工具庫，供 services 與 apps 引用，如告警分類、欄位轉換、config loader 等 |
| `orchestrator/` | 平台調度總控邏輯，負責解析 `pages.yaml` 組合模組頁面並呈現在 Web UI 上 |
| `conf/` | 儲存全域設定檔、機密、環境變數（如 `.env`, `secrets.toml`, DB migration） |
| `scripts/` | 常用腳本與 DevOps 輔助工具，可搭配 Makefile 使用 |
| `Makefile` | 頂層建構與執行腳本整合入口 |
| `docker-compose.yml` | 快速啟動本地整合環境 |
| `README.md` | 本說明文件 |
