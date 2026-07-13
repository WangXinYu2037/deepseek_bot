package loger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Loger *zap.Logger
var ModelName = "[Loger]"
var logFile *os.File

func InitLog() {
	ConsoleLogConfig := zap.NewProductionEncoderConfig()
	ConsoleLogConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	ConsoleLogConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	var ConsoleLogerEcoder zapcore.Encoder
	ConsoleLogerEcoder = zapcore.NewConsoleEncoder(ConsoleLogConfig)

	logDir := "./data/log"
	_, err := os.Stat(logDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0775)
		if err != nil {
			panic(err)
		}
	}

	logFilePath := filepath.Join(logDir, time.Now().Format("2006-01-02_15_04_05")+".log")
	logFile, err = os.Create(logFilePath)
	if err != nil {
		panic(err)
	}

	FileConfig := zap.NewProductionEncoderConfig()
	FileConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	FileEcoder := zapcore.NewConsoleEncoder(FileConfig)
	core := zapcore.NewTee(
		zapcore.NewCore(ConsoleLogerEcoder, os.Stdout, zap.InfoLevel),
		zapcore.NewCore(FileEcoder, logFile, zap.InfoLevel),
	)
	Loger = zap.New(core)
	Loger.Info("[Loger]OK")
}

func CloseLog() {
	if Loger != nil {
		Loger.Sync()
	}
	if logFile != nil {
		logFile.Close()
	}
}
