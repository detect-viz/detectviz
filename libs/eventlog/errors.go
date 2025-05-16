package eventlog

import "errors"

var (
	// ErrEventNotFound 表示查無指定事件
	ErrEventNotFound = errors.New("event not found")

	// ErrEventSaveFailed 表示事件儲存失敗
	ErrEventSaveFailed = errors.New("failed to save event")

	// ErrEventListFailed 表示事件查詢失敗
	ErrEventListFailed = errors.New("failed to list recent events")
)
