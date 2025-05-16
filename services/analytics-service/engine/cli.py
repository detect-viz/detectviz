#!/usr/bin/env python3
import sys
import json
from detector import (
    ProphetDetector, 
    MovingAverageDetector, 
    IsolationForestDetector,
    AbsoluteThresholdDetector,
    PercentageThresholdDetector
)

def main():
    # 從標準輸入讀取 JSON
    data = sys.stdin.read()
    try:
        input_data = json.loads(data)
        
        detector_type = input_data.get('type', 'prophet')
        config = input_data.get('config', {})
        
        # 檢查數據需求
        data_requirements = {
            'absolute_threshold': {'current': True, 'history': False},
            'percentage_threshold': {'current': True, 'history': True},
            'moving_average': {'current': True, 'history': True},
            'isolation_forest': {'current': True, 'history': True},
            'prophet': {'current': False, 'history': True}
        }
        
        # 檢查必要的數據是否存在
        requirements = data_requirements.get(detector_type)
        if requirements:
            data = input_data.get('data', {})
            if requirements['current'] and 'current' not in data:
                raise ValueError(f"{detector_type} 需要 current 數據")
            if requirements['history'] and 'history' not in data:
                raise ValueError(f"{detector_type} 需要 history 數據")
        
        # 檢測器映射
        detectors = {
            'prophet': ProphetDetector,
            'moving_average': MovingAverageDetector,
            'isolation_forest': IsolationForestDetector,
            'absolute_threshold': AbsoluteThresholdDetector,
            'percentage_threshold': PercentageThresholdDetector
        }
        
        # 獲取檢測器類
        detector_class = detectors.get(detector_type)
        if not detector_class:
            raise ValueError(f"不支持的檢測器類型: {detector_type}")
            
        # 創建檢測器實例
        detector = detector_class(config)
        
        # 執行檢測
        results = detector.detect(
            current=input_data.get('data', {}).get('current', []),
            history=input_data.get('data', {}).get('history', [])
        )
        
        # 輸出結果
        print(json.dumps({"data": results}))
        
    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main() 