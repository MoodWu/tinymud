// npc/npc.go
package npc

import "game/ai"

type NPC struct {
	Name        string
	Personality string
	Client      *ai.Client
}

func (n *NPC) Talk(input string) (string, error) {
	msgs := ai.BuildNPCPrompt(n.Name, n.Personality, input)
	return n.Client.Chat(msgs)
}
