package main

import (
	"best-ticker/message"
	"github.com/drinkthere/okx/events/public"
)

func startTickerMessage() {
	// 监听okx行情信息
	okxFuturesTickerChan := make(chan *public.Tickers)
	okxSpotTickerChan := make(chan *public.Tickers)
	message.StartOkxTickerWs(&globalConfig, &globalContext, okxFuturesTickerChan, okxSpotTickerChan)

	// 开启Okx行情数据收集整理
	message.StartGatherOkxFuturesTicker(okxFuturesTickerChan, &globalContext)
	message.StartGatherOkxSpotTicker(okxSpotTickerChan, &globalContext)

	// 开启
}
