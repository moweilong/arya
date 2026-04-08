package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReceiveIdType(t *testing.T) {
	tests := []struct {
		name     string
		idType   ReceiveIdType
		expected string
	}{
		{"user_id", ReceiveIdTypeUserID, "user_id"},
		{"open_id", ReceiveIdTypeOpenID, "open_id"},
		{"chat_id", ReceiveIdTypeChatID, "chat_id"},
		{"email", ReceiveIdTypeEmail, "email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.idType))
		})
	}
}

func TestMsgType(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MsgType
		expected string
	}{
		{"text", MsgTypeText, "text"},
		{"post", MsgTypePost, "post"},
		{"interactive", MsgTypeInteractive, "interactive"},
		{"image", MsgTypeImage, "image"},
		{"file", MsgTypeFile, "file"},
		{"audio", MsgTypeAudio, "audio"},
		{"media", MsgTypeMedia, "media"},
		{"sticker", MsgTypeSticker, "sticker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.msgType))
		})
	}
}

func TestNewMessageService(t *testing.T) {
	service := NewMessageService(nil)
	assert.NotNil(t, service)
}

func TestMessage_Fields(t *testing.T) {
	msg := &Message{
		ReceiveIdType: string(ReceiveIdTypeUserID),
		ReceiveId:     "test-user-id",
		MsgType:       string(MsgTypeInteractive),
		Content:       `{"card":{}}`,
	}

	assert.Equal(t, "user_id", msg.ReceiveIdType)
	assert.Equal(t, "test-user-id", msg.ReceiveId)
	assert.Equal(t, "interactive", msg.MsgType)
	assert.Equal(t, `{"card":{}}`, msg.Content)
}
