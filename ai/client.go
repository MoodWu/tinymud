// ai/client.go
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type Client struct {
	APIKey string
	URL    string
	Model  string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	body := map[string]interface{}{
		"model":    c.Model,
		"messages": messages,
	}

	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", c.URL, bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	json.NewDecoder(resp.Body).Decode(&result)

	return result.Choices[0].Message.Content, nil
}
