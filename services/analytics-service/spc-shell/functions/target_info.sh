#!/bin/bash

# 獲取目標群組資訊
get_target_group_info() {
  local group="$1"
  local name="$2"
  local group_csv="$GROUP_CSV"
  local usl=""
  local lsl=""
  local control_chart_scale=""
  local zone_a_scale=""
  local zone_b_scale=""
  local pr1_points=""
  local pr1_threshold=""
  local pr2_points=""
  local pr2_threshold=""
  local pr3_points=""
  local pr3_threshold=""
  local pr4_points=""
  local pr4_threshold=""
  local pr5_points=""
  local pr5_threshold=""
  local banks=""

  while IFS=, read -r g u l c z1 z2 p1 t1 p2 t2 p3 t3 p4 t4 p5 t5 b; do
    if [[ "$g" == "$group" ]]; then
      usl="$u"
      lsl="$l"
      control_chart_scale="$c"
      zone_a_scale="$z1"
      zone_b_scale="$z2"
      pr1_points="$p1"
      pr1_threshold="$t1"
      pr2_points="$p2"
      pr2_threshold="$t2"
      pr3_points="$p3"
      pr3_threshold="$t3"
      pr4_points="$p4"
      pr4_threshold="$t4"
      pr5_points="$p5"
      pr5_threshold="$t5"
      banks="$b"
      break
    fi
  done < <(tail -n +2 "$group_csv")

  # 輸出參數供後續使用
  echo "$usl $lsl $control_chart_scale $zone_a_scale $zone_b_scale $pr1_points $pr1_threshold $pr2_points $pr2_threshold $pr3_points $pr3_threshold $pr4_points $pr4_threshold $pr5_points $pr5_threshold $banks"
}
