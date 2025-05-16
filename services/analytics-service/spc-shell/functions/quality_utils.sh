#!/bin/bash

# 檢查質量代碼
quality_code() {
  local mean="$1"
  local stddev="$2"
  local usl="$3"
  local lsl="$4"
  local name="$5"
  local desc=""
  local code=0

  if (( $(echo "(${mean} < 0.001 && ${mean} > -0.001)" | bc -l) )); then
    desc="mean too small"; code=1
  elif (( $(echo "${stddev} < 0.001" | bc -l) )); then
    desc="stddev too small"; code=2
  elif (( $(echo "${stddev} > 10 * (${mean} < 0 ? -${mean} : ${mean})" | bc -l) )); then
    desc="stddev much larger than mean"; code=3
  elif [[ "$usl" == "$lsl" ]]; then
    desc="usl == lsl (no tolerance)"; code=4
  elif (( $(echo "(${usl} - ${lsl}) < 3 * ${stddev}" | bc -l) )); then
    desc="tolerance too small for SPC"; code=5
  elif (( $(echo "(${usl} - ${lsl}) > 100 * ${stddev}" | bc -l) )); then
    desc="tolerance too large, may be incorrect"; code=6
  fi

  if [[ "$code" -ne 0 ]]; then
    echo "[WARN][$name] quality_code=$code, reason: $desc" >&2
  fi

  echo "$code"
}

# 計算並寫入 lp 檔案 
calc_value_to_lp() {
  local metric="$1"
  local mean="$2"
  local stddev="$3"
  local usl="$4"
  local lsl="$5"
  local bank_name="$6"
  local target_name="$7"
  local fab="$8"
  local control_chart_scale="$9"
  local zone_a_scale="${10}"
  local zone_b_scale="${11}"
  local qcode="${12}"
  local ucl lcl cp cpk_raw_u cpk_raw_l cpk_raw cpk
  local zone_a_upper zone_a_lower zone_b_upper zone_b_lower
  local LINES=()

  echo "[DEBUG][$bank_name][$target_name] mean: $mean, stddev: $stddev, usl: $usl, lsl: $lsl" >&2

  # 檢查必要的值是否存在
  if [[ -z "$mean" || -z "$stddev" || -z "$usl" || -z "$lsl" || -z "$bank_name" || -z "$target_name" ]]; then
    echo "Error: Missing required values for calculation" >&2
    return 1
  fi

  # 防呆處理：若 stddev 太小或為 0，設定為 MIN_STDDEV（避免除以 0）
  if (( $(echo "$stddev <= 0" | bc -l) )) || (( $(echo "$stddev < $MIN_STDDEV" | bc -l) )); then
    echo "[calc_value_to_lp][$bank_name][$target_name] stddev=$stddev too small or zero, force to MIN_STDDEV=$MIN_STDDEV" >&2
    stddev="$MIN_STDDEV"
  fi

  # 防呆處理：若 usl 與 lsl 相等，視為無效範圍
  if [[ "$usl" == "$lsl" ]]; then
    echo "[calc_value_to_lp][$bank_name][$target_name] usl == lsl ($usl), invalid range" >&2
    cp=0
    cpk=0
    ucl="$mean"
    lcl="$mean"
    zone_a_upper="$mean"
    zone_a_lower="$mean"
    zone_b_upper="$mean"
    zone_b_lower="$mean"
  else
    # 正常情況：執行計算
    ucl=$(printf "%.2f" "$(echo "scale=4; $mean + $control_chart_scale * $stddev" | bc -l)")
    lcl=$(printf "%.2f" "$(echo "scale=4; $mean - $control_chart_scale * $stddev" | bc -l)")
    zone_a_upper=$(printf "%.2f" "$(echo "scale=4; $mean + $zone_a_scale * $stddev" | bc -l)")
    zone_a_lower=$(printf "%.2f" "$(echo "scale=4; $mean - $zone_a_scale * $stddev" | bc -l)")
    zone_b_upper=$(printf "%.2f" "$(echo "scale=4; $mean + $zone_b_scale * $stddev" | bc -l)")
    zone_b_lower=$(printf "%.2f" "$(echo "scale=4; $mean - $zone_b_scale * $stddev" | bc -l)")
    cp=$(printf "%.2f" "$(echo "scale=4; ($usl - $lsl) / (6 * $stddev)" | bc -l)")
    cpk_raw_u=$(echo "scale=4; $usl - $mean" | bc -l)
    cpk_raw_l=$(echo "scale=4; $mean - $lsl" | bc -l)
    cpk_raw=$(echo -e "$cpk_raw_u\n$cpk_raw_l" | sort -n | head -n1)
    cpk=$(printf "%.2f" "$(echo "scale=4; $cpk_raw / (3 * $stddev)" | bc -l)")
  fi

  echo "[DEBUG][$bank_name][$target_name] mean: $mean, stddev: $stddev, usl: $usl, lsl: $lsl, ucl: $ucl, lcl: $lcl, cp: $cp, cpk: $cpk, zone_a_upper: $zone_a_upper, zone_a_lower: $zone_a_lower, zone_b_upper: $zone_b_upper, zone_b_lower: $zone_b_lower" >&2

  # 寫入 InfluxDB
  local influx_cmd="influx write --bucket $WRITE_BUCKET --precision s --host $WRITE_INFLUX_URL --token $WRITE_INFLUX_TOKEN --org $WRITE_ORG \"$QUERY_MEASUREMENT,metric=$metric,fab=$fab,bank=$bank_name,name=$target_name UCL=$ucl,LCL=$lcl,CP=$cp,CPK=$cpk,zone_a_upper=$zone_a_upper,zone_a_lower=$zone_a_lower,zone_b_upper=$zone_b_upper,zone_b_lower=$zone_b_lower,quality_code=$qcode $WRITE_TIMESTAMP\""
  echo "[DEBUG] Executing command: $influx_cmd" >&2
  eval "$influx_cmd" 2>&1

  # 返回 ucl 和 lcl 值
  echo "$ucl $lcl"
} 