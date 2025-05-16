package eventlog

import "log"

// MemoryStore 是一個記憶體中的事件儲存器，實作 EventStore 介面。
// 僅適用於本地測試與模擬用途，不建議用於正式部署。
type MemoryStore struct {
	logs []EventLog
}

// NewMemoryStore 建立一個新的記憶體事件儲存器。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{logs: []EventLog{}}
}

// SaveEvent 將事件儲存至記憶體中。
func (s *MemoryStore) SaveEvent(e *EventLog) error {
	s.logs = append(s.logs, *e)
	log.Printf("[eventlog] saved: %+v\n", *e)
	return nil
}

// ListRecent 回傳最近 N 筆事件。
func (s *MemoryStore) ListRecent(limit int) ([]EventLog, error) {
	if limit > len(s.logs) {
		limit = len(s.logs)
	}
	return s.logs[len(s.logs)-limit:], nil
}
