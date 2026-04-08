package sse

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/schema"
)

type Writer struct {
	id      string                          // 当前会话 ID
	w       http.ResponseWriter             // HTTP 响应写入器
	flusher http.Flusher                    // 支持流式刷新的接口
	closed  bool                            // 是否已关闭
	Done    func(http.ResponseWriter) error // 完成时回调
}

func NewWriter(id string, w http.ResponseWriter) *Writer {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil
	}

	SetupSSEHeaders(w)

	return &Writer{
		id:      id,
		w:       w,
		flusher: flusher,
		closed:  false,
	}
}

func (w *Writer) SetDone(f func(http.ResponseWriter) error) {
	w.Done = f
}

func (w *Writer) Close() {
	w.closed = true
}

func (w *Writer) IsClosed() bool {
	return w.closed
}

// 写入单个 Event
func (w *Writer) WriteEvent(event *Event) error {
	if event == nil || w.closed {
		return nil
	}

	_, err := w.w.Write([]byte(event.String()))
	if err != nil {
		w.closed = true
		return err
	}

	w.flusher.Flush()
	return nil
}

func (w *Writer) WriteEventSimple(id, eventType string, data []byte) error {
	event := NewEvent()
	defer event.Release()

	if id != "" {
		event.SetID(id)
	}
	if eventType != "" {
		event.SetEvent(eventType)
	}
	if data != nil {
		event.SetData(data)
	}

	return w.WriteEvent(event)
}

// 字符串数据版本
func (w *Writer) WriteEventString(id, eventType, data string) error {
	return w.WriteEventSimple(id, eventType, []byte(data))
}

// JSON 数据版本
func (w *Writer) WriteEventJSON(id, eventType string, data any) error {
	jsonData, err := sonic.Marshal(data)
	if err != nil {
		return err
	}

	return w.WriteEventSimple(id, eventType, jsonData)
}

// 快速写入数据（使用默认 ID）
func (w *Writer) WriteJSONData(data any) error {
	jsonData, err := sonic.Marshal(data)
	if err != nil {
		return err
	}

	return w.WriteData(jsonData)
}

func (w *Writer) WriteData(data []byte) error {
	event := NewEvent()
	defer event.Release()
	event.SetID(w.id)
	event.SetData(data)
	return w.WriteEvent(event)
}

func (w *Writer) WriteDataString(data string) error {
	return w.WriteData([]byte(data))
}

// 写入注释（用于心跳/保活）
func (w *Writer) WriteComment(comment string) error {
	_, err := fmt.Fprintf(w.w, ": %s\n\n", comment)
	if err != nil {
		return err
	}

	w.flusher.Flush()
	return nil
}

// 发送心跳保活
func (w *Writer) WriteKeepAlive() error {
	return w.WriteComment("keep-alive")
}

// 发送结束信号 `[DONE]`
func (w *Writer) WriteDone() error {
	if w.Done != nil {
		err := w.Done(w.w)
		if err != nil {
			return err
		}
		w.flusher.Flush()
		return nil
	}

	_, err := fmt.Fprintf(w.w, "data: [DONE]\n\n")
	if err != nil {
		return err
	}

	w.flusher.Flush()
	return nil
}

// 上下文，支持取消/超时
// Eino 流式消息读取器
// 消息转换函数，可自定义输出格式
func (w *Writer) Stream(ctx context.Context, stream *schema.StreamReader[*schema.Message], fn func(output *schema.Message, index int) any) error {
	if fn == nil {
		fn = func(output *schema.Message, index int) any {
			return output
		}
	}

	index := 0
	for {
		// 检查上下文是否被取消或 Writer 是否已关闭
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if w.IsClosed() {
				return fmt.Errorf("writer is closed")
			}
		}

		// 接收数据
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return w.WriteDone() // 正常结束
			}
			return err
		}

		// 如果 chunk 为空为空，则结束流
		if chunk == nil || (chunk.Content == "" && chunk.ReasoningContent == "") {
			continue
		}

		// 转换消息格式
		newChunk := fn(chunk, index)
		if newChunk == nil {
			continue
		}

		index++
		b, err := sonic.Marshal(newChunk)
		if err != nil {
			return err
		}

		// 写入 SSE 流
		if err = w.WriteData(b); err != nil {
			return err
		}
	}
}
