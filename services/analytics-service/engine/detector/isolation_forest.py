from typing import List, Dict
import pandas as pd
import numpy as np
from sklearn.ensemble import IsolationForest
from .base import BaseDetector

class IsolationForestDetector(BaseDetector):
    def detect(self, current: List[Dict], history: List[Dict]) -> List[Dict]:
        """
        使用隔離森林檢測異常
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
            
            # 創建 DataFrame
            current_df = pd.DataFrame(current_data)
            history_df = pd.DataFrame(history_data)
            
            # 合併數據進行訓練
            all_data = pd.concat([history_df, current_df])
            
            # 配置隔離森林
            clf = IsolationForest(
                contamination=float(self.config['contamination']),
                n_estimators=int(self.config['n_estimators']),
                max_samples=self.config['max_samples'],
                max_features=float(self.config['max_features']),
                random_state=int(self.config['random_state'])
            )
            
            # 訓練並預測
            clf.fit(all_data[['value']])
            predictions = clf.predict(current_df[['value']])
            
            # 格式化輸出
            return [
                {
                    'timestamp': int(row['timestamp']),
                    'value': round(float(row['value']), 2),
                    'anomaly': bool(pred == -1)  # 轉換為 Python 原生布爾值
                }
                for (_, row), pred in zip(current_df.iterrows(), predictions)
            ]
            
        except Exception as e:
            raise Exception(f"隔離森林檢測失敗: {str(e)}") 