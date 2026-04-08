package message

import (
	"context"
	"fmt"

	"github.com/gookit/slog"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// ReceiveIdType 接收人类型
type ReceiveIdType string

const (
	ReceiveIdTypeUserID ReceiveIdType = "user_id" // 用户ID
	ReceiveIdTypeOpenID ReceiveIdType = "open_id" // Open ID
	ReceiveIdTypeChatID ReceiveIdType = "chat_id" // 群聊ID
	ReceiveIdTypeEmail  ReceiveIdType = "email"   // 邮箱
)

// MsgType 消息类型
type MsgType string

const (
	MsgTypeText       MsgType = "text"       // 文本消息
	MsgTypePost       MsgType = "post"       // 富文本消息
	MsgTypeInteractive MsgType = "interactive" // 卡片消息
	MsgTypeImage      MsgType = "image"      // 图片消息
	MsgTypeFile       MsgType = "file"       // 文件消息
	MsgTypeAudio      MsgType = "audio"      // 语音消息
	MsgTypeMedia      MsgType = "media"      // 媒体消息
	MsgTypeSticker    MsgType = "sticker"    // 表情包消息
)

// Service 消息服务
type Service struct {
	larkClient *lark.Client
}

// MessageService 消息服务别名
type MessageService = Service

// NewMessageService 创建消息服务
func NewMessageService(larkClient *lark.Client) *MessageService {
	return &MessageService{
		larkClient: larkClient,
	}
}

// Send 发送消息
// https://open.feishu.cn/document/server-docs/im-v1/message/create
func (s *MessageService) Send(ctx context.Context, msg *Message) (string, error) {
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(msg.ReceiveIdType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(msg.ReceiveId).
			MsgType(msg.MsgType).
			Content(msg.Content).
			Build()).
		Build()

	resp, err := s.larkClient.Im.V1.Message.Create(ctx, req)
	if err != nil {
		slog.Errorf("发送飞书消息失败: %v", err)
		return "", err
	}

	if !resp.Success() {
		errMsg := fmt.Sprintf("logId: %s, error response: %s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		slog.Error(errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	if resp.Data == nil || resp.Data.MessageId == nil {
		return "", fmt.Errorf("响应数据为空")
	}

	return *resp.Data.MessageId, nil
}

// SendCard 发送卡片消息
// receiveIdType: 接收人类型 (user_id, open_id, chat_id, email)
// receiveId: 接收人ID
// cardContent: 卡片 JSON 内容
// 返回 messageID 可用于后续更新卡片
func (s *MessageService) SendCard(ctx context.Context, receiveIdType ReceiveIdType, receiveId, cardContent string) (string, error) {
	msg := &Message{
		ReceiveIdType: string(receiveIdType),
		ReceiveId:     receiveId,
		MsgType:       string(MsgTypeInteractive),
		Content:       cardContent,
	}
	return s.Send(ctx, msg)
}

// SendCardToUser 发送卡片消息给指定用户
func (s *MessageService) SendCardToUser(ctx context.Context, userID, cardContent string) (string, error) {
	return s.SendCard(ctx, ReceiveIdTypeUserID, userID, cardContent)
}

// SendCardToOpenID 发送卡片消息给指定 OpenID
func (s *MessageService) SendCardToOpenID(ctx context.Context, openID, cardContent string) (string, error) {
	return s.SendCard(ctx, ReceiveIdTypeOpenID, openID, cardContent)
}

// SendCardToChat 发送卡片消息到群聊
func (s *MessageService) SendCardToChat(ctx context.Context, chatID, cardContent string) (string, error) {
	return s.SendCard(ctx, ReceiveIdTypeChatID, chatID, cardContent)
}

// SendCardToEmail 发送卡片消息给指定邮箱
func (s *MessageService) SendCardToEmail(ctx context.Context, email, cardContent string) (string, error) {
	return s.SendCard(ctx, ReceiveIdTypeEmail, email, cardContent)
}

// SendText 发送文本消息
func (s *MessageService) SendText(ctx context.Context, receiveIdType ReceiveIdType, receiveId, text string) (string, error) {
	content := fmt.Sprintf(`{"text":"%s"}`, text)
	msg := &Message{
		ReceiveIdType: string(receiveIdType),
		ReceiveId:     receiveId,
		MsgType:       string(MsgTypeText),
		Content:       content,
	}
	return s.Send(ctx, msg)
}
