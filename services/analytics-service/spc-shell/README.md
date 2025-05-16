經完成了代碼重構，主要做了以下改動：

1. 將 `main.sh` 的日期範圍處理功能整合到 `main.sh` 中：
   - 添加了日期範圍參數處理
   - 添加了 GNU date 檢測
   - 添加了主處理循環

2. 改進了參數處理：
   - 現在支持可選的結束日期和查詢欄位
   - 提供了更清晰的使用說明和示例
   - 保持了原有的 dry-run 功能

3. 修改了 `process_target` 函數：
   - 添加了 `current_date` 參數
   - 將日期相關變數的設置移到函數內部
   - 更新了文件路徑生成邏輯

4. 刪除了 `main.sh`，因為其功能已經整合到 `main.sh` 中

使用方式現在更加靈活：

# 處理單一日期
./main.sh 2025-04-08 2025-04-08 current

# 處理日期範圍
./main.sh 2025-04-08 2025-04-09 current

4. 使用 dry-run 模式：
```bash
./main.sh 2025-04-09 2025-05-09 current --dry-run
```


## 統計類欄位

| 欄位名         | 說明             | 樣本區間 | 公式/邏輯 |
|---------------|------------------|----------|-----------|
| base_mean     | 前7日基線平均     | -7d      | 前7日 value 平均值 |
| base_std      | 前7日基線標準差   | -7d      | 前7日 value 標準差 |
| mean          | 當日平均         | -1d      | 當日 value 平均值 |
| std           | 當日標準差       | -1d      | 當日 value 標準差 |
| max           | 當日最大值       | -1d      | 當日 value 最大值 |
| min           | 當日最小值       | -1d      | 當日 value 最小值 |
| p25           | 第25百分位數     | -1d      | quantile(value, 0.25) |
| p75           | 第75百分位數     | -1d      | quantile(value, 0.75) |
| ucl           | 管制上限         | -1d      | mean + 3 × std |
| lcl           | 管制下限         | -1d      | mean - 3 × std |
| cp            | 製程能力 Cp      | -1d      | (usl - lsl) / (6 × std) |
| cpk           | 製程能力 Cpk     | -1d      | min(usl - mean, mean - lsl) / (3 × std) |
| data_points   | 當日筆數         | -1d      | 當日資料筆數 |
| usl           | 規格上限         | -1d      | 客戶提供或自動推算 |
| lsl           | 規格下限         | -1d      | 客戶提供或以lcl為值 |
| zone_a_hi     | A區上限          | -1d      | mean + 2 × std |
| zone_a_lo     | A區下限          | -1d      | mean - 2 × std |
| zone_b_hi     | B區上限          | -1d      | mean + 1 × std |
| zone_b_lo     | B區下限          | -1d      | mean - 1 × std |
| date_str      | 日期文字         | -1d      | 當日日期字串 |
| pr1_trigger   | PR1觸發值        | -1d      | 超過UCL或LCL時標記 |
| pr2_trigger   | PR2觸發值        | -1d      | 連續3點中有2點進入A區 |
| pr3_trigger   | PR3觸發值        | -1d      | 連續5點中有4點進入B區 |
| pr4_trigger   | PR4觸發值        | -1d      | 連續8點以上全部落在CL同一側 |
| pr5_trigger   | PR5觸發值        | -1d      | 連續7點持續上升或下降 |
| pr1_counter   | 當日PR1觸發次數      | -1d      | count(pr1_trigger) |
| pr2_counter   | 當日PR2觸發次數      | -1d      | count(pr2_trigger) |
| pr3_counter   | 當日PR3觸發次數      | -1d      | count(pr3_trigger) |
| pr4_counter   | 當日PR4觸發次數      | -1d      | count(pr4_trigger) |
| pr5_counter   | 當日PR5觸發次數      | -1d      | count(pr5_trigger) |
| pr1_rate      | 當日PR1觸發率        | -1d      | pr1_counter / data_points |
| pr2_rate      | 當日PR2觸發率        | -1d      | pr2_counter / data_points |
| pr3_rate      | 當日PR3觸發率        | -1d      | pr3_counter / data_points |
| pr4_rate      | 當日PR4觸發率        | -1d      | pr4_counter / data_points |
| pr5_rate      | 當日PR5觸發率        | -1d      | pr5_counter / data_points |

## 統計資料可信度

| quality_code | 品質代碼說明 |
|--------------|-------------|
| 0 | 正常（無特殊異常） |
| 1 | 異常：平均值接近0（abs(mean) < 0.001） |
| 2 | 異常：標準差過小（stddev < 0.001） |
| 3 | 異常：標準差過大（stddev > 10 × abs(mean)） |
| 4 | 異常：規格上下限相等（usl == lsl） |
| 5 | 異常：規格區間過小（usl - lsl < 3 × stddev） |
| 6 | 異常：規格區間過大（usl - lsl > 100 × stddev） |

## 需求
負載 （PHASE/BANK）
三相不平衡10%(異常點數/PR1~5，OOB/OOC)-KPI(R/S/T)
左右差10%(OOB/OOC)-KPI(R/S/T)
總表（異常點數/PR1~5，OOB/OOC）-->基礎趨勢或散點圖-->分析圖(P圖/C圖/u圖/NP圖)/SPC指標（CP/CPK）/4週比較圖（P圖帶盒鬚圖

*   `mean = average(value)`
    
*   `stddev = 標準差(value)`
    
*   `ucl = mean + 3 × stddev`
    
*   `lcl = mean - 3 × stddev`

*   `zoneA`
    start: mean + 2 * stddev
		stop:  mean - 2 * stddev

*   `zoneB`
    start: mean + 1 * stddev
		stop:  mean - 1 * stddev




## KPI
控制圖 + 平均 + UCL/LCL 線 + 異常點
做出 SPC 控制圖
算出 CPK / OOB / 三相不平衡

| 指標 | 說明 |
| --- | --- |
| `mean`, `stddev` | 平均與標準差 |
| `ucl`, `lcl` | 控制上下限 |
| `cp`, `cpk` | 程序能力 |
| `oob_p98/p2` | OOB 範圍 |
| `k_shift_flag` | 連續偏移偵測 |
| `box_min/p25/p75/max/mean` | 4週比較與穩定性 |

📦 所有公式說明（彙整）
-------------

### SPC 控制線：

*   `mean = average(value)`
    
*   `stddev = 標準差(value)`
    
*   `ucl = mean + N × stddev`
    
*   `lcl = mean - N × stddev`
    

### CP / CPK：

*   `cp = (usl - lsl) / (6 * σ)`  
*   `cpk = min((usl - mean), (mean - lsl)) / (3 * σ)`



    

### OOB 判斷：

*   `P98 = quantile(value, 0.98)`
    
*   `P2 = quantile(value, 0.02)`
    
*   `oob_flag = value > P98 or value < P2`
    

### 三相不平衡：

*   `imbalance_ratio = (max(L1,L2,L3) - min(L1,L2,L3)) / avg(L1,L2,L3)`
    

### 左右差：

*   `bank_diff_ratio = abs(BANK_L - BANK_R) / ((BANK_L + BANK_R)/2)`
   

### Flux：
mean_val = data |> mean()
std_val = data |> stddev()
ucl = mean_val + std_val * 3.0
lcl = mean_val - std_val * 3.0
max/min/p25/p75

### 現有電流資料結構：

```bash
_measurement: pdu
_field: current
_value: 2.3
_time: 2025-04-22T10:00:00Z
pdu_name: F12P8DC1R3Y67PL
bank: L1
```

#### tag1 schema 
```bash
key: pdu_name
value: F12P8DC1R3Y67PL、F12P8DC1R3Y67PR ...
```

#### tag2 schema
```bash
key: phase
value: L1, L2, L3
```

#### tag3 schema
```bash
key: bank
value: L1-1, L1-2, L2-1, L2-2, L3-1, L3-2
```
### USL
客戶提供或自己推算（例：最大15A×0.5×0.8）
### LSL
取最近7天的 P5分位數（quantile(0.05)）

### Chart
```csv
系統名稱,計算項目,系統內公式,Flux對應語法,需自定閾值,對應圖表
SPC Rule - UCL/LCL,"UCL, LCL, Target, σ","UCL = mean + kσ, LCL = mean - kσ","mean(), stddev(), map()","是（k 倍數, 通常為 3）",控制圖（Time Series）
OOB Rule,"Percentile (P2, P98)、OOB 異常點",值超過 percentile(P98) 或低於 P2 為異常,"quantile(q: 0.98), quantile(q: 0.02)",是（百分位數門檻）,Time Series + 點異常標註
K-Shift,連續 K 點上升/下降,K 點皆大於/小於前一點,reduce() / state tracking 自寫,是（K 值，如 5 點）,異常標記/計數報表
Median Shift,連續點偏離中位數,連續 N 點在中位數同側,"median(), reduce() 自寫",是（N 值、偏差方向）,異常標記或累計
Box Chart,"P01-P99, Min, Max, Mean, P25, P75",統計五數概括 + 均值,"quantile(), mean(), min(), max()",否（圖表顯示用）,Box Plot（Grafana 插件或外部彙總）
Trend Chart,任意時段均值/波動趨勢,時間序列展示,"aggregateWindow(), filter()",否,時序圖（Grafana Time Series）

```

## install



### Centos

yum install python3
yum install vim
yum install tree
yum install zip
yum install unzip
yum install perl
yum install jq -y
sudo yum install perl-Time-Piece

sudo yum install -y epel-release
sudo yum install -y python3-pip
pip3 install pandas
pip3 install pandas python-dateutil
pip3 install --user

pip3 install pyyaml


mkdir -p /tmp/rpm-list
cd /tmp/rpm-list/
rpm -qa --qf "%{NAME}\n" | sort > /tmp/rpm-list/installed-packages.txt
rpm -qa | sort > /tmp/rpm-list/installed-packages-full2.txt
systemctl list-unit-files --type=service --state=enabled > /tmp/rpm-list/enabled-services3.txt

firewall-cmd --list-services
firewall-cmd --reload
systemctl start firewalld
firewall-cmd --zone=public --add-port=3000/tcp --permanent
firewall-cmd --zone=public --add-port=8086/tcp --permanent
firewall-cmd --reload


### grafana
```ini
[auth.anonymous]
# enable anonymous access
enabled = true

# specify role for unauthenticated users
org_role = Viewer
;http_port = 8080
```

### influxdb
influx config set --active \
  -n admin \
  --host-url http://localhost:8086

sudo apt update && sudo apt upgrade -y
pip install pandas python-dateutil pyyaml
apt install jq -y

./main.sh 2025-04-09 2025-05-09 current

