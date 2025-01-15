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

type BinanceConfig struct {
	BinanceAPIKey    string
	BinanceSecretKey string
}

type Source struct {
	Exchange      Exchange      // 交易所
	IP            string        // local IP to connect to OKx websocket
	Colo          bool          // is co-location with Okx
	Channels      []Channel     // 改IP下需要订阅的channel
	OkxConfig     OkxConfig     // Okx配置
	BinanceConfig BinanceConfig // Binance配置
}

type Config struct {
	// 日志配置
	LogLevel zapcore.Level
	LogPath  string

	Service Exchange // Binance就是binance的best-ticker服务，Okx就是okx的best-ticker服务

	Sources             []Source // ip、channel等配置信息
	OkxTickerZMQIPC     string
	OkxOrderBookZMQIPC  string
	BinanceTickerZMQIPC string
	InstIDs             []string // 永续交易对
	SpotInstIDs         []string // 现货交易对
	PprofListenAddress  string   // pprof
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
