// npc/npc.go
package npc

import (
	"context"
	"game/ai"
)

type NPC struct {
	Name        string
	Personality string
	Service     *ai.AIService
	Memory      map[string]*Memory // key: playerID
}
type Memory struct {
	Messages []ai.Message
	MaxSize  int
}

func (n *NPC) Talk(ctx context.Context, playerID, input string) (string, error) {
	mem := n.getMemory(playerID)

	msgs := ai.BuildNPCPrompt(n.Name, n.Personality, input, mem.Messages)
	reply, err := n.Service.Chat(ctx, msgs)

	mem.Messages = append(mem.Messages, ai.Message{Role: "user", Content: input})
	mem.Messages = append(mem.Messages, ai.Message{Role: "assistant", Content: reply})

	if len(mem.Messages) > mem.MaxSize {
		mem.Messages = mem.Messages[len(mem.Messages)-mem.MaxSize:]
	}

	return reply, err
}

func (n *NPC) getMemory(playerID string) *Memory {
	m, ok := n.Memory[playerID]
	if !ok {
		m = &Memory{MaxSize: 10}
		n.Memory[playerID] = m
	}
	return m
}
