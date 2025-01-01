package main

import (
	"best-ticker/config"
	"best-ticker/message"
	binanceSpot "github.com/dictxwang/go-binance"
	binanceFutures "github.com/dictxwang/go-binance/futures"
	"github.com/drinkthere/okx/events/public"
)

func startTickerMessage() {
	if globalConfig.Service == config.OkxExchange {

		// 监听okx行情信息
		okxFuturesTickerChan := make(chan *public.Tickers)
		okxSpotTickerChan := make(chan *public.Tickers)
		message.StartOkxTickerWs(&globalConfig, &globalContext, okxFuturesTickerChan, okxSpotTickerChan)

		// 开启Okx行情数据收集整理
		message.StartGatherOkxFuturesTicker(okxFuturesTickerChan, &globalContext)
		message.StartGatherOkxSpotTicker(okxSpotTickerChan, &globalContext)

		// 监听Okx Trades数据
		message.StartOkxTradesWs(&globalConfig, &globalContext)
	} else if globalConfig.Service == config.BinanceExchange {

		// 监听binance行情信息并收集整理
		binanceSpotTickerChan := make(chan *binanceSpot.WsBookTickerEvent)
		binanceFuturesTickerChan := make(chan *binanceFutures.WsBookTickerEvent)
		message.StartBinanceMarketWs(&globalConfig, &globalContext, binanceFuturesTickerChan, binanceSpotTickerChan)

		// 开启Okx行情数据收集整理
		message.StartGatherBinanceSpotBookTicker(binanceSpotTickerChan, &globalContext)
		message.StartGatherBinanceFuturesBookTicker(binanceFuturesTickerChan, &globalContext)
	}
}
