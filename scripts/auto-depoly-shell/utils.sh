#!/bin/bash
SHELL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SHELL_DIR/log.sh"

# =================== IP 轉換函數 ===================
# Function to convert IP to integer
ip_to_int() {
    func_name="ip_to_int"
    local a b c d
    IFS=. read -r a b c d <<< "$1"
    local result=$((a * 256 ** 3 + b * 256 ** 2 + c * 256 + d))
    echo "$result"
    log_debug "$func_name" "將 IP $1 轉換為整數 $result"
}


# Function to convert integer to IP
int_to_ip() {
    func_name="int_to_ip"
    local num=$1
    local ip="$((num >> 24 & 255)).$((num >> 16 & 255)).$((num >> 8 & 255)).$((num & 255))"
    echo "$ip"
    log_debug "$func_name" "將整數 $num 轉換為 IP $ip"
}

# 生成 IP 範圍
generate_ip_range() {
    func_name="generate_ip_range"
    local start_ip="$1"
    local end_ip="$2"
    
    log_info "$func_name" "生成 IP 範圍從 $start_ip 到 $end_ip"
    
    local start_num
    local end_num
    start_num=$(ip_to_int "$start_ip" | head -n1)
    end_num=$(ip_to_int "$end_ip" | head -n1)
    
    for ((i=start_num; i<=end_num; i++)); do
        int_to_ip "$i"
    done
}
