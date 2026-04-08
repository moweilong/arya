package card

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/gookit/slog"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkcardkit "github.com/larksuite/oapi-sdk-go/v3/service/cardkit/v1"
)

// Service 卡片服务
type Service struct {
	larkClient *lark.Client

	// cardSequence 卡片操作序号管理（必须严格递增）
	cardSequence map[string]int32
	cardSeqMu    sync.Mutex
}

// CardService 卡片服务别名
type CardService = Service

// NewCardService 创建卡片服务
func NewCardService(larkClient *lark.Client) *CardService {
	return &CardService{
		larkClient:   larkClient,
		cardSequence: make(map[string]int32),
	}
}

// GetNextSequence 获取并递增卡片操作序号（必须严格递增）
func (s *CardService) GetNextSequence(cardID string) int32 {
	s.cardSeqMu.Lock()
	defer s.cardSeqMu.Unlock()

	seq := s.cardSequence[cardID] + 1
	s.cardSequence[cardID] = seq
	return seq
}

// Create 创建卡片实体，返回 card_id
func (s *CardService) Create(ctx context.Context, card *Card) (string, error) {
	cardDataStr := card.String()
	if cardDataStr == "" || cardDataStr == "{}" {
		return "", fmt.Errorf("卡片数据为空")
	}

	req := larkcardkit.NewCreateCardReqBuilder().
		Body(larkcardkit.NewCreateCardReqBodyBuilder().
			Type("card_json").
			Data(cardDataStr).
			Build()).
		Build()

	resp, err := s.larkClient.Cardkit.V1.Card.Create(ctx, req)
	if err != nil {
		slog.Errorf("创建飞书卡片失败: %v", err)
		return "", err
	}

	if !resp.Success() {
		errMsg := fmt.Sprintf("logId: %s, error response: %s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		slog.Error(errMsg)
		return "", fmt.Errorf("%s", errMsg)
	}

	if resp.Data == nil || resp.Data.CardId == nil {
		return "", fmt.Errorf("响应数据为空")
	}

	return *resp.Data.CardId, nil
}

// CreateWithContent 创建简单的 Markdown 卡片
func (s *CardService) CreateWithContent(ctx context.Context, content string, title string) (string, error) {
	builder := NewBuilder()
	if title != "" {
		builder.WithTitle(title)
	}
	card := builder.Build()
	card.Body.Elements = []Element{
		{
			Tag:       "markdown",
			ElementId: "markdown",
			Content:   content,
		},
	}
	return s.Create(ctx, card)
}

// UpdateContent 更新卡片内容
func (s *CardService) UpdateContent(ctx context.Context, cardID, content string) error {
	seq := int(s.GetNextSequence(cardID))
	uuidStr := uuid.New().String()

	req := larkcardkit.NewContentCardElementReqBuilder().
		CardId(cardID).
		ElementId("markdown").
		Body(larkcardkit.NewContentCardElementReqBodyBuilder().
			Uuid(uuidStr).
			Content(content).
			Sequence(seq).
			Build()).
		Build()

	resp, err := s.larkClient.Cardkit.V1.CardElement.Content(ctx, req)
	if err != nil {
		slog.Errorf("更新飞书卡片失败: %v", err)
		return err
	}

	if !resp.Success() {
		errMsg := fmt.Sprintf("logId: %s, error response: %s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		slog.Error(errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

// UpdateElement 更新指定元素
func (s *CardService) UpdateElement(ctx context.Context, cardID, elementID, content string) error {
	seq := int(s.GetNextSequence(cardID))
	uuidStr := uuid.New().String()

	req := larkcardkit.NewContentCardElementReqBuilder().
		CardId(cardID).
		ElementId(elementID).
		Body(larkcardkit.NewContentCardElementReqBodyBuilder().
			Uuid(uuidStr).
			Content(content).
			Sequence(seq).
			Build()).
		Build()

	resp, err := s.larkClient.Cardkit.V1.CardElement.Content(ctx, req)
	if err != nil {
		slog.Errorf("更新飞书卡片元素失败: %v", err)
		return err
	}

	if !resp.Success() {
		errMsg := fmt.Sprintf("logId: %s, error response: %s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		slog.Error(errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}
