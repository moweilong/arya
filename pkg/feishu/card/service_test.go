package card

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCardService(t *testing.T) {
	service := NewCardService(nil)
	assert.NotNil(t, service)
	assert.NotNil(t, service.cardSequence)
}

func TestCardService_GetNextSequence(t *testing.T) {
	service := NewCardService(nil)

	// 测试同一卡片的序号递增
	cardID := "test-card-id"

	seq1 := service.GetNextSequence(cardID)
	assert.Equal(t, int32(1), seq1)

	seq2 := service.GetNextSequence(cardID)
	assert.Equal(t, int32(2), seq2)

	seq3 := service.GetNextSequence(cardID)
	assert.Equal(t, int32(3), seq3)
}

func TestCardService_GetNextSequence_DifferentCards(t *testing.T) {
	service := NewCardService(nil)

	cardID1 := "card-1"
	cardID2 := "card-2"

	// 不同卡片的序号应该独立
	seq1 := service.GetNextSequence(cardID1)
	assert.Equal(t, int32(1), seq1)

	seq2 := service.GetNextSequence(cardID2)
	assert.Equal(t, int32(1), seq2)

	seq3 := service.GetNextSequence(cardID1)
	assert.Equal(t, int32(2), seq3)
}

func TestCardService_GetNextSequence_Concurrent(t *testing.T) {
	service := NewCardService(nil)
	cardID := "concurrent-card"

	var wg sync.WaitGroup
	results := make(chan int32, 100)

	// 并发获取序号
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seq := service.GetNextSequence(cardID)
			results <- seq
		}()
	}

	wg.Wait()
	close(results)

	// 收集所有序号
	sequences := make(map[int32]bool)
	for seq := range results {
		sequences[seq] = true
	}

	// 验证序号从1到100都是唯一的
	assert.Equal(t, 100, len(sequences))
	for i := int32(1); i <= 100; i++ {
		assert.True(t, sequences[i], "序号 %d 应该存在", i)
	}
}
