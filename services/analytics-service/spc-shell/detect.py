#!/usr/bin/env python3

import pandas as pd
from dateutil.parser import parse
import os
import sys
import argparse
import json
import yaml

def build_pdu_line(measurement, tags, fields, ts):
    tag_str = ",".join([f"{k}={v}" for k, v in tags.items()])
    field_str = ",".join([f"{k}={v}" for k, v in fields.items()])
    return f"{measurement},{tag_str} {field_str} {ts}"

def process_data(baseline_path, input_path, decimals, metric, fab, measurement, ts,
                usl, lsl, control_chart_scale, zone_a_scale, zone_b_scale,
                pr1_points, pr1_threshold, pr2_points, pr2_threshold,
                pr3_points, pr3_threshold, pr4_points, pr4_threshold,
                pr5_points, pr5_threshold, name, bank, current_date,
                field_definitions):
    """
    處理數據並返回計算結果
    返回格式：
    {
        "stats": {統計數據},
        "pr_results": [PR規則結果],
        "quality_code": 品質代碼
    }
    """
   
    # 在 baseline 檔案不存在或為空時直接返回
    if not os.path.exists(baseline_path) or os.stat(baseline_path).st_size == 0:
        print(f"[WARN] Baseline file not found or empty: {baseline_path}")
        return None

    # 1.先處理長期的資料，計算 mean/stddev
    try:
        baseline_df = pd.read_csv(baseline_path, skiprows=3)
    except Exception as e:
        print(f"[ERROR] Failed to read baseline: {baseline_path}\n{e}")
        return None
    mean = round(baseline_df["_value"].mean(), decimals)
    stddev = round(baseline_df["_value"].std(ddof=0), decimals)
    if pd.isna(mean) or pd.isna(stddev):
        print(f"[WARN] Baseline data is empty or invalid: {baseline_path}")
        return None
    ucl = round(mean + control_chart_scale * stddev, decimals)
    lcl = round(mean - control_chart_scale * stddev, decimals)

    if os.stat(input_path).st_size == 0:
        print(f"[WARN] Empty file skipped: {input_path}")
        return None

    try:
        df = pd.read_csv(input_path, skiprows=3)
    except pd.errors.EmptyDataError:
        print(f"[WARN] No columns to parse in file: {input_path}")
        return None

    def resolve_spec(value, df):
        if isinstance(value, str) and value.upper().startswith("P"):
            try:
                quantile_val = float(value[1:]) / 100
                return round(df["_value"].quantile(quantile_val), decimals)
            except:
                return 0.0
        try:
            return float(value)
        except:
            return 0.0

    spec_usl = resolve_spec(usl, df)
    spec_lsl = resolve_spec(lsl, df)

    # 2.統計 yesterday 資料 (需注入 mean/stddev)
    data_points = len(df)
    daily_mean = round(df["_value"].mean(), decimals)
    daily_stddev = round(df["_value"].std(ddof=0), decimals)
    daily_max = round(df["_value"].max(), decimals)
    daily_min = round(df["_value"].min(), decimals)
    daily_P25 = round(df["_value"].quantile(0.25), decimals)
    daily_P75 = round(df["_value"].quantile(0.75), decimals)
    daily_P02 = round(df["_value"].quantile(0.02), decimals)
    daily_P98 = round(df["_value"].quantile(0.98), decimals)
    daily_lcl = round(mean - control_chart_scale * stddev, decimals)
    daily_ucl = round(mean + control_chart_scale * stddev, decimals)
    daily_cp = round((spec_usl - spec_lsl) / (6 * stddev), decimals) if stddev > 0 else 0
    daily_cpk = round(min((spec_usl - daily_mean), (daily_mean - spec_lsl)) / (3 * stddev), decimals) if stddev > 0 else 0
    zone_a_upper = round(mean + zone_a_scale * stddev, decimals)
    zone_a_lower = round(mean - zone_a_scale * stddev, decimals)
    zone_b_upper = round(mean + zone_b_scale * stddev, decimals)
    zone_b_lower = round(mean - zone_b_scale * stddev, decimals)

    # 3.計算 PR 規則
    df["pr1_flag"] = 0
    df["pr2_flag"] = 0
    df["pr3_flag"] = 0
    df["pr4_flag"] = 0
    df["pr5_flag"] = 0
    # Removed: df["dev"] = abs(df["_value"] - mean)

    # PR1: 有一點超過UCL或LCL
    for i in range(len(df)):
        if df.loc[i, "_value"] > ucl or df.loc[i, "_value"] < lcl:
            df.loc[i, "pr1_flag"] = 1

    # PR2: 連續N點中有M點進入A區
    for i in range(pr2_points-1, len(df)):
        window = df.loc[i-(pr2_points-1):i, "_value"]
        if sum(abs(window - mean) > zone_a_scale*stddev) >= pr2_threshold:
            df.loc[i-(pr2_points-1):i, "pr2_flag"] = 1

    # PR3: 連續N點中有M點進入B區
    for i in range(pr3_points-1, len(df)):
        window = df.loc[i-(pr3_points-1):i, "_value"]
        if sum(abs(window - mean) > zone_b_scale*stddev) >= pr3_threshold:
            df.loc[i-(pr3_points-1):i, "pr3_flag"] = 1

    # PR4: 連續N點以上全部落在CL同一側
    for i in range(pr4_points-1, len(df)):
        window = df.loc[i-(pr4_points-1):i, "_value"]
        if all(window > mean) or all(window < mean):
            df.loc[i-(pr4_points-1):i, "pr4_flag"] = 1

    # PR5: 連續N點持續上升或下降
    for i in range(pr5_points-1, len(df)):
        window = df.loc[i-(pr5_points-1):i, "_value"].tolist()
        if all(window[j] < window[j+1] for j in range(len(window)-1)) or \
           all(window[j] > window[j+1] for j in range(len(window)-1)):
            df.loc[i-(pr5_points-1):i, "pr5_flag"] = 1

    # 4.收集 PR 結果
    pr_results = []
    for _, row in df.iterrows():
        ts_row = int(parse(row["_time"]).timestamp())
        pr_entry = {"timestamp": ts_row}
        for pr in ["pr1_flag", "pr2_flag", "pr3_flag", "pr4_flag", "pr5_flag"]:
            if row[pr] == 1:
                pr_entry[pr] = 1
        if len(pr_entry) > 1:
            pr_results.append(pr_entry)

    # 5.計算品質代碼
    quality_code = 0
    if abs(daily_mean) < 0.001:
        quality_code = 1
    elif daily_stddev < 0.001:
        quality_code = 2
    elif daily_stddev > 10 * abs(daily_mean):
        quality_code = 3
    elif spec_usl == spec_lsl:
        quality_code = 4
    elif (spec_usl - spec_lsl) < 3 * daily_stddev:
        quality_code = 5
    elif (spec_usl - spec_lsl) > 100 * daily_stddev:
        quality_code = 6

    # 6.收集統計數據
    stats = {
        "timestamp": int(ts),
        "base_mean": float(mean),
        "base_std": float(stddev),
        "data_points": int(data_points),
        "mean": float(daily_mean),
        "std": float(daily_stddev),
        "max": float(daily_max),
        "min": float(daily_min),
        "p25": float(daily_P75),
        "p75": float(daily_P75),
        "p02": float(daily_P02),
        "p98": float(daily_P98),
        "lcl": float(daily_lcl),
        "ucl": float(daily_ucl),
        "cp": float(daily_cp),
        "cpk": float(daily_cpk),
        "usl": float(spec_usl),
        "lsl": float(spec_lsl),
        "zone_a_hi": float(zone_a_upper),
        "zone_a_lo": float(zone_a_lower),
        "zone_b_hi": float(zone_b_upper),
        "zone_b_lo": float(zone_b_lower),
        "quality_code": int(quality_code),
        "pr1_counter": int(df["pr1_flag"].sum()),
        "pr1_rate": float(df["pr1_flag"].sum()) / data_points if data_points > 0 else 0,
        "pr2_counter": int(df["pr2_flag"].sum()),
        "pr2_rate": float(df["pr2_flag"].sum()) / data_points if data_points > 0 else 0,
        "pr3_counter": int(df["pr3_flag"].sum()),
        "pr3_rate": float(df["pr3_flag"].sum()) / data_points if data_points > 0 else 0,
        "pr4_counter": int(df["pr4_flag"].sum()),
        "pr4_rate": float(df["pr4_flag"].sum()) / data_points if data_points > 0 else 0,
        "pr5_counter": int(df["pr5_flag"].sum()),
        "pr5_rate": float(df["pr5_flag"].sum()) / data_points if data_points > 0 else 0,
        "pp": float((spec_usl - spec_lsl) / (6 * daily_stddev)) if daily_stddev > 0 else 0,
        "ppk": float(min(spec_usl - daily_mean, daily_mean - spec_lsl) / (3 * daily_stddev)) if daily_stddev > 0 else 0,
        "ooc_counter": int(((df["_value"] > ucl) | (df["_value"] < lcl)).sum()),
        "ooc_rate": float(((df["_value"] > ucl) | (df["_value"] < lcl)).sum()) / data_points if data_points > 0 else 0,
        "oos_counter": int(((df["_value"] > spec_usl) | (df["_value"] < spec_lsl)).sum()),
        "oos_rate": float(((df["_value"] > spec_usl) | (df["_value"] < spec_lsl)).sum()) / data_points if data_points > 0 else 0
    }

    allowed_fields = set(field_definitions.keys())
    # Apply alias if present in field_definitions
    aliased_stats = {}
    for k, v in stats.items():
        if k in field_definitions:
            alias = field_definitions[k].get("alias", k)
            aliased_stats[alias] = v
    stats = aliased_stats

    # 7.組合結果
    result = stats  # 只回傳統計數據

    # 8.組裝 line protocol 格式字串
    lp_lines = []

    tags = {"fab": fab, "name": name, "bank": bank, "metric": metric}
    fields = {k: v for k, v in stats.items() if isinstance(v, (int, float))}
    fields["date_str"] = f'"{current_date}"'
    lp_lines.append(build_pdu_line(measurement, tags, fields, ts))

    # PR規則類
    for entry in pr_results:
        ts_line = entry["timestamp"]
        value_at_ts = df[df["_time"].apply(lambda x: int(parse(x).timestamp())) == ts_line]["_value"]
        if value_at_ts.empty:
            continue
        current_val = round(value_at_ts.values[0], decimals)
        fields = {k: current_val for k in entry if k != "timestamp"}
        line = build_pdu_line(measurement, tags, fields, ts_line)
        lp_lines.append(line)

    return {"stats": result, "lines": lp_lines}

def main():
    parser = argparse.ArgumentParser(description='Process data and generate statistics')
    parser.add_argument('--name', required=True, help='Name')
    parser.add_argument('--bank', required=True, help='Bank')
    parser.add_argument('--field_definitions', required=True, help='Field definitions YAML file path')
    parser.add_argument('--baseline', required=True, help='Baseline CSV file path')
    parser.add_argument('--target', required=True, help='Target CSV file path')
    parser.add_argument('--decimals', required=True, type=int, help='Number of decimal places')
    parser.add_argument('--metric', required=True, help='Metric name')
    parser.add_argument('--fab', required=True, help='Fab name')
    parser.add_argument('--measurement', required=True, help='Measurement name')
    parser.add_argument('--timestamp', required=True, type=int, help='Timestamp')
    parser.add_argument('--usl', required=True, help='Upper specification limit (number or Pxx)')
    parser.add_argument('--lsl', required=True, help='Lower specification limit (number or Pxx)')
    parser.add_argument('--control-chart-scale', required=True, type=float, help='Control chart scale factor')
    parser.add_argument('--zone-a-scale', required=True, type=float, help='Zone A scale factor')
    parser.add_argument('--zone-b-scale', required=True, type=float, help='Zone B scale factor')
    parser.add_argument('--pr1-points', required=True, type=int, help='PR1 points threshold')
    parser.add_argument('--pr1-threshold', required=True, type=int, help='PR1 threshold')
    parser.add_argument('--pr2-points', required=True, type=int, help='PR2 points threshold')
    parser.add_argument('--pr2-threshold', required=True, type=int, help='PR2 threshold')
    parser.add_argument('--pr3-points', required=True, type=int, help='PR3 points threshold')
    parser.add_argument('--pr3-threshold', required=True, type=int, help='PR3 threshold')
    parser.add_argument('--pr4-points', required=True, type=int, help='PR4 points threshold')
    parser.add_argument('--pr4-threshold', required=True, type=int, help='PR4 threshold')
    parser.add_argument('--pr5-points', required=True, type=int, help='PR5 points threshold')
    parser.add_argument('--pr5-threshold', required=True, type=int, help='PR5 threshold')
    parser.add_argument('--current-date', required=True, help='Current date (e.g., 2025-04-20)')
    # removed --output-lp and --output-json arguments
    
    args = parser.parse_args()

    with open(args.field_definitions, "r") as f:
        field_definitions = yaml.safe_load(f)
    
    result = process_data(
        args.baseline,
        args.target,
        args.decimals,
        args.metric,
        args.fab,
        args.measurement,
        args.timestamp,
        args.usl,
        args.lsl,
        args.control_chart_scale,
        args.zone_a_scale,
        args.zone_b_scale,
        args.pr1_points,
        args.pr1_threshold,
        args.pr2_points,
        args.pr2_threshold,
        args.pr3_points,
        args.pr3_threshold,
        args.pr4_points,
        args.pr4_threshold,
        args.pr5_points,
        args.pr5_threshold,
        args.name,
        args.bank,
        args.current_date,
        field_definitions
    )

    if result:
        print(json.dumps(result))

if __name__ == "__main__":
    main()