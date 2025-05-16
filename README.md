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
.
├── apps
│   ├── pdu-data-collector
│   │   ├── config.example.yml
│   │   ├── controller
│   │   │   ├── dc.go
│   │   │   └── factory.go
│   │   ├── databases
│   │   │   ├── influxdb.go
│   │   │   ├── migrate.go
│   │   │   ├── mssql.go
│   │   │   └── mysql.go
│   │   ├── env_data.yml
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── job_data.yml
│   │   ├── log_data.yml
│   │   ├── migrations
│   │   │   ├── 000006_feed_envs_table.up.sql
│   │   │   ├── 000008_feed_jobs_table.up.sql
│   │   │   └── 000010_feed_logs_table.up.sql
│   │   ├── models
│   │   │   ├── env.go
│   │   │   ├── gorm.go
│   │   │   ├── job.go
│   │   │   └── log.go
│   │   ├── pdu.csv
│   │   ├── README.md
│   │   ├── rule.sql
│   │   ├── services
│   │   │   ├── dc
│   │   │   ├── demo.go
│   │   │   ├── demoInsert.go
│   │   │   ├── factory
│   │   │   ├── general.go
│   │   │   ├── simulate.go
│   │   │   └── testApi.go
│   │   └── 軟體架構.jpg
│   ├── README.md
│   ├── viot-dc
│   │   └── README.md
│   ├── viot-fab
│   │   └── README.md
│   ├── viot-README.md
│   └── website
│       ├── cmd
│       │   └── main.go
│       ├── data
│       │   ├── main.log
│       │   ├── registry.csv
│       │   ├── scan.csv
│       │   ├── status_DC1.csv
│       │   ├── status_DC2.csv
│       │   ├── status.csv
│       │   └── tag.csv
│       ├── go.mod
│       ├── go.sum
│       ├── index.html
│       ├── layout.html
│       ├── old-config
│       │   ├── config.yaml
│       │   ├── devices.yaml
│       │   ├── fping-5.3.md
│       │   ├── ip_range.yaml
│       │   ├── pdu_list.csv
│       │   ├── position
│       │   ├── processor
│       │   ├── scripts
│       │   ├── settings
│       │   ├── task
│       │   ├── telegraf.conf
│       │   ├── templates
│       │   └── yamls
│       ├── README.md
│       └── static
│           ├── css
│           ├── fonts
│           ├── images
│           └── js
├── conf
│   ├── alert-page.md
│   ├── config
│   │   ├── cleanup.sql
│   │   ├── conf.d
│   │   │   ├── code.yaml
│   │   │   ├── metric_rule.yaml
│   │   │   ├── metric.yaml
│   │   │   ├── tag.yaml
│   │   │   ├── template-default-alerting.yaml
│   │   │   └── template-default-resolved.yaml
│   │   ├── config.yaml
│   │   ├── docs
│   │   │   ├── 000003_create_metric_rules_table.up.sql
│   │   │   ├── alert_contacts.yaml
│   │   │   ├── alert_rules.yaml
│   │   │   ├── constants.md
│   │   │   ├── init.sql
│   │   │   ├── main_menus_table.up.sql
│   │   │   ├── menus_table.up.sql
│   │   │   ├── metrics.csv
│   │   │   ├── migrations
│   │   │   ├── migrations_2
│   │   │   ├── parser_awrrpt_files.sql
│   │   │   ├── rule.sql
│   │   │   ├── scheduler.yaml
│   │   │   ├── Spec.md
│   │   │   └── wire.md
│   │   ├── m.yaml
│   │   ├── mi_insert.sql
│   │   ├── migrations
│   │   │   ├── 000001_create_targets_table.down.sql
│   │   │   ├── 000001_create_targets_table.up.sql
│   │   │   ├── 000002_create_contacts_table.down.sql
│   │   │   ├── 000002_create_contacts_table.up.sql
│   │   │   ├── 000003_create_templates_table.down.sql
│   │   │   ├── 000003_create_templates_table.up.sql
│   │   │   ├── 000004_create_rules_table.down.sql
│   │   │   ├── 000004_create_rules_table.up.sql
│   │   │   ├── 000005_create_rule_contacts_table.down.sql
│   │   │   ├── 000005_create_rule_contacts_table.up.sql
│   │   │   ├── 000006_create_rule_states_table.down.sql
│   │   │   ├── 000006_create_rule_states_table.up.sql
│   │   │   ├── 000007_create_triggered_logs_table.down.sql
│   │   │   ├── 000007_create_triggered_logs_table.up.sql
│   │   │   ├── 000008_create_notify_logs_table.down.sql
│   │   │   └── 000008_create_notify_logs_table.up.sql
│   │   ├── oracle-monitor-script
│   │   │   ├── all.sql
│   │   │   ├── connection
│   │   │   ├── README.md
│   │   │   └── tablespace
│   │   ├── pdu_list.numbers
│   │   ├── provisioning
│   │   │   ├── dashboards
│   │   │   └── notifiers
│   │   ├── test
│   │   │   ├── CPU-001_alert.json
│   │   │   ├── CPU-001_normal.json
│   │   │   ├── CPU-002_alert.json
│   │   │   ├── CPU-002_normal.json
│   │   │   ├── CPU-003_alert.json
│   │   │   ├── CPU-003_normal.json
│   │   │   ├── DATABASE-001_alert.json
│   │   │   ├── DATABASE-001_normal.json
│   │   │   ├── DISK-001_alert.json
│   │   │   ├── DISK-001_normal.json
│   │   │   ├── FILESYSTEM-001_alert.json
│   │   │   ├── FILESYSTEM-001_normal.json
│   │   │   ├── FILESYSTEM-002_alert.json
│   │   │   ├── FILESYSTEM-002_normal.json
│   │   │   ├── MEMORY-001_alert.json
│   │   │   ├── MEMORY-001_normal.json
│   │   │   ├── MEMORY-002_alert.json
│   │   │   ├── MEMORY-002_normal.json
│   │   │   ├── NETWORK-001_alert.json
│   │   │   ├── NETWORK-001_normal.json
│   │   │   ├── README.md
│   │   │   ├── TABLESPACE-001_alert.json
│   │   │   └── TABLESPACE-001_normal.json
│   │   └── 功能保留紀錄.md
│   ├── custom.yaml
│   ├── default.ini
│   ├── provisioning
│   │   └── analytics
│   │       ├── panels.yaml
│   │       ├── profiles.yaml
│   │       └── rules.yaml
│   ├── secret.ini
│   └── settings.json
├── docker-compose.yml
├── libs
│   ├── alert
│   │   ├── alert.go
│   │   ├── interface.go
│   │   ├── monitor.go
│   │   ├── notify.go
│   │   ├── README.md
│   │   ├── tools.go
│   │   ├── wire_gen.go
│   │   └── wire.go
│   ├── api
│   │   ├── controller
│   │   │   ├── alert_page.go
│   │   │   ├── alert.go
│   │   │   ├── contact.go
│   │   │   ├── process.go
│   │   │   └── rule.go
│   │   ├── errors
│   │   │   └── error.go
│   │   ├── middleware
│   │   │   └── middleware.go
│   │   ├── response
│   │   │   └── responce.go
│   │   └── router.go
│   ├── auth
│   │   ├── interface.go
│   │   ├── keycloak
│   │   │   ├── client.go
│   │   │   ├── interface.go
│   │   │   ├── login.go
│   │   │   ├── provider.go
│   │   │   ├── realm.go
│   │   │   └── user.go
│   │   ├── middleware
│   │   │   └── auth.go
│   │   └── README.md
│   ├── config
│   │   ├── interfaces
│   │   │   └── config.go
│   │   ├── loader
│   │   │   └── loader.go
│   │   ├── manager
│   │   │   └── manager.go
│   │   ├── models
│   │   │   └── config.go
│   │   └── README.md
│   ├── contacts
│   │   ├── interface.go
│   │   └── service.go
│   ├── infra
│   │   ├── archiver
│   │   │   ├── backup.go
│   │   │   ├── interface.go
│   │   │   ├── rotate.go
│   │   │   └── service.go
│   │   ├── logger
│   │   │   ├── interface.go
│   │   │   ├── README.md
│   │   │   └── service.go
│   │   └── scheduler
│   │       ├── interface.go
│   │       └── service.go
│   ├── labels
│   │   ├── interface.go
│   │   └── service.go
│   ├── licensing
│   │   ├── interfaces
│   │   │   └── license.go
│   │   ├── manager
│   │   │   └── manager.go
│   │   ├── mock
│   │   │   └── mock.go
│   │   ├── models
│   │   │   └── license.go
│   │   └── README.md
│   ├── main.go
│   ├── models
│   │   ├── alert
│   │   │   ├── config.go
│   │   │   ├── contact.go
│   │   │   ├── metric_rule.go
│   │   │   ├── notify_log.go
│   │   │   ├── payload.go
│   │   │   ├── rule_state.go
│   │   │   ├── rule.go
│   │   │   ├── snapshot.go
│   │   │   ├── target.go
│   │   │   ├── template.go
│   │   │   └── triggered_log.go
│   │   ├── common
│   │   │   ├── archiver.go
│   │   │   ├── gorm.go
│   │   │   ├── notifier.go
│   │   │   ├── response.go
│   │   │   ├── task.go
│   │   │   └── user.go
│   │   ├── config
│   │   │   ├── alert.go
│   │   │   ├── auth.go
│   │   │   ├── config.go
│   │   │   ├── database.go
│   │   │   ├── logger.go
│   │   │   ├── parser.go
│   │   │   └── server.go
│   │   ├── dto
│   │   │   └── label.go
│   │   ├── label
│   │   │   └── model.go
│   │   ├── logger
│   │   │   └── logger.go
│   │   ├── models.go
│   │   ├── mute
│   │   │   └── mute.go
│   │   ├── notifier
│   │   │   └── channel.go
│   │   ├── parser
│   │   │   ├── file.go
│   │   │   └── metric.go
│   │   ├── resource
│   │   │   └── resource.go
│   │   ├── scheduler
│   │   │   └── job.go
│   │   └── template
│   │       └── data.go
│   ├── mutes
│   │   ├── interface.go
│   │   └── service.go
│   ├── notifier
│   │   ├── email.go
│   │   ├── errors
│   │   │   └── errors.go
│   │   ├── interface.go
│   │   ├── README.md
│   │   ├── service.go
│   │   ├── utils
│   │   │   ├── http.go
│   │   │   ├── time.go
│   │   │   └── utils.go
│   │   ├── validate
│   │   │   ├── common.go
│   │   │   ├── email.go
│   │   │   ├── line.go
│   │   │   ├── url.go
│   │   │   ├── validator.go
│   │   │   └── webhook.go
│   │   └── webhook.go
│   ├── plugins
│   │   ├── inputs
│   │   ├── manager.go
│   │   ├── outputs
│   │   └── parsers
│   │       └── interface.go
│   ├── README.md
│   ├── rules
│   │   ├── interface.go
│   │   └── service.go
│   ├── storage
│   │   ├── influxdb
│   │   │   ├── interfaces
│   │   │   ├── v2
│   │   │   ├── v3
│   │   │   └── wire.go
│   │   └── mysql
│   │       ├── alert_notify_log.go
│   │       ├── alert_rule_state.go
│   │       ├── alert_triggered_log.go
│   │       ├── alert.go
│   │       ├── cleanup.go
│   │       ├── contact.go
│   │       ├── error.go
│   │       ├── gorm.go
│   │       ├── interface.go
│   │       ├── label.go
│   │       ├── migrate.go
│   │       ├── mute.go
│   │       ├── query.go
│   │       ├── rule.go
│   │       ├── service.go
│   │       ├── target.go
│   │       └── template.go
│   └── templates
│       ├── interface.go
│       ├── README.md
│       └── service.go
├── Makefile
├── orchestrator
│   └── config
│       └── pages.yaml
├── README.md
├── scripts
│   └── auto-depoly-shell
├── services
│   ├── accesscontrol-service
│   ├── alert-service
│   ├── analytics-service
│   │   ├── engine
│   │   │   ├── api.py
│   │   │   ├── cli.py
│   │   │   ├── detector
│   │   │   ├── main.py
│   │   │   ├── README.md
│   │   │   └── requirements.txt
│   │   ├── entities
│   │   │   ├── metric.go
│   │   │   ├── metrics.go
│   │   │   ├── prompt.go
│   │   │   ├── report.go
│   │   │   ├── rule.go
│   │   │   └── types.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── internal
│   │   │   ├── analyzer
│   │   │   ├── config
│   │   │   ├── processor
│   │   │   └── reporter
│   │   ├── main.go
│   │   ├── middleware
│   │   │   └── middleware.go
│   │   ├── README.md
│   │   └── spc-shell

│   ├── automation-service
│   ├── collector-service
│   ├── healthcheck-service
│   ├── llm-service
│   ├── notifier-service
│   └── report-service

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

---

## 📘 模組補充說明

### `rules/` 與 `labels/`

這兩個模組主要提供共用的 CRUD API 工具：

- `rules/`：提供告警規則（Rule）定義的資料操作邏輯與查詢介面，支援 alert-service 與前端設定 UI 使用。
- `labels/`：管理可配置的標籤（Label）分類，用於事件過濾、模組歸類或多維統計條件，支援通用查詢與 CRUD。

這些模組不負責核心運算邏輯，而是提供彈性設定與 metadata 管理能力，適合作為告警模組（alert-service）與報表模組的擴展支援元件。
