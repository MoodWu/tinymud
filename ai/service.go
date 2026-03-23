package ai

import "context"

type AIService struct {
	Client *Client

	// 限制并发
	Sem chan struct{}
}

func (s *AIService) Chat(ctx context.Context, msgs []Message) (string, error) {
	// 获取“令牌”
	s.Sem <- struct{}{}
	defer func() { <-s.Sem }()

	return s.Client.Chat(ctx, msgs)
}
