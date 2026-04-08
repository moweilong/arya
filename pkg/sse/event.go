package sse

import (
	"strconv"
	"sync"
	"time"
)

// 对象池
var eventPool = sync.Pool{
	New: func() any {
		return &Event{}
	},
}

// 事件
type Event struct {
	ID    string        // 事件 ID
	Type  string        // 事件类型
	Data  []byte        // 事件数据
	Retry time.Duration // 重连间隔

	bitset uint8 // 内部标记位，使用位操作标记字段是否被设置，避免写入空字段
}

const (
	isSetID    uint8 = 1 << iota // 0001
	isSetType                    // 0010
	isSetData                    // 0100
	isSetRetry                   // 1000
)

// 从对象池获取 Event
func NewEvent() *Event {
	event := eventPool.Get().(*Event)
	event.reset()
	return event
}

func (e *Event) reset() {
	e.ID = ""
	e.Type = ""
	e.Data = nil
	e.Retry = 0
	e.bitset = 0
}

// 归还对象池
func (e *Event) Release() {
	e.reset()
	eventPool.Put(e)
}

// 深拷贝 Event
func (e *Event) Clone() *Event {
	newEvent := NewEvent()
	newEvent.ID = e.ID
	newEvent.Type = e.Type
	newEvent.Data = make([]byte, len(e.Data))
	copy(newEvent.Data, e.Data)
	newEvent.Retry = e.Retry
	newEvent.bitset = e.bitset
	return newEvent
}

// 设置字段（同时标记 bitset）
func (e *Event) SetID(id string) {
	e.ID = id
	e.bitset |= isSetID
}

func (e *Event) SetEvent(eventType string) {
	e.Type = eventType
	e.bitset |= isSetType
}

func (e *Event) SetData(data []byte) {
	e.Data = data
	e.bitset |= isSetData
}

func (e *Event) SetDataString(data string) {
	e.Data = []byte(data)
	e.bitset |= isSetData
}

// 追加数据
func (e *Event) AppendData(data []byte) {
	e.Data = append(e.Data, data...)
	e.bitset |= isSetData
}

func (e *Event) AppendDataString(data string) {
	e.Data = append(e.Data, []byte(data)...)
	e.bitset |= isSetData
}

func (e *Event) SetRetry(retry time.Duration) {
	e.Retry = retry
	e.bitset |= isSetRetry
}

// 检查字段是否已设置
func (e *Event) IsSetID() bool {
	return e.bitset&isSetID != 0
}

func (e *Event) IsSetType() bool {
	return e.bitset&isSetType != 0
}

func (e *Event) IsSetData() bool {
	return e.bitset&isSetData != 0
}

func (e *Event) IsSetRetry() bool {
	return e.bitset&isSetRetry != 0
}

// 序列化为 SSE 格式
func (e *Event) String() string {
	var result []byte

	if e.IsSetID() {
		result = append(result, "id: "...)
		result = append(result, e.ID...)
		result = append(result, '\n')
	}

	if e.IsSetType() {
		result = append(result, "event: "...)
		result = append(result, e.Type...)
		result = append(result, '\n')
	}

	if e.IsSetData() { // SSE 协议规定，如果数据里有换行，必须每一行都单独加 data:
		lines := splitLines(e.Data)
		for _, line := range lines {
			result = append(result, "data: "...)
			result = append(result, line...)
			result = append(result, '\n')
		}
	}

	if e.IsSetRetry() {
		result = append(result, "retry: "...)
		result = append(result, strconv.FormatInt(e.Retry.Milliseconds(), 10)...)
		result = append(result, '\n')
	}

	result = append(result, '\n')
	return string(result)
}

// a\r\nb\n
// ["a", "b", ""]
func splitLines(data []byte) [][]byte {
	if len(data) == 0 {
		return [][]byte{nil}
	}

	var lines [][]byte
	start := 0

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '\n':
			lines = append(lines, data[start:i])
			start = i + 1
		case '\r':
			lines = append(lines, data[start:i])
			if i+1 < len(data) && data[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}

	if start < len(data) {
		// 剩下的内容不是空 → 加入最后一行
		lines = append(lines, data[start:])
	} else if start == len(data) && len(lines) > 0 {
		// 结尾是换行符 → 补充一个空行
		lines = append(lines, nil)
	}

	return lines
}
