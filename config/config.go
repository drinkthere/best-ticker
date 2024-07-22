package config

import (
	"encoding/json"
	"go.uber.org/zap/zapcore"
	"os"
)

type OkxConfig struct {
	OkxAPIKey    string
	OkxSecretKey string
	OkxPassword  string
}

type Source struct {
	IP        string    // local IP to connect to OKx websocket
	Colo      bool      // is co-location with Okx
	Channels  []Channel // 改IP下需要订阅的channel
	OkxConfig OkxConfig // Okx配置
}

type Config struct {
	// 日志配置
	LogLevel zapcore.Level
	LogPath  string

	Sources         []Source // 公网IP
	TickerZMQIPC    string
	OrderBookZMQIPC string
	InstIDs         []string // 要套利的交易对
}

func LoadConfig(filename string) *Config {
	config := new(Config)
	reader, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	// 加载配置
	decoder := json.NewDecoder(reader)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	return config
}
