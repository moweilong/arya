package sse

import (
	"net/http"
)

const (
	// SSE 协议标识
	MIMEType = "text/event-stream"
)

// 设置 SSE 响应头
func SetupSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", MIMEType)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// 判断请求是否接受 SSE
func IsSSEAcceptable(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return accept == MIMEType || accept == "*/*" || accept == ""
}
