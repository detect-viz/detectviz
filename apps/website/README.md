# VIOT Platform

一個基於 Echo + HTMX + a‑h/templ 打造的現代化監控管理平台。本專案採用完全配置驅動的開發模式，專注於高效能、離線部署與模組化設計。

## 🌟 平台特色

- **配置驅動**：所有頁面與功能皆由 `config/pages.yaml` 控制，支援 CRUD、腳本執行與 Grafana 整合。
- **互動式介面**：搭配 HTMX 實現無刷新的按鈕、表單、對話框與條件顯示。
- **資料來源多元**：支援 CSV、ENV、YAML、LOG 等格式，並可嵌入 Grafana iframe。
- **元件化 UI**：使用 `a-h/templ` 架構，介面元件獨立，可複用、易維護。
- **無 JS 框架依賴**：介面互動純粹由 HTMX 控制，無需 Vue 或 React。
- **完全離線化**：所有靜態資源皆已本地化，適用於內網與封閉環境。

## 🔐 認證與授權

### 登入機制
- 使用 Cookie-based Session 認證
- 預設帳號密碼由環境變數 `LOGIN_USER` 和 `LOGIN_PASS` 設定
- Session 有效期由 `COOKIE_MAX_AGE` 控制（預設 7 天）
- 所有頁面（除登入頁外）皆需驗證

### 登入流程
1. 訪問任何頁面時，系統檢查 `session` cookie
2. 若未登入，自動重導向至 `/login` 頁面
3. 輸入正確帳密後，設置 `session=ok` cookie
4. 重導向回原本要訪問的頁面

### 登出機制
- 點擊右上角登出按鈕
- 清除 `session` cookie
- 重導向至登入頁面


## 🎨 CSS 與 Icon 使用規範

### ✅ CSS 樣式規範

| 項目 | 規範說明 |
|------|----------|
| 框架來源 | 使用 AdminLTE 4.0 Beta3 的純 CSS，已下載離線版 |
| 組件樣式 | 使用 `card`, `badge`, `table`, `form-select-sm`, `btn-outline-*` 等標準 class |
| 響應式設計 | 保留 AdminLTE 的 `container-fluid`, `row`, `col-*` 架構 |
| 禁用 JS 動效 | 不使用 collapsible、modal、treeview 等需 JS 的元件 |

### ✅ Icon 圖示規範

| 項目 | 規範說明 |
|------|----------|
| 圖示來源 | Material Icons（離線字體包） |
| 使用方式 | `<i class="material-icons">icon_name</i>` |
| 圖示風格 | 預設使用 filled，可選 outlined/round/sharp/two-tone |
| 顏色控制 | 透過 style 或類似 tailwind 類別控制顏色（如 `text-muted`, `text-danger`） |
| 大小控制 | 使用內建 class `md-18 / md-24 / md-36` 控制大小，不寫 inline CSS |
| 側邊欄高亮 | icon 搭配 `PageConfig.Color`，作為選單分群與高亮依據 |

### ✅ AdminLTE 4.0 Beta3 參考檔案

#### 核心頁面參考
| 頁面類型 | 參考檔案 | 說明 |
|----------|----------|------|
| 登入頁面 | `docs/adminlte-reference/login-v2.html` | 登入表單與驗證 |
| 主控台 | `docs/adminlte-reference/index.html` | 整體布局與側邊欄 |
| 表格頁面 | `docs/adminlte-reference/simple.html` | 數據表格與分頁 |
| 表單頁面 | `docs/adminlte-reference/general.html` | 表單元素與驗證 |
| 狀態頁面 | `docs/adminlte-reference/general.html` | 狀態顯示與圖示 |

#### 組件參考
| 組件類型 | 參考檔案 | 說明 |
|----------|----------|------|
| 導航欄 | `docs/adminlte-reference/fixed-sidebar.html` | 頂部導航與用戶選單 |
| 側邊欄 | `docs/adminlte-reference/sidebar-mini.html` | 側邊欄選單與折疊 |
| 卡片 | `docs/adminlte-reference/cards.html` | 卡片布局與樣式 |
| 資訊卡片 | `docs/adminlte-reference/info-box.html` | 資訊展示卡片 |
| 小型卡片 | `docs/adminlte-reference/small-box.html` | 統計數據卡片 |
| 按鈕 | `docs/adminlte-reference/general.html` | 按鈕樣式與狀態 |
| 表格 | `docs/adminlte-reference/simple.html` | 表格樣式與響應式 |
| 表單 | `docs/adminlte-reference/general.html` | 表單元素與布局 |
| 對話框 | `docs/adminlte-reference/general.html` | 模態框與提示 |
| 標籤 | `docs/adminlte-reference/general.html` | 標籤與徽章樣式 |

#### 轉換注意事項
1. **布局結構**
   - 使用 `wrapper` 作為最外層容器
   - 側邊欄使用 `sidebar` 類別
   - 主內容區使用 `content-wrapper`

2. **組件轉換**
   - 將 HTML class 轉換為 templ 屬性
   - 使用條件渲染替代 JavaScript 動態效果
   - 保持響應式設計的 class 結構

3. **互動處理**
   - 使用 HTMX 替代 JavaScript 事件
   - 表單提交使用 `hx-post`
   - 動態加載使用 `hx-get`

4. **樣式保持**
   - 保留 AdminLTE 的 CSS 類別
   - 使用 Material Icons 替代 Font Awesome
   - 保持一致的顏色主題

## 📁 專案目錄

```bash
viot/
├── cmd/                    # 主程式入口
│   └── main.go            # 程式進入點
├── internal/              # 內部套件
│   ├── config/           # 配置相關
│   │   ├── config.go    # 配置結構定義
│   │   └── loader.go    # YAML 載入邏輯
│   ├── handler/         # 請求處理器
│   │   ├── auth.go     # 認證相關（登入/登出/驗證）
│   │   ├── csv.go      # CSV 頁面處理
│   │   ├── files.go    # 檔案頁面處理
│   │   ├── exec.go     # 工具頁面處理
│   │   ├── log.go      # 日誌頁面處理
│   │   └── status.go   # 狀態圖處理
│   ├── middleware/      # 中間件
│   │   ├── auth.go     # 認證中間件（Session 檢查）
│   │   └── logger.go   # 日誌中間件
│   ├── model/          # 資料模型
│   │   ├── page.go    # 頁面相關結構
│   │   └── user.go    # 使用者相關結構
│   ├── service/        # 業務邏輯
│   │   ├── auth.go    # 認證邏輯實作
│   │   ├── csv.go     # CSV 操作邏輯
│   │   ├── files.go     # 檔案操作邏輯
│   │   └── exec.go    # 工具執行邏輯
│   └── view/           # Templ 元件
│       ├── components/ # 共用元件
│       │   ├── layout.templ    # 主布局
│       │   ├── navbar.templ    # 導航欄（含登出按鈕）
│       │   ├── sidebar.templ   # 側邊欄
│       │   ├── table.templ     # 表格組件
│       │   ├── form.templ      # 表單組件
│       │   ├── modal.templ     # 對話框
│       │   └── status.templ    # 狀態圖組件
│       └── pages/      # 頁面元件
│           ├── login.templ     # 登入頁面
│           ├── csv.templ      # CSV 頁面
│           ├── file.templ      # 檔案頁面
│           ├── exec.templ     # 工具頁面
│           ├── log.templ      # 日誌頁面
│           └── status.templ   # 狀態圖頁面
├── config/              # 配置文件
│   └── pages.yaml      # 頁面配置
├── data/               # 數據文件
│   ├── .env           # 環境變量（含登入設定）
│   ├── main.log       # 系統日誌
│   ├── registry.csv   # 設備清單
│   ├── scan.csv       # 掃描結果
│   ├── status.csv     # 設備狀態
│   └── tag.csv        # 設備標籤
├── scripts/           # 執行腳本
├── static/           # 靜態資源
│   ├── css/         # 樣式文件
│   │   ├── adminlte.min.css
│   │   ├── material-icons.css
│   │   └── style.css
│   ├── js/          # JavaScript
│   │   ├── adminlte.min.js
│   │   ├── bootstrap.bundle.min.js
│   │   └── htmx.min.js
│   └── fonts/       # 字體文件
│       ├── NotoSansTC-*.ttf
│       └── material-icons.*
├── go.mod           # Go 模組文件
├── go.sum           # Go 依賴版本
└── README.md        # 專案說明
```

## 🚀 快速開始

1. **安裝依賴**
   ```bash
   git clone <repository-url>
   cd viot-web
   go mod tidy
   ```

2. **環境配置**

```bash
   # 登入驗證設定
   LOGIN_USER=admin
   LOGIN_PASS=admin
   
   # Session 相關
   COOKIE_MAX_AGE=604800
   
   # 系統服務設定
   PORT=8080
   LOG_LEVEL=INFO
   GRAFANA_URL=http://localhost:3000
   
   # 設定檔路徑
   PAGE_CONFIG_PATH=config/pages.yaml
   
   # 下拉選單
   SET_FACTORY_PHASE=F12:P7,F12:P8,F14:P6     # Factory + Phase 組合
   SET_DC_OF_PHASE=F12P7:2,F12P8:1,F14P6:1     # 每個 FactoryPhase 有幾個 DC
   SET_ROOM_OF_DC=F12P7DC1:8,F12P7DC2:4,F12P8DC1:1,F14P6DC1:2
   
   # 標籤顏色配置
   LEVEL_NORMAL_BACKGROUND_COLOR=#28a745       # 正常 (status_code=0)
   LEVEL_WARN_BACKGROUND_COLOR=#ffc107         # 警告 (status_code=1)
   LEVEL_ERROR_BACKGROUND_COLOR=#dc3545        # 異常 (status_code=2)
   
   # PDU平面圖顏色配置
   PDU_EMPTY_BACKGROUND_COLOR=#f8f9fa         # 空櫃底色 (is_pdu=0)
   PDU_EMPTY_TEXT_COLOR=#6c757d               # 空櫃字體色 (is_pdu=0)
   PDU_DEFAULT_BACKGROUND_COLOR=#dee2e6       # 一般底色 (is_pdu=1)
   PDU_DEFAULT_TEXT_COLOR=#343a40             # 一般字體色 (is_pdu=1)
```

3. **編譯運行**
   ```bash
   # 生成 templ
   templ generate
   
   # 運行服務
   go run main.go
   ```

## 🧾 config/pages.yaml 格式與說明

每個頁面皆由一筆 `PageConfig` 組成，支援以下欄位：

### 頁面共用欄位定義
| 欄位           | 類型         | 說明                                 |
|----------------|--------------|--------------------------------------|
| name           | string       | 路由與識別名稱（唯一）              |
| label          | string       | 左側選單與頁面標題顯示               |
| icon           | string       | Material Icons 圖示名稱              |
| color          | string       | icon 文字顏色（如 text-primary）     |
| type           | string       | 頁面類型（csv/files/log/status/exec/panels/dashboard） |
| path           | string       | 對應資料來源（CSV 路徑或 log） |
| fields         | []Field      | 欄位清單（適用於 csv/status）   |
| actions        | []Action     | 工具按鈕定義（僅 type=exec）         |
| url            | string       | 嵌入面板網址（僅 type=dashboard）    |
| panels         | []Panel      | 多面板嵌入（僅 type=panels）         |
| layout         | Layout       | panels 類型時控制 Grid 佈局         |

### Field 欄位格式（fields[]）
| 欄位         | 類型     | 說明                            |
|--------------|----------|---------------------------------|
| name         | string   | 欄位名稱（對應資料來源欄位）     |
| label        | string   | 顯示用中文欄位名                 |
| type         | string   | 可為 text/hidden（選填）         |
| readonly     | bool     | 是否唯讀                         |
| enum         | []string | 下拉選單選項（如 [true, false]） |

### Action 欄位格式（actions[]，僅 type=exec）
| 欄位         | 說明                            |
|--------------|---------------------------------|
| label        | 顯示的按鈕名稱                  |
| script_path  | 要執行的 shell 指令路徑         |
| icon         | Material Icons 圖示             |
| desc         | 描述說明（顯示於 hover）         |
| type         | `exec` 或 `download`             |
| disabled     | 是否禁用該功能                   |

### Panel 與 Layout 格式（僅 type=panels）
```yaml
layout:
  type: grid
  columns: 2
  gap: 16
panels:
  - title: PDU 即時狀態
    url: http://...
    position:
      row: 1
      col: 1
      colSpan: 2
      rowSpan: 1
```

## 📋 頁面功能說明

### 1. 登入頁面（type: login）
- 使用 `login.templ` 呈現登入畫面，支援 HTMX 表單送出與錯誤回饋。 
- 登入頁面是唯一允許未登入存取的頁面。
- 支援 HTMX 表單提交，避免頁面刷新
- 錯誤訊息即時顯示，提升使用者體驗
- 登入成功後自動重導向至原請求頁面
- 整合 AdminLTE 登入頁面樣式
- 所有請求皆通過 `auth` 中間件驗證
- 登入狀態由 `session` cookie 控制
- 登入失敗時記錄錯誤日誌
- 支援記住登入狀態（由 `COOKIE_MAX_AGE` 控制）
- 登出時清除所有相關 cookie

### 2. PDU 狀態圖頁（type: status）
- 根據 `row`, `col`, `room` 等欄位呈現 PDU 分布圖格狀排列。
- 該頁面支援設備雙邊（L/R）排版與 Grid 資料格組合邏輯。
- 點擊任一 PDU 將以 modal 顯示詳細資訊與對應欄位內容。
- 篩選支援：廠(factory), 區(phase), 資料中心(dc), 房間(room)。
- 使用 `status.csv` 為來源，支援欄位包含 ip,factory,phase,dc,room,row,col,side,is_pdu,name,has_data,protocol_status,status_code,enabled,message
- 將左/右側 PDU（L/R）合併為 `Grid[row][col] = {Left, Right}` 結構
- 支援 `status_code = 0/1/2` （由 env 控制顏色）
- 會計算異常百分比 `abnormalPercent = 紅格 / 總格數` 並顯示於畫面


```bash
   # 標籤顏色配置
   LEVEL_NORMAL_BACKGROUND_COLOR=#28a745       # 正常
   LEVEL_WARN_BACKGROUND_COLOR=#ffc107         # 警告
   LEVEL_ERROR_BACKGROUND_COLOR=#dc3545        # 異常
   LEVEL_DEBUG_BACKGROUND_COLOR=#198754        # 維護
```

- `is_pdu=0` 顯示為空格（非設備格）
```bash
   # PDU平面圖顏色配置
   PDU_EMPTY_BACKGROUND_COLOR=#f8f9fa         # 空櫃底色 (is_pdu=0)
   PDU_EMPTY_TEXT_COLOR=#6c757d               # 空櫃字體色 (is_pdu=0)
   PDU_DEFAULT_BACKGROUND_COLOR=#dee2e6       # 一般底色 (is_pdu=1)
   PDU_DEFAULT_TEXT_COLOR=#343a40             # 一般字體色 (is_pdu=1)
```
- 可讀取 `has_data`, `protocol_status`, `enabled`, `message` 作為詳細訊息（未來可綁定 tooltip 或 modal）
- Grid 排序會將 `AA` 排在最末；col 會轉 int 排序（避免字串排序錯位）



### 3.CSV 類型 CRUD
- 檔案路徑由 `page.Path` 指定。
- 支援 `create`, `edit`, `update`, `delete` 各項操作。
- 每一列轉換為 `Row{ID, Fields}` 並渲染。
- 如有欄位 `interval`, `template_version` 等，也將轉換為下拉、日期等對應欄位型態。
- 可設定 `Upload: true` 允許上傳 CSV 檔案，副檔名需為 `.csv`。
- 每次修改或新增紀錄時皆會自動補上 timestamp 欄位。
- 若有欄位名為 `create_at`, `update_at` 則系統自動填入時間戳。

### 4. 文件設定頁（type: files）
- 可操作 `.env`, `.yaml`, `.txt` 等常見設定檔案。
- 每筆檔案會渲染為一顆按鈕，按下後會顯示全文內容。
- 點選後為全文模式，支援顯示行數與快速搜尋（未來可擴充）。
- `type: view`：唯讀檢視，無法編輯。
- `type: edit`：可進入編輯畫面修改內容。
- 將原始檔案重新命名為 `xxx.20250421-173012.bak` 備份。
- 覆寫該檔案內容為使用者最新儲存內容。
- 目前支援副檔名：`.env`, `.yaml`, `.yml`, `.txt`。

### 5. 工具操作頁（type: exec）
- 每個 Action 對應一段 bash script。
- 每顆按鈕皆對應一支實體 shell script，可透過環境變數控制禁用或切換執行權限。
- 使用者點選按鈕觸發後端執行指令，結果回傳顯示於 modal 或頁面中。
- 結構來源為 `Actions[]`：`label`, `type`, `script_path`。
- `actions[]` 定義執行項目，每項包含 label、path、type(exec/download)、disabled。
- `/exec/run` 實際執行 bash 指令，回傳執行輸出（需允許 server shell）。
- `/exec/download?path=xxx` 限定僅能下載 `data/` 或 `scripts/` 目錄內檔案。

### 6. 系統日誌頁（type: log）
- 讀取 `main.log` 或指定路徑，tail 最末 n 行顯示。
- 每行自動解析 `log level`：INFO / DEBUG / WARN / ERROR。
- 支援多級篩選與雙向滑動，適用於大型部署時快速排查。
- 不同等級標示不同顏色：INFO, DEBUG, WARN, ERROR。

```bash
   # 標籤顏色配置
   LEVEL_NORMAL_BACKGROUND_COLOR=#28a745       # 正常
   LEVEL_WARN_BACKGROUND_COLOR=#ffc107         # 警告
   LEVEL_ERROR_BACKGROUND_COLOR=#dc3545        # 異常
   LEVEL_DEBUG_BACKGROUND_COLOR=#198754        # 維護
```
- 支援 log 訊息多行折疊顯示與語意著色（未來擴充）。
- 支援快速切換排序：`top`（最新在上）/ `tail`（最新在下）。
- 可未來擴充 filter 關鍵字、複製等功能。
- 支援 btn-group 多選切換（INFO / DEBUG / WARN / ERROR）以篩選欲查看的等級。
- 提供懸浮按鈕快速滑動至最上／最下，方便檢視大量日誌。
- 可根據 pages.yaml 中的 `level: DEBUG` 控制最低顯示等級。
- 每行 log 自動解析 `[LEVEL]` 並加上顏色 class（由 template 判斷）。
- 尚不支援頁面搜尋與切換排序，但保留 future 擴充空間。

### 7. Grafana 單一面板（type: dashboard）
- 使用 `page.URL` 作為 iframe src，內嵌完整 Grafana 儀表板頁面
- 通常為 `/d/xxxxx?orgId=1&panelId=1&kiosk` 類連結
- 使用 `iframe.templ` 或簡化版面呈現

### 8. Grafana 多面板（type: panels）
- 使用 `page.Panels[]` 陣列指定多個面板標題與 URL
- 可搭配 `Layout.Columns` 控制面板排列（Grid）
- 透過 `panel_grid.templ` 動態產出 iframe 區塊


---

> 本項目不依賴 AdminLTE JavaScript，純粹使用其 CSS 樣式。所有互動均通過 HTMX 實現。
> 專案已完整下載所有必要的 CSS、字型與 JS 檔案，確保可在離線環境部署。
