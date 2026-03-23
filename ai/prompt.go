// ai/prompt.go
package ai

import "fmt"

func BuildNPCPrompt(npcName, personality, playerInput string, history []Message) []Message {
	system := fmt.Sprintf(
		"You are %s, %s. Reply concisely in character.",
		npcName, personality,
	)

	ret := []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: playerInput},
	}
	ret = append(ret, history...)
	return ret

}
