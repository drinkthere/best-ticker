package main

import (
	"best-ticker/message"
	"github.com/drinkthere/okx/events/public"
)

func startDepthMessage() {
	// 监听okx行情信息
	okxFuturesDepthChan := make(chan *public.OrderBook)
	okxSpotDepthChan := make(chan *public.OrderBook)
	message.StartOkxDepthWs(&globalConfig, &globalContext, okxFuturesDepthChan, okxSpotDepthChan)

	//// 开启Okx行情数据收集整理
	//message.StartGatherOkxFuturesDepthWs(okxFuturesDepthChan, &globalContext)
	//message.StartGatherOkxSpotDepthWs(okxSpotDepthChan, &globalContext)
}
