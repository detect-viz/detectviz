


source "$BASE_DIR/lib/env.sh"
load_env_constants



source "$BASE_DIR/usr/lib/viot/scripts/util/log.sh"
log_init

log_info "main" "==================== START VIOT DEPLOY ===================="
log_info "main" "Mode: $mode | Time: $(date '+%Y-%m-%d %H:%M:%S')"
log_info "main" "Working directory: $BASE_DIR"

---

## 📘 log.sh 日誌模組使用說明

此專案內建 `log.sh` 為通用日誌輸出模組，支援 console level 與 log 檔輸出控制，可作為 shell 專案的輕量 scaffold 使用。

### 匯入方式

```bash
source \"\$BASE_DIR/lib/log.sh\"
log_init
```

### 參數控制

```bash
export LOG_LEVEL=DEBUG           # 控制畫面輸出層級（預設 INFO）
export LOG_FILE_LEVEL=INFO       # 控制寫入檔案層級（預設 INFO）
export LOG_FILE_PATH=/var/log/viot/main.log  # 寫入檔案路徑（若無則不輸出）
export LOG_ROTATE_BY_DAY=true     # 每日自動產生 log.2024-04-14.log 格式
```

### 使用方式

```bash
log_info \"module\" \"任務開始\"
log_warn \"module\" \"警告資訊\"
log_error \"module\" \"錯誤後中止\"
```

### 執行起始範例

```bash
log_info \"main\" \"==================== START VIOT DEPLOY ====================\"
log_info \"main\" \"Mode: \$mode | Time: \$(date '+%Y-%m-%d %H:%M:%S')\"
log_info \"main\" \"Working directory: \$BASE_DIR\"
```

### Log 分日（Rotate by Day）

此模組支援每日自動分檔功能，預設為開啟。

```bash
export LOG_ROTATE_BY_DAY=true  # 每日自動產生 log.2024-04-14.log 格式
```

如有設定 `LOG_FILE_PATH=/var/log/viot/main.log`，實際寫入會是：

```
/var/log/viot/main.2024-04-14.log
```

若要關閉此功能，請設定：

```bash
export LOG_ROTATE_BY_DAY=false
```