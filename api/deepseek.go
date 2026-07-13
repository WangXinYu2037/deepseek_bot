package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"deepseek_bot/loger"
)

type DeepSeekConfig struct {
	APIKey string
	APIURL string
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func GetReply(config DeepSeekConfig, history []ChatMessage, message string) (string, error) {
	messages := append(history, ChatMessage{
		Role:    "user",
		Content: message,
	})

	reqBody := ChatCompletionRequest{
		Model:    "deepseek-v4-flash",
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		loger.Loger.Error("[DeepSeek]failed to marshal request", zap.Error(err))
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", config.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		loger.Loger.Error("[DeepSeek]failed to create request", zap.Error(err))
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		loger.Loger.Error("[DeepSeek]failed to send request", zap.Error(err))
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		loger.Loger.Error("[DeepSeek]api returned status code", zap.Int("code", resp.StatusCode))
		return "", fmt.Errorf("api returned status code %d", resp.StatusCode)
	}

	var response ChatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		loger.Loger.Error("[DeepSeek]failed to decode response", zap.Error(err))
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(response.Choices) == 0 {
		loger.Loger.Error("[DeepSeek]no choices in response")
		return "", fmt.Errorf("no choices in response")
	}

	reply := response.Choices[0].Message.Content
	loger.Loger.Info("[DeepSeek]received reply")
	return reply, nil
}
