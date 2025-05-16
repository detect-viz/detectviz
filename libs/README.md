
```bash
libs/
├── alert/             # 告警邏輯（計算、分類、狀態）
│   ├── model.go
│   ├── engine.go
│   └── severity.go
│
├── api/               # 通用 API 工具
│   ├── middleware/
│   ├── errors/
│   └── response/
│
├── auth/              # 驗證與使用者管理
│   └── keycloak/
│
├── config/            # 設定載入與分層配置管理
│   ├── loader/
│   ├── manager/
│   └── interfaces/
│
├── contacts/          # 通知對象資料邏輯
│
├── infra/             # 系統基礎建設工具
│   ├── archiver/
│   ├── logger/
│   └── scheduler/
│
├── labels/            # Label CRUD 模組
│
├── licensing/         # 授權與 license 控管
│
├── mutes/             # 告警靜音規則模組
│
├── notifier/          # 發送通知的核心模組
│
├── plugins/           # Plugin 架構與插件註冊
│   ├── manager.go
│   ├── inputs/
│   ├── outputs/
│   └── parsers/
│
├── rules/             # Rule CRUD
│
├── storage/           # 儲存介面統一與實作
│   ├── interface.go      # interface 定義
│   ├── mysql/
│   └── influxdb/
│
└── templates/         # 通知與報表樣板處理
    └── service.go
```


## 模組補充說明

---

## 🛠 子模組結構優化建議

目前 `libs/` 目錄下部分子模組已具備內部檔案（如 `model.go`, `service.go`），但大多尚未統一格式。為提升可維護性與一致性，建議每個子模組採用以下結構：

```bash
libs/模組名稱/
├── model.go       # 結構定義（若該模組有 struct）
├── service.go     # 該模組的主要對外邏輯（Handler / Processor）
├── interface.go   # interface（若該模組需被 mock / 抽象）
├── errors.go      # 錯誤處理（可選）
├── config.go      # 該模組內部配置（可選）
└── README.md      # 簡要說明模組用途與對外接口
```

> 若該模組較大，可再加子目錄如 `internal/`, `handler/`, `util/` 等。


建議優先調整目標模組：
- `contacts/`（目前為空）
- `labels/`（應補上 `model.go`, `service.go`）
- `notifier/`（建議補上 `interface.go`, `channel/` 子模組）
- `alert/`, `rules/`（若存在重複邏輯，應依責任劃分整併）

調整後，每個 libs 子模組都應可獨立作為套件使用與測試，提升可讀性與團隊協作效率。

---

## 📌 各模組應實現的功能與檢查重點

以下為各 `libs/` 子模組應具備的功能與建議檢查項目，供重構與檢查時參考：

### `contacts/`
- 功能：管理告警聯絡人、聯絡群組、通道對應（email/LINE）
- 應實作：
  - `model.go`：Contact, Group 結構
  - `service.go`：CRUD 與搜尋介面
  - 建議有：`FindByID`, `FindByChannel`, `ListAll`

### `labels/`
- 功能：定義通用標籤（如設備分類、等級分類）
- 應實作：
  - `model.go`：Label 結構（含 name, key, value）
  - `service.go`：CRUD + 搜尋 by type
  - 適用於 event 分類、報表統計欄位

### `notifier/`
- 功能：根據 alert 結果發送多通道通知
- 應實作：
  - `interface.go`：`type Sender interface { Send(ctx, msg) error }`
  - `channel/`：子目錄每通道實作（email.go, line.go）
  - `service.go`：主調度器，負責根據 `contacts` 路由訊息

### `alert/`
- 功能：告警判斷與狀態管理
- 應實作：
  - `model.go`：Alert 結構（包含 id, type, severity, state, source）
  - `engine.go`：邏輯判定（如 CompareThreshold, WithinRange）
  - `state.go`：狀態流轉（new → active → resolved）

### `rules/`
- 功能：CRUD 操作的告警條件設定（動態規則）
- 應實作：
  - `model.go`：Rule 結構（name, target, condition, threshold）
  - `service.go`：CRUD + 查詢（By target type / field）
  - 若需驗證條件語法，建議加入 `validator.go`

### `storage/`
- 功能：提供資料儲存介面，統一封裝 InfluxDB/MySQL 等
- 應實作：
  - `interface.go`：`type Store interface`
  - `influxdb/`, `mysql/`：各自實作 Store
  - `NewStore(config)` 工廠函數，依環境建立對應儲存器

### `plugins/`
- 功能：定義可插拔模組（inputs/outputs/parsers）
- 應實作：
  - `manager.go`：PluginRegistry 註冊/載入
  - 每 plugin 需實作 `Plugin` interface（Start / Stop / Status）
  - 註冊方式可參考 map + init 時自動註冊

---

每個模組建議補上 `README.md` 文件，簡要描述模組責任、interface 用法與對外依賴。Cursor 在檢查時可依據此清單快速確認是否已具備必要檔案與實作內容。