from typing import List, Dict
import pandas as pd
import numpy as np
from prophet import Prophet
from datetime import datetime
from .base import BaseDetector

class ProphetDetector(BaseDetector):
    def detect(self, current: List[Dict], history: List[Dict]) -> List[Dict]:
        """
        使用 Prophet 進行時間序列預測和異常檢測
        """
        try:
            # 檢查輸入數據
            if not history:
                return []
            
            # 將 Pydantic 模型轉換為字典
            history_data = [
                {
                    'timestamp': point.timestamp,
                    'value': point.value
                }
                for point in history
            ]
            
            # 創建 DataFrame 並轉換時間戳
            df = pd.DataFrame(history_data)
            df['ds'] = pd.to_datetime(df['timestamp'], unit='s')
            df['y'] = df['value']
            
            # 配置 Prophet 模型
            try:
                model = Prophet(
                    seasonality_mode=self.config['seasonality_mode'],
                    changepoint_prior_scale=float(self.config['changepoint_prior_scale']),
                    interval_width=float(self.config['interval_width']),
                    yearly_seasonality=bool(self.config['yearly_seasonality']),
                    weekly_seasonality=bool(self.config['weekly_seasonality']),
                    daily_seasonality=bool(self.config['daily_seasonality']),
                    growth='flat'  # 使用平坦增長模式
                )
                
                # 添加小時級別的季節性
                model.add_seasonality(
                    name='hourly',
                    period=24,
                    fourier_order=5
                )
            except Exception as e:
                raise Exception(f"Prophet 模型配置失敗: {str(e)}")
            
            # 訓練模型
            try:
                model.fit(df[['ds', 'y']])
            except Exception as e:
                raise Exception(f"Prophet 模型訓練失敗: {str(e)}")
            
            # 生成預測時間點
            last_timestamp = df['ds'].max()
            future = model.make_future_dataframe(
                periods=self.config['horizon'],
                freq=self.config['period'],
                include_history=False  # 只生成未來時間點
            )
            
            # 進行預測
            forecast = model.predict(future)
            
            # 格式化輸出
            results = []
            for _, row in forecast.iterrows():
                if row['ds'] > last_timestamp:  # 只返回預測部分
                    results.append({
                        'timestamp': int(row['ds'].timestamp()),
                        'value': round(float(row['yhat']), 2),
                        'min': round(float(row['yhat_lower']), 2),
                        'max': round(float(row['yhat_upper']), 2)
                    })
            
            return results
            
        except Exception as e:
            raise Exception(f"Prophet 預測失敗: {str(e)}") 