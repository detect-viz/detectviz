

# eventlog

此模組用於集中處理場域端傳來的操作與事件紀錄，支援中控 `viot-dc` 使用者查詢與分析異常行為。

## 功能包含：

- 定義 `EventLog` 結構
- 提供 `EventStore` interface
- 基礎實作：記憶體儲存器（可日後換成 MySQL, SQLite）

## 使用方式

```go
store := eventlog.NewMemoryStore()
store.SaveEvent(&eventlog.EventLog{
    FabCode: "F12P7DC1R2",
    Type: "deploy",
    Severity: "info",
    Message: "部署完成",
    CreatedAt: time.Now(),
})
```