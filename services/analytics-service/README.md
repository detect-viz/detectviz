# analytics

一個基於 AI 的系統監控報告生成工具。

## 功能特點

- 支持多種異常檢測模型
  - 絕對閾值檢測
  - 百分比閾值檢測
  - 移動平均檢測
  - Isolation Forest 檢測
  - Prophet 時序預測
- 靈活的配置系統
  - 支持全局默認配置
  - 支持規則級別配置覆蓋
  - 忽略空值配置
- 多種 LLM 支持
  - OpenAI API
  - Azure OpenAI
  - Anthropic
  - 本地 LLM 部署

## 配置說明

### 1. 基礎配置 (settings.yaml)

```yaml
server:
  port: ":8080"

python_api:
  host: "http://localhost:5001"
  timeout: 30
  models:
    absolute_threshold:
      window_size: 60
    percentage_threshold:
      thresholds:
        critical: 90.0
        warning: 80.0
    moving_average:
      window_size: 60
      std_dev: 2.0
    isolation_forest:
      contamination: 0.1
      n_estimators: 100
    prophet:
      horizon: "86400s"
      period: "60s"

llm_api:
  mode: "api"  # api 或 local
  api:
    provider: "openai"
    url: "https://api.openai.com/v1"
    model: "gpt-3.5-turbo"
  local:
    host: "http://localhost:8001"
    model_path: "./models/llama2"
```

### 2. 規則配置 (rules.yaml)

```yaml
rules:
  cpu_usage:
    metric_name: "cpu_usage"
    model: "percentage_threshold"
    config:
      thresholds:
        critical: 95  # 覆蓋默認配置
        warning: 85   # 覆蓋默認配置
```

### 3. 面板配置 (panels.yaml)

```yaml
panels:
  PANEL-001:
    title: "系統資源監控"
    rules:
      - cpu_usage
      - memory_usage

prediction_panels:
  PREDICTION-001:
    title: "資源使用預測"
    metrics:
      - cpu_usage
    prophet_settings:
      horizon: "24h"  # 可選，覆蓋默認配置
```

## 配置優先級

1. 規則配置 (`rules.yaml`) 優先於全局配置
2. 空值或未設定的配置項會使用全局默認值
3. 每個規則可以選擇性地覆蓋部分配置

## 使用說明

1. 修改配置文件
2. 啟動服務：`go run main.go`
3. 調用 API 生成報告

## API 文檔

- GET `/health`: 健康檢查
- GET `/api/v1/metrics/{profile_id}`: 獲取指定 profile 需要的 metrics 列表
- POST `/api/v1/detect`: 執行異常檢測並生成報告

詳細的 API 文檔請參考 [API.md](./API.md)

## 專案結構
```bash
.
├── README.md
├── config/
│   ├── rules.yaml     # 定義檢測規則和模型
│   ├── panels.yaml    # 定義面板和指標分組
│   └── settings.yaml  # 系統配置設定
├── services/
│   └── analyzer/      # 異常分析服務
│       ├── analyzer.go      # 主要分析邏輯
│       ├── python_client.go # Python 服務調用
│       ├── panels.go       # 面板處理邏輯
│       ├── config.go       # 配置處理邏輯
│       └── errors.go       # 錯誤處理
└── pkg/
    └── python/        # Python 異常檢測服務
        ├── api.py           # FastAPI 服務
        ├── cli.py           # 命令行介面
        └── detector/        # 檢測器實現
            ├── __init__.py
            ├── base.py           # 基礎檢測器
            ├── absolute_threshold.py
            ├── percentage_threshold.py
            ├── moving_average.py
            ├── isolation_forest.py
            └── prophet.py
```

## 面板設計
### 一般面板
用於監控當前指標的異常情況，支持多種檢測模型：
```yaml
panels:
  PANEL-001:
    title: "系統資源監控"
    description: "顯示 CPU、記憶體、負載等關鍵指標"
    prompt_template: "系統關鍵指標出現異常，請根據以下異常數據分析可能原因，並提供最佳優化方案。"
    rules:
      - cpu_usage
      - memory_usage
      - load_average
```

### 預測面板
使用 Prophet 模型進行預測分析：
```yaml
prediction_panels:
  PREDICT-001:
    title: "資源使用趨勢預測"
    description: "預測系統資源使用趨勢"
    metrics:
      - cpu_usage
      - memory_usage
    prophet_settings:
      horizon: 24
      period: "H"
```

## 異常檢測模型

### 數據需求
| 檢測器               | 數據需求               | 輸出                    |
| -------------------- | ---------------------- | ----------------------- |
| absolute_threshold   | 只需要 current         | critical/warning/normal |
| percentage_threshold | 需要 current + history | critical/warning/normal |
| moving_average       | 需要 current + history | critical/warning/normal |
| isolation_forest     | 需要 current + history | true/false              |
| prophet              | 只需要 history         | 預測 + 異常檢測         |

### 檢測器說明
1. **absolute_threshold**: 明確固定閾值檢測
2. **percentage_threshold**: 基於歷史數據百分位的動態閾值
3. **moving_average**: 移動平均值加減標準差範圍
4. **isolation_forest**: 基於數據分佈的自動異常檢測
5. **prophet**: 預測未來趨勢並進行異常檢測

## API 使用說明

### 健康檢查
```bash
curl -X GET "http://localhost:8080/health"
```

#### 回應格式
```json
{
  "status": "ok"
}
```

### 獲取 Metrics 列表
```bash
curl -X GET "http://localhost:8080/api/v1/metrics/SYSTEM-001"
```

#### 回應格式
```json
{
  "metrics": [
      {"metric_name": "cpu_usage","unit": "%"},
      {"metric_name": "memory_usage","unit": "MB"},
      {"metric_name": "disk_usage","unit": "GB"}
    ],
    "current": true,
    "history": true
}
```

#### 說明
- `metrics`: 需要提供的指標列表，包含名稱和單位
- `current`: 是否需要提供當前數據
- `history`: 是否需要提供歷史數據

### 異常檢測
```bash
curl -X POST "http://localhost:8080/api/v1/detect" \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "profile_id": "PANEL-001",
      "current": [
        {
          "timestamp": 1738171489,
          "metric": "cpu_usage",
          "value": 95.1
        }
      ],
      "history": [
        {
          "timestamp": 1738167889,
          "metric": "cpu_usage",
          "value": 60.0
        }
      ]
    }
  }'
```

## 如何擴展
1. 新增監控指標：在 `rules.yaml` 添加新的 metrics
2. 修改 LLM 提示詞：調整 panels 中的 prompt_template
3. 新增面板：在 `panels.yaml` 中添加新的面板配置
4. 新增檢測器：在 detector/ 目錄實現新的檢測器

## 開發說明
1. Python API 支持 FastAPI 和 CLI 兩種調用方式
2. 使用環境變量 USE_CLI=true 切換到 CLI 模式
3. 檢測器實現必須繼承 BaseDetector
4. 所有配置支持全局默認值和規則級別的覆蓋

## 使用場景

### 1. 系統健康報告
監控和分析系統關鍵指標：
- CPU、記憶體、磁碟使用率
- 網絡流量和延遲
- 系統負載和進程數

### 2. 應用性能報告
監控應用服務的性能指標：
- 響應時間
- 請求成功率
- 錯誤率和類型
- 並發連接數

### 3. 資源預測報告
預測系統資源使用趨勢：
- 容量規劃
- 資源擴展建議
- 成本優化方案

## 開發指南

### 1. 添加新的檢測器
1. 在 `detector/` 目錄創建新的檢測器文件
2. 實現 `BaseDetector` 介面
3. 在 `__init__.py` 中註冊新檢測器
4. 更新 `rules.yaml` 添加新的模型配置

### 2. 自定義 LLM 提示詞
提示詞支持的變量：
- `.Metrics`: 異常指標列表
- `.Timestamp`: 當前時間戳
- `.PanelTitle`: 面板標題
- `.Severity`: 異常程度

### 3. 錯誤處理
主要錯誤類型：
- `DetectorError`: 檢測器相關錯誤
- `ConfigError`: 配置相關錯誤
- `APIError`: API 調用錯誤

## 貢獻指南
1. Fork 專案
2. 創建功能分支
3. 提交更改
4. 發起 Pull Request

## 授權
MIT License

