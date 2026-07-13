package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"deepseek_bot/db"
	"deepseek_bot/loger"
	"deepseek_bot/processor"
)

type Config struct {
	DeepSeekAPIKey string `json:"deepseek_api_key"`
	DeepSeekAPIURL string `json:"deepseek_api_url"`
	DBPath         string `json:"db_path"`
}

func loadConfig(path string) (Config, error) {
	var config Config
	file, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}

func main() {
	loger.InitLog()
	loger.Loger.Info("[Main]starting deepseek_bot...")

	config, err := loadConfig("./config.json")
	if err != nil {
		loger.Loger.Fatal("[Main]failed to load config", zap.Error(err))
	}

	err = db.InitDB(config.DBPath)
	if err != nil {
		loger.Loger.Fatal("[Main]failed to initialize database", zap.Error(err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		loger.Loger.Info("[Main]received shutdown signal, cleaning up...")
		db.CloseDB()
		loger.CloseLog()
		os.Exit(0)
	}()

	processor.HandleInput(processor.Config{
		DeepSeekAPIKey: config.DeepSeekAPIKey,
		DeepSeekAPIURL: config.DeepSeekAPIURL,
	})
}
