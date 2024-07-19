package main

import (
	"best-ticker/message"
)

func startDepthMessage() {
	// 监听okx行情信息
	message.StartOkxDepthWs(&globalConfig, &globalContext)
}
