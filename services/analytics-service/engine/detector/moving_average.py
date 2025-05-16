from typing import List, Dict
import pandas as pd
import numpy as np
from .base import BaseDetector

class MovingAverageDetector(BaseDetector):
    def detect(self, current: List[Dict], history: List[Dict]) -> List[Dict]:
        """
        使用移動平均和標準差檢測異常
        """
        try:
            # 檢查輸入數據
            if not current or not history:
                return []
            
            # 將 Pydantic 模型轉換為字典
            current_data = [
                {
                    'timestamp': point.timestamp,
                    'value': point.value
                }
                for point in current
            ]
            
            history_data = [
                {
                    'timestamp': point.timestamp,
                    'value': point.value
                }
                for point in history
            ]
            
            # 創建 DataFrame 並設置時間戳為索引
            current_df = pd.DataFrame(current_data).set_index('timestamp')
            history_df = pd.DataFrame(history_data).set_index('timestamp')
            
            # 確保索引排序
            current_df = current_df.sort_index()
            history_df = history_df.sort_index()
            
            # 合併數據計算移動平均
            all_data = pd.concat([history_df, current_df])
            all_data = all_data[~all_data.index.duplicated(keep='last')]  # 處理重複索引
            
            # 使用配置的窗口大小
            window = self.config.get('window', 24)
            std_multiplier = self.config.get('std_multiplier', 2.0)
            min_periods = self.config.get('min_periods', 12)
            
            # 計算移動平均和標準差
            rolling = all_data['value'].rolling(window=window, min_periods=min_periods)
            moving_avg = rolling.mean()
            moving_std = rolling.std()
            
            # 計算上下界
            upper_bound = moving_avg + (moving_std * std_multiplier)
            lower_bound = moving_avg - (moving_std * std_multiplier)
            
            # 只保留當前時間段的數據
            results = current_df.copy()
            results['min'] = round(lower_bound[current_df.index], 2)
            results['max'] = round(upper_bound[current_df.index], 2)
            
            # 標記嚴重程度
            results['severity'] = 'normal'
            results.loc[results['value'] > upper_bound[current_df.index], 'severity'] = 'critical'
            results.loc[results['value'] < lower_bound[current_df.index], 'severity'] = 'critical'
            
            # 格式化輸出
            return [
                {
                    'timestamp': int(idx),
                    'value': round(float(row['value']), 2),
                    'severity': row['severity'],
                    'min': round(float(row['min']), 2),
                    'max': round(float(row['max']), 2)
                }
                for idx, row in results.iterrows()
            ]
            
        except Exception as e:
            raise Exception(f"移動平均檢測失敗: {str(e)}") 