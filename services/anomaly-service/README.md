# Python Anomaly Detection API
- 異常檢測 API 服務，提供多種檢測算法。

## 專案結構
```bash
.
├── README.md
├── requirements.txt
├── api.py                      # FastAPI 服務
└── detector/
    ├── __init__.py            # 導出檢測器
    ├── base.py                # 基礎檢測器類
    ├── absolute_threshold.py   # 絕對閾值檢測
    ├── percentage_threshold.py # 百分位閾值檢測
    ├── moving_average.py      # 移動平均檢測
    ├── isolation_forest.py    # 隔離森林檢測
    └── prophet.py             # Prophet 預測
```

## 啟動說明
```bash
# 1. 進入專案目錄
cd pkg/python

# 2. 創建虛擬環境
python3 -m venv venv

# 3. 啟動虛擬環境
source venv/bin/activate  # macOS/Linux
# 或
venv\Scripts\activate     # Windows

# 4. 安裝依賴
pip install -r requirements.txt

# 5. 啟動服務
uvicorn api:app --host 0.0.0.0 --port 8000
```

## 依賴項
```bash
# 基礎依賴
fastapi==0.104.1
uvicorn==0.24.0
pandas==2.1.3
numpy==1.26.2
scikit-learn==1.3.2

# Prophet 依賴
prophet==1.1.5
```


## 數據需求說明
| 檢測器               | 數據需求               |
| -------------------- | ---------------------- |
| absolute_threshold   | 只需要 current         |
| percentage_threshold | 需要 current + history |
| moving_average       | 需要 current + history |
| isolation_forest     | 需要 current + history |
| prophet              | 只需要 history         |

## 健康檢查
```bash
GET /health

回應：
{
  "status": "healthy"
}
```

## API 端點

### 1. 絕對閾值檢測 `/absolute-threshold`
用於檢測數值是否超過固定閾值。

#### 請求參數
```json
{
  "data": {
    "current": [
      {"timestamp": 1738171489, "value": 92.0}
    ]
  },
  "config": {
    "critical_threshold": 90.0,
    "warning_threshold": 80.0,
    "operator": ">="
  }
}
```

#### 回應格式
```json
{
  "data": [
    {
      "timestamp": 1738171489,
      "value": 92.0,
      "severity": "critical",
      "threshold": 90.0
    }
  ]
}
```

#### 閾值邏輯
1. 如果同時設置了 critical 和 warning：
   - 當 `operator(value, critical_threshold)` 為真時：severity = "critical"
   - 當 `operator(value, warning_threshold)` 為真且不是 critical 時：severity = "warning"
   - 其他情況：severity = "normal"

2. 如果只設置了 critical：
   - 當 `operator(value, critical_threshold)` 為真時：severity = "critical"
   - 其他情況：severity = "normal"

#### 運算符說明
- `">"`: 大於
- `">="`: 大於等於
- `"<"`: 小於
- `"<="`: 小於等於

#### 常見用例
1. CPU 使用率過高警告：
```json
{
  "config": {
    "critical_threshold": 90.0,
    "warning_threshold": 80.0,
    "operator": ">="
  }
}
```

2. 記憶體剩餘量過低警告：
```json
{
  "config": {
    "critical_threshold": 10.0,
    "warning_threshold": 20.0,
    "operator": "<="
  }
}
```

### 2. 百分位閾值檢測 `/percentage-threshold`
基於歷史數據的百分位數進行檢測。

#### 請求參數
```json
{
  "data": {
    "current": [
      {"timestamp": 1738171200, "value": 92.5},
      {"timestamp": 1738171260, "value": 85.3},
      {"timestamp": 1738171320, "value": 88.7},
      {"timestamp": 1738171380, "value": 91.2},
      {"timestamp": 1738171440, "value": 86.9}
    ],
    "history": [
      {"timestamp": 1738167600, "value": 75.0},
      {"timestamp": 1738164000, "value": 72.3},
      {"timestamp": 1738160400, "value": 78.9},
      {"timestamp": 1738156800, "value": 71.5},
      {"timestamp": 1738153200, "value": 70.2},
      {"timestamp": 1738149600, "value": 73.8},
      {"timestamp": 1738146000, "value": 76.4},
      {"timestamp": 1738142400, "value": 74.1},
      {"timestamp": 1738138800, "value": 77.5},
      {"timestamp": 1738135200, "value": 75.9},
      {"timestamp": 1738131600, "value": 72.8},
      {"timestamp": 1738128000, "value": 71.3},
      {"timestamp": 1738124400, "value": 73.6},
      {"timestamp": 1738120800, "value": 75.2},
      {"timestamp": 1738117200, "value": 74.7},
      {"timestamp": 1738113600, "value": 76.1},
      {"timestamp": 1738110000, "value": 73.4},
      {"timestamp": 1738106400, "value": 72.9},
      {"timestamp": 1738102800, "value": 75.6},
      {"timestamp": 1738099200, "value": 74.3}
    ]
  },
  "config": {
    "critical_percentile": 95,  // 95th 百分位
    "warning_percentile": 85,   // 85th 百分位
    "operator": ">="           // 大於等於百分位值時觸發警告
  }
}
```

#### 回應格式
```json
{
  "data": [
    {
      "timestamp": 1738171489,
      "value": 92.0,
      "severity": "critical",
      "threshold": 85.0  # 95th 百分位對應的實際值
    }
  ]
}
```

### 3. 移動平均檢測 `/moving-average`
使用移動平均和標準差檢測異常。

#### 請求參數
```json
{
  "data": {
    "current": [
      {"timestamp": 1738171200, "value": 92.5},
      {"timestamp": 1738171260, "value": 85.3},
      {"timestamp": 1738171320, "value": 88.7},
      {"timestamp": 1738171380, "value": 91.2},
      {"timestamp": 1738171440, "value": 86.9}
    ],
    "history": [
      {"timestamp": 1738167600, "value": 75.0},
      {"timestamp": 1738164000, "value": 72.3},
      {"timestamp": 1738160400, "value": 78.9},
      {"timestamp": 1738156800, "value": 71.5},
      {"timestamp": 1738153200, "value": 70.2},
      {"timestamp": 1738149600, "value": 73.8},
      {"timestamp": 1738146000, "value": 76.4},
      {"timestamp": 1738142400, "value": 74.1},
      {"timestamp": 1738138800, "value": 77.5},
      {"timestamp": 1738135200, "value": 75.9},
      {"timestamp": 1738131600, "value": 72.8},
      {"timestamp": 1738128000, "value": 71.3},
      {"timestamp": 1738124400, "value": 73.6},
      {"timestamp": 1738120800, "value": 75.2}
    ]
  },
  "config": {
    "window": 24,           # 移動平均窗口大小
    "std_multiplier": 2.0,  # 標準差倍數
    "min_periods": 12       # 最小數據點數量
  }
}
```

#### 回應格式
```json
{
    "data": [
        {
            "timestamp": 1738171200,
            "value": 92.5,
            "severity": "critical",
            "min": 64.8,
            "max": 86.0
        },
        {
            "timestamp": 1738171260,
            "value": 85.3,
            "severity": "normal",
            "min": 64.64,
            "max": 87.4
        },
        {
            "timestamp": 1738171320,
            "value": 88.7,
            "severity": "normal",
            "min": 64.15,
            "max": 89.38
        },
        {
            "timestamp": 1738171380,
            "value": 91.2,
            "severity": "normal",
            "min": 63.56,
            "max": 91.57
        },
        {
            "timestamp": 1738171440,
            "value": 86.9,
            "severity": "normal",
            "min": 63.79,
            "max": 92.33
        }
    ]
}
```

### 4. 隔離森林檢測 `/isolation-forest`
使用機器學習方法檢測異常點。

#### 請求參數
```json
{
  "data": {
    "current": [
      {"timestamp": 1738171200, "value": 92.5},
      {"timestamp": 1738171260, "value": 85.3},
      {"timestamp": 1738171320, "value": 88.7},
      {"timestamp": 1738171380, "value": 91.2},
      {"timestamp": 1738171440, "value": 86.9}
    ],
    "history": [
      {"timestamp": 1738167600, "value": 75.0},
      {"timestamp": 1738164000, "value": 72.3},
      {"timestamp": 1738160400, "value": 78.9},
      {"timestamp": 1738156800, "value": 71.5},
      {"timestamp": 1738153200, "value": 70.2},
      {"timestamp": 1738149600, "value": 73.8},
      {"timestamp": 1738146000, "value": 76.4},
      {"timestamp": 1738142400, "value": 74.1},
      {"timestamp": 1738138800, "value": 77.5},
      {"timestamp": 1738135200, "value": 75.9},
      {"timestamp": 1738131600, "value": 72.8},
      {"timestamp": 1738128000, "value": 71.3},
      {"timestamp": 1738124400, "value": 73.6},
      {"timestamp": 1738120800, "value": 75.2},
      {"timestamp": 1738117200, "value": 74.7},
      {"timestamp": 1738113600, "value": 76.1},
      {"timestamp": 1738110000, "value": 73.4},
      {"timestamp": 1738106400, "value": 72.9},
      {"timestamp": 1738102800, "value": 75.6},
      {"timestamp": 1738099200, "value": 74.3}
    ]
  },
  "config": {
    "contamination": 0.1,     # 預期異常比例 (10%)
    "n_estimators": 100,      # 樹的數量
    "max_samples": "auto",    # 每棵樹的樣本數
    "max_features": 1.0,      # 特徵使用比例
    "random_state": 42        # 隨機種子
  }
}
```

#### 回應格式
```json
{
  "data": [
    {
      "timestamp": 1738171200,
      "value": 92.5,
      "anomaly": true
    },
    {
      "timestamp": 1738171260,
      "value": 85.3,
      "anomaly": false
    },
    {
      "timestamp": 1738171320,
      "value": 88.7,
      "anomaly": false
    },
    {
      "timestamp": 1738171380,
      "value": 91.2,
      "anomaly": true
    },
    {
      "timestamp": 1738171440,
      "value": 86.9,
      "anomaly": false
    }
  ]
}{
  "data": [
    {
      "timestamp": 1738171200,
      "value": 92.5,
      "anomaly": true
    },
    {
      "timestamp": 1738171260,
      "value": 85.3,
      "anomaly": false
    },
    {
      "timestamp": 1738171320,
      "value": 88.7,
      "anomaly": false
    },
    {
      "timestamp": 1738171380,
      "value": 91.2,
      "anomaly": true
    },
    {
      "timestamp": 1738171440,
      "value": 86.9,
      "anomaly": false
    }
  ]
}
```

### 5. Prophet 檢測 `/prophet`
使用 Facebook Prophet 進行時間序列預測和異常檢測。

#### 請求參數
```json
{
  "data": {
    "history": [
      {"timestamp": 1738167600, "value": 75.0},
      {"timestamp": 1738164000, "value": 72.3},
      {"timestamp": 1738160400, "value": 78.9},
      {"timestamp": 1738156800, "value": 71.5},
      {"timestamp": 1738153200, "value": 70.2},
      {"timestamp": 1738149600, "value": 73.8},
      {"timestamp": 1738146000, "value": 76.4},
      {"timestamp": 1738142400, "value": 74.1},
      {"timestamp": 1738138800, "value": 77.5},
      {"timestamp": 1738135200, "value": 75.9},
      {"timestamp": 1738131600, "value": 72.8},
      {"timestamp": 1738128000, "value": 71.3},
      {"timestamp": 1738124400, "value": 73.6},
      {"timestamp": 1738120800, "value": 75.2},
      {"timestamp": 1738117200, "value": 74.7},
      {"timestamp": 1738113600, "value": 76.1},
      {"timestamp": 1738110000, "value": 73.4},
      {"timestamp": 1738106400, "value": 72.9},
      {"timestamp": 1738102800, "value": 75.6},
      {"timestamp": 1738099200, "value": 74.3},
      {"timestamp": 1738095600, "value": 73.1},
      {"timestamp": 1738092000, "value": 75.8},
      {"timestamp": 1738088400, "value": 74.2},
      {"timestamp": 1738084800, "value": 76.3}
    ]
  },
  "config": {
    "seasonality_mode": "multiplicative",  # 季節性模式
    "changepoint_prior_scale": 0.01,       # 變點先驗尺度
    "interval_width": 0.95,                # 預測區間寬度 (95% 置信區間)
    "uncertainty_samples": 1000,           # 不確定性採樣數
    "horizon": 24,                         # 預測時間長度 (小時)
    "period": "H",                         # 時間週期 (小時)
    "holidays_prior_scale": 10.0,          # 節假日先驗尺度
    "weekly_seasonality": false,           # 週季節性
    "daily_seasonality": true,             # 日季節性
    "yearly_seasonality": false            # 年季節性
  }
}
```

#### 參數說明
- changepoint_prior_scale: 數據平滑時用小值(0.01-0.1)
- weekly_seasonality: 數據少於一週時設為 `false`
- daily_seasonality: 小時級數據設為 `true`
- yearly_seasonality: 數據少於一年時設為 `false`

#### 回應格式
```json
{
    "data": [
        {
            "timestamp": 1738171200,
            "value": 72.25,
            "min": 69.26,
            "max": 75.12
        },
        {
            "timestamp": 1738174800,
            "value": 71.08,
            "min": 68.07,
            "max": 74.22
        },
        {
            "timestamp": 1738178400,
            "value": 71.25,
            "min": 68.19,
            "max": 74.36
        },
        {
            "timestamp": 1738182000,
            "value": 72.13,
            "min": 69.16,
            "max": 75.14
        },
        {
            "timestamp": 1738185600,
            "value": 72.7,
            "min": 69.89,
            "max": 75.72
        },
        {
            "timestamp": 1738189200,
            "value": 72.74,
            "min": 69.71,
            "max": 75.7
        },
        {
            "timestamp": 1738192800,
            "value": 72.93,
            "min": 69.94,
            "max": 75.87
        },
        {
            "timestamp": 1738196400,
            "value": 73.98,
            "min": 71.0,
            "max": 76.82
        },
        {
            "timestamp": 1738200000,
            "value": 75.62,
            "min": 72.63,
            "max": 78.61
        },
        {
            "timestamp": 1738203600,
            "value": 76.78,
            "min": 73.73,
            "max": 79.81
        },
        {
            "timestamp": 1738207200,
            "value": 76.64,
            "min": 73.58,
            "max": 79.5
        },
        {
            "timestamp": 1738210800,
            "value": 75.65,
            "min": 72.8,
            "max": 78.63
        },
        {
            "timestamp": 1738214400,
            "value": 75.3,
            "min": 72.18,
            "max": 78.31
        },
        {
            "timestamp": 1738218000,
            "value": 76.72,
            "min": 73.82,
            "max": 79.56
        },
        {
            "timestamp": 1738221600,
            "value": 79.5,
            "min": 76.76,
            "max": 82.29
        },
        {
            "timestamp": 1738225200,
            "value": 81.91,
            "min": 79.03,
            "max": 85.17
        },
        {
            "timestamp": 1738228800,
            "value": 82.39,
            "min": 79.47,
            "max": 85.29
        },
        {
            "timestamp": 1738232400,
            "value": 81.0,
            "min": 77.96,
            "max": 83.77
        },
        {
            "timestamp": 1738236000,
            "value": 79.36,
            "min": 76.24,
            "max": 82.39
        },
        {
            "timestamp": 1738239600,
            "value": 79.26,
            "min": 76.29,
            "max": 82.3
        },
        {
            "timestamp": 1738243200,
            "value": 81.0,
            "min": 78.14,
            "max": 84.2
        },
        {
            "timestamp": 1738246800,
            "value": 83.24,
            "min": 80.24,
            "max": 86.25
        },
        {
            "timestamp": 1738250400,
            "value": 84.25,
            "min": 81.34,
            "max": 87.02
        },
        {
            "timestamp": 1738254000,
            "value": 83.51,
            "min": 80.63,
            "max": 86.45
        }
    ]
}
```
