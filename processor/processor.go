package processor

import (
	"bufio"
	"fmt"
	"os"

	"go.uber.org/zap"

	"deepseek_bot/api"
	"deepseek_bot/db"
	"deepseek_bot/loger"
	"deepseek_bot/utils"
)

type Config struct {
	DeepSeekAPIKey string
	DeepSeekAPIURL string
}

func HandleInput(config Config) {
	sessionID := generateSessionID()
	loger.Loger.Info("[Input]new session started", zap.String("session_id", sessionID))

	scanner := bufio.NewScanner(os.Stdin)
	loger.Loger.Info("[Input]waiting for input...")

	for scanner.Scan() {
		content := scanner.Text()
		if content == "" {
			continue
		}

		loger.Loger.Info("[Input]received message", zap.String("content", content))

		id, err := db.InsertMessage(sessionID, content)
		if err != nil {
			loger.Loger.Error("[Input]failed to insert message", zap.Error(err))
			continue
		}

		loger.Loger.Info("[Input]message saved to database", zap.Int64("id", id))

		history, err := db.GetSessionHistory(sessionID)
		if err != nil {
			loger.Loger.Error("[Processor]failed to get session history", zap.Error(err))
			continue
		}

		loger.Loger.Info("[Processor]sending request to deepseek", zap.Int("history_length", len(history)))

		apiHistory := make([]api.ChatMessage, len(history))
		for i, msg := range history {
			apiHistory[i] = api.ChatMessage{Role: msg.Role, Content: msg.Content}
		}

		reply, err := api.GetReply(api.DeepSeekConfig{
			APIKey: config.DeepSeekAPIKey,
			APIURL: config.DeepSeekAPIURL,
		}, apiHistory, content)

		if err != nil {
			loger.Loger.Error("[Processor]failed to get reply from deepseek", zap.Int("id", int(id)), zap.Error(err))
			continue
		}

		err = db.UpdateReply(int(id), reply)
		if err != nil {
			loger.Loger.Error("[Processor]failed to update reply", zap.Int("id", int(id)), zap.Error(err))
			continue
		}

		loger.Loger.Info("[Processor]message processed successfully")
		fmt.Println("\n" + utils.RenderMarkdown(reply))
		loger.Loger.Info("[Processor]reply saved to database")
	}

	if err := scanner.Err(); err != nil {
		loger.Loger.Error("[Input]error reading input", zap.Error(err))
	}
}

func generateSessionID() string {
	return fmt.Sprintf("%d", os.Getpid())
}
