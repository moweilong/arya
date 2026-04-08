package message

type Message struct {
	ReceiveIdType string `json:"receive_id_type"` // 接收人类型，可选值：user_id、open_id、chat_id、email
	ReceiveId     string `json:"receive_id"`      // 消息接收者的 ID，ID 类型与查询参数 receive_id_type 的取值一致。
	MsgType       string `json:"msg_type"`        // 消息类型。interactive：卡片
	Content       string `json:"content"`         // 卡片内容
}
