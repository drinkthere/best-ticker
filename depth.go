package main

import (
	"best-ticker/config"
	"best-ticker/message"
)

func startDepthMessage() {
	if globalConfig.Service == config.OkxExchange {
		// 监听okx行情信息
		message.StartOkxDepthWs(&globalConfig, &globalContext)
	}
}
