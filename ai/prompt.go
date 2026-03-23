// ai/prompt.go
package ai

import "fmt"

func BuildNPCPrompt(npcName, personality, playerInput string) []Message {
	system := fmt.Sprintf(
		"You are %s, %s. Reply concisely in character.",
		npcName, personality,
	)

	return []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: playerInput},
	}
}
