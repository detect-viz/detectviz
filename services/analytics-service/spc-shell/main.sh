#!/bin/bash

# 載入配置和工具函數
source config.sh
source functions/flux_utils.sh
source functions/quality_utils.sh
source functions/target_info.sh

# 處理命令行參數
if [[ $# -lt 2 ]]; then
    echo "Usage: $0 <start_date> <end_date> [query_field] [--dry-run]"
    echo "Example: $0 2025-04-08 2025-04-08 current"
    echo "Example: $0 2025-04-08 2025-04-09 current"
    exit 1
fi

# 設定日期範圍
START_DATE="$1"
END_DATE="$2"
QUERY_FIELD="${3:-current}" # 如果沒有提供查詢欄位，使用 current

export OUTPUT_LINE="==============================================================================="

# 處理 dry-run 參數
DRY_RUN=0
if [[ "$4" == "--dry-run" ]]; then
    DRY_RUN=1
fi

# 檢查日期格式
if ! perl -MTime::Piece -e 'Time::Piece->strptime($ARGV[0], "%Y-%m-%d")' "$START_DATE" 2>/dev/null; then
    echo "Error: Invalid start date format. Please use YYYY-MM-DD"
    exit 1
fi
if ! perl -MTime::Piece -e 'Time::Piece->strptime($ARGV[0], "%Y-%m-%d")' "$END_DATE" 2>/dev/null; then
    echo "Error: Invalid end date format. Please use YYYY-MM-DD"
    exit 1
fi

# 檢查結束日期不能早於開始日期
START_EPOCH=$(perl -MTime::Piece -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->epoch' "$START_DATE")
END_EPOCH=$(perl -MTime::Piece -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->epoch' "$END_DATE")
if [[ $END_EPOCH -lt $START_EPOCH ]]; then
    echo "Error: End date cannot be earlier than start date"
    exit 1
fi

# 產生查詢用 flux 腳本
generate_query_flux() {
    local name="$1"
    local bank="$2"
    local start="$3"
    local stop="$4"

    # 直接返回 flux 查詢字符串
    cat <<EOF
from(bucket: "${QUERY_BUCKET}")
  |> range(start: ${start}, stop: ${stop})
  |> filter(fn: (r) => r["_measurement"] == "${QUERY_MEASUREMENT}")
  |> filter(fn: (r) => r["_field"] == "${QUERY_FIELD}")
  |> filter(fn: (r) => r["bank"] == "${bank}")
  |> filter(fn: (r) => r["name"] == "${name}")
  |> keep(columns: ["_time", "bank", "_value", "name"])
EOF
}

# 設定日期相關變數 DAILY_START_TIME/DAILY_STOP_TIME/BASELINE_START_TIME/BASELINE_STOP_TIME/WRITE_TIME/WRITE_TIMESTAMP
setup_date_variables() {
    local base_date="$1"

    # 當日開始
    export DAILY_START_TIME=$(perl -MTime::Piece -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->strftime("%Y-%m-%dT00:00:00Z")' "$base_date")
    # 當日結束
    export DAILY_STOP_TIME=$(perl -MTime::Piece -MTime::Seconds -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->add(ONE_DAY)->strftime("%Y-%m-%dT00:00:00Z")' "$base_date")
    # 基線開始
    export BASELINE_START_TIME=$(perl -MTime::Piece -MTime::Seconds -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->add(-$ENV{BASELINE_LOOKBACK_DAYS} * ONE_DAY)->strftime("%Y-%m-%dT00:00:00Z")' "$base_date")
    # 基線結束
    export BASELINE_STOP_TIME="$DAILY_START_TIME"
    # 寫入時間
    export WRITE_TIME="$DAILY_START_TIME"
    # 寫入時間戳記
    export WRITE_TIMESTAMP=$(perl -MTime::Piece -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%dT%H:%M:%SZ")->epoch' "$WRITE_TIME")
}
init_daily_points() {
    local bank="$1"
    local name="$2"
    local metric="$3"
    local date="$4"   # 2025-04-08
    local points="$5" # 144

    local line_tmpl="pdu_line,bank=$bank,name=$name,metric=$metric {replace_content} {timestamp}"
    local content=""

    # 轉換為 base timestamp（秒）
    local base_ts=$(perl -MTime::Piece -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->epoch' "$date")
    local interval=$((86400 / points)) # 1天 = 86400 秒

    for ((i = 0; i < points; i++)); do
        local ts=$((base_ts + i * interval))
        content+=$(echo "$line_tmpl" | sed "s/{timestamp}/$ts/")$'\n'
    done

    echo "$content"
}

process_target() {
    local group="$1"
    local fab="$2"
    local name="$3"
    local current_date="$4"

    echo "$OUTPUT_LINE"
    echo "Processing target: $name (group: $group, fab: $fab) for date: $current_date" >&2
    echo "$OUTPUT_LINE"

    # 設定日期相關變數 DAILY_START_TIME/DAILY_STOP_TIME/BASELINE_START_TIME/BASELINE_STOP_TIME/WRITE_TIME/WRITE_TIMESTAMP
    setup_date_variables "$current_date"
    echo "DAILY_RANGE: $DAILY_START_TIME ~ $DAILY_STOP_TIME"
    echo "BASELINE_RANGE: $BASELINE_START_TIME ~ $BASELINE_STOP_TIME"
    echo "WRITE_TIME: $WRITE_TIME [$WRITE_TIMESTAMP]"

    # 獲取目標群組資訊 target_info.sh
    read -r usl lsl control_chart_scale zone_a_scale zone_b_scale \
        pr1_points pr1_threshold pr2_points pr2_threshold \
        pr3_points pr3_threshold pr4_points pr4_threshold \
        pr5_points pr5_threshold banks <<<"$(get_target_group_info "$group" "$name")"

    # 從 banks 字符串中讀取 bank 列表
    IFS=':' read -ra bank_array <<<"$banks"

    # 建立必要的目錄
    mkdir -p "${BUILD_DIR}"
    mkdir -p "${RESULT_DIR}"

    # 遍歷 bank 列表
    for bank in "${bank_array[@]}"; do
        process_code="${current_date}_${name}_${bank}_${QUERY_FIELD}"

        # 預先產生 144 筆資料點模板
        DAILY_CONTENT=$(init_daily_points "$bank" "$name" "$QUERY_FIELD" "$current_date" "$DAILY_POINTS")

        # 產生 baseline csv
        BASELINE_CSV="$BUILD_DIR/${process_code}_baseline.csv"
        flux_query=$(generate_query_flux "$name" "$bank" "$BASELINE_START_TIME" "$BASELINE_STOP_TIME")
        influx_cmd="influx query '$flux_query' --raw > $BASELINE_CSV"

        if [[ $DRY_RUN -eq 0 ]]; then
            eval "$influx_cmd" 2>&1
        fi
        #echo "CMD: $influx_cmd"

        # 檢查文件是否存在且包含數據
        if [[ ! -f "$BASELINE_CSV" ]]; then
            echo "[WARN] Processed $process_code → $BASELINE_CSV not found" >&2
            continue
        fi
        if [[ ! -s "$BASELINE_CSV" || $(wc -l <"$BASELINE_CSV") -le 3 ]]; then
            echo "[WARN] Processed $process_code → $BASELINE_CSV is empty or contains no data" >&2
            rm -f "$BASELINE_CSV"
            continue
        fi

        # 產生 daily csv
        DAILY_CSV="${BUILD_DIR}/${process_code}_daily.csv"
        flux_query=$(generate_query_flux "$name" "$bank" "$DAILY_START_TIME" "$DAILY_STOP_TIME")
        influx_cmd="influx query '$flux_query' --raw > $DAILY_CSV"

        if [[ $DRY_RUN -eq 0 ]]; then
            eval "$influx_cmd" 2>&1
        fi
        #echo "CMD: $influx_cmd"

        # 檢查文件是否存在且包含數據
        if [[ ! -f "$DAILY_CSV" ]]; then
            echo "[WARN] Processed $process_code → $DAILY_CSV not found" >&2
            continue
        fi
        if [[ ! -s "$DAILY_CSV" || $(wc -l <"$DAILY_CSV") -le 3 ]]; then
            echo "[WARN] Processed $process_code → $DAILY_CSV is empty or contains no data" >&2
            rm -f "$DAILY_CSV"
            continue
        fi

        # 設定 detect.py 所需的參數
        METRIC=${METRIC:-current}
        MEASUREMENT=${MEASUREMENT:-pdu}
        TIMESTAMP=${WRITE_TIMESTAMP}
        RESULT_JSON="${RESULT_DIR}/${process_code}_result.json"

        # 組裝 detect.py 指令
        detect_cmd=(python3 detect.py)

        detect_cmd+=(
            --decimals "$DECIMALS"
            --name "$name"
            --bank "$bank"
            --field_definitions "$FIELD_YAML"
            --metric "$METRIC"
            --fab "$fab"
            --measurement "$MEASUREMENT"
            --timestamp "$TIMESTAMP"
            --baseline "$BASELINE_CSV"
            --target "$DAILY_CSV"
            --usl "$USL_MODE"
            --lsl "$LSL_MODE"
            --control-chart-scale "$control_chart_scale"
            --zone-a-scale "$zone_a_scale"
            --zone-b-scale "$zone_b_scale"
            --pr1-points "$pr1_points"
            --pr1-threshold "$pr1_threshold"
            --pr2-points "$pr2_points"
            --pr2-threshold "$pr2_threshold"
            --pr3-points "$pr3_points"
            --pr3-threshold "$pr3_threshold"
            --pr4-points "$pr4_points"
            --pr4-threshold "$pr4_threshold"
            --pr5-points "$pr5_points"
            --pr5-threshold "$pr5_threshold"
            --current-date "$current_date"
        )

        # 取代統一欄位內容（動態取得欄位）
        # 定義靜態欄位鍵名清單
        STATIC_KEYS=("base_mean" "mean" "lcl" "ucl" "zone_a_hi" "zone_a_lo" "zone_b_hi" "zone_b_lo")

        # ➤ 根據 OUTPUT_LP 控制 detect.py 輸出與寫入方式
        if [[ "$OUTPUT_LP" == "true" ]]; then
            # ➤ 執行 detect.py 並將其標準輸出導入 influx write 寫入資料庫（先捕獲內容再寫入）
            CONTENT=$("${detect_cmd[@]}")

            # 從 .stats 取得欄位鍵值對（統計值）
            STATIC_FIELDS=$(echo "$CONTENT" | jq -r --argjson keys "$(printf '%s\n' "${STATIC_KEYS[@]}" | jq -R . | jq -s .)" '
            .stats as $stats |
            $keys
            | map("\(.)=" + ($stats[.]|tostring))
            | join(",")')

            REPLACED_CONTENT=$(echo "$DAILY_CONTENT" | sed "s/{replace_content}/$STATIC_FIELDS/g")

            # 正確展開 JSON 陣列每一行為 line protocol 格式（避免包含中括號[]）
            LP_STATIC="$(echo "$CONTENT" | jq -r '.lines[]')"
            LP_CONTENT="$LP_STATIC"$'\n'"$REPLACED_CONTENT"

            echo "$LP_CONTENT" | influx write \
                --bucket "$WRITE_BUCKET" \
                --precision s \
                --host "$WRITE_INFLUX_URL" \
                --token "$WRITE_INFLUX_TOKEN" \
                --org "$WRITE_ORG"

            echo "[INFO] Processed $process_code → OK"
        else
            # ➤ 僅顯示標準輸出，適用於除錯或 dry-run 模式
            "${detect_cmd[@]}"
        fi

        rm -f "$DAILY_CSV" "$BASELINE_CSV"

    done
    return 0
}

# 主處理循環
current_date="$START_DATE"
while [[ $(perl -MTime::Piece -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->epoch' "$current_date") -le $END_EPOCH ]]; do
    echo "[$(perl -MTime::Piece -e 'print localtime->strftime("%Y-%m-%d %H:%M:%S")' 2>/dev/null)] Processing $current_date"

    # 處理所有目標
    while IFS=, read -r group fab name; do
        # 跳過標題行
        [[ "$group" == "group" ]] && continue
        echo "$(date +%s) [INFO] Processed Start $name"
        echo "$(date +%s) [DEBUG] read: group=$group fab=$fab name=$name"
        process_target "$group" "$fab" "$name" "$current_date" || true &
        if [[ $(jobs -r | wc -l) -ge $WORKERS ]]; then
            wait -n
        fi
        echo "$(date +%s) [INFO] Processed Done $name"
    done <"$TARGET_CSV"
    wait

    # 更新日期到下一天
    current_date=$(perl -MTime::Piece -MTime::Seconds -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->add(ONE_DAY)->strftime("%Y-%m-%d")' "$current_date")
done
