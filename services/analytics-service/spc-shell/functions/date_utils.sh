#!/bin/bash

# 設定日期相關變數
setup_date_variables() {
  local base_date="$1"
  
  # 使用 perl 處理日期轉換
  export DAILY_START_TIME=$(perl -MTime::Piece -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->strftime("%Y-%m-%dT00:00:00Z")' "$base_date")
  export DAILY_STOP_TIME=$(perl -MTime::Piece -MTime::Seconds -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->add(ONE_DAY)->strftime("%Y-%m-%dT00:00:00Z")' "$base_date")
  export BASELINE_START_TIME=$(perl -MTime::Piece -MTime::Seconds -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%d")->add(-$ENV{BASELINE_LOOKBACK_DAYS} * ONE_DAY)->strftime("%Y-%m-%dT00:00:00Z")' "$base_date")
  export BASELINE_STOP_TIME="$DAILY_START_TIME"
  export WRITE_TIME="$DAILY_START_TIME"
  export WRITE_TIMESTAMP=$(perl -MTime::Piece -e 'print Time::Piece->strptime($ARGV[0], "%Y-%m-%dT%H:%M:%SZ")->epoch' "$WRITE_TIME")
} 