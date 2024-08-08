package message

import (
	"best-ticker/config"
	"best-ticker/container"
	"best-ticker/context"
	"best-ticker/utils/logger"
	binanceSpot "github.com/dictxwang/go-binance"
	binanceFutures "github.com/dictxwang/go-binance/futures"
	"math/rand"
	"strconv"
	"time"
)

func StartGatherBinanceSpotBookTicker(tickChan chan *binanceSpot.WsBookTickerEvent, globalContext *context.GlobalContext) {

	r := rand.New(rand.NewSource(2))
	go func() {
		defer func() {
			logger.Warn("[GatherBSTickerErr] Binance Spot Ticker Gather Exited.")
		}()
		for t := range tickChan {
			tickerMsg := convertSpotEventToBinanceTickerMessage(t)
			result := globalContext.BinanceSpotTickerComposite.UpdateTicker(tickerMsg)
			if result {
				globalContext.TickerUpdateChan <- &tickerMsg
			}

			if r.Int31n(10000) < 2 {
				logger.Info("[GatherBSTickerErr] Receive Binance Spot Ticker %+v", t)
			}
		}
	}()

	logger.Info("[GatherBSTickerErr] Start Gather Binance Spot Book Ticker")
}

func StartGatherBinanceFuturesBookTicker(tickChan chan *binanceFutures.WsBookTickerEvent, globalContext *context.GlobalContext) {

	r := rand.New(rand.NewSource(2))
	go func() {
		defer func() {
			logger.Warn("[GatherBFTickerErr] Binance Futures Ticker Gather Exited.")
		}()
		for t := range tickChan {
			tickerMsg := convertFuturesEventToBinanceTickerMessage(t)
			result := globalContext.BinanceFuturesTickerComposite.UpdateTicker(tickerMsg)
			if result {
				globalContext.TickerUpdateChan <- &tickerMsg
			}

			if r.Int31n(10000) < 2 {
				logger.Info("[GatherBFTickerErr] Receive Binance Futures Ticker %+v", t)
			}
		}
	}()

	logger.Info("[GatherBFTickerErr] Start Gather Binance Futures Flash Ticker")
}

func convertSpotEventToBinanceTickerMessage(ticker *binanceSpot.WsBookTickerEvent) container.TickerWrapper {
	bestAskPrice, _ := strconv.ParseFloat(ticker.BestAskPrice, 64)
	bestAskQty, _ := strconv.ParseFloat(ticker.BestAskQty, 64)
	bestBidPrice, _ := strconv.ParseFloat(ticker.BestBidPrice, 64)
	bestBidQty, _ := strconv.ParseFloat(ticker.BestBidQty, 64)
	return container.TickerWrapper{
		Exchange:     config.BinanceExchange,
		InstType:     config.SpotInstrument,
		InstID:       ticker.Symbol,
		Channel:      config.NoChannel,
		AskPrice:     bestAskPrice,
		AskSize:      bestAskQty,
		BidPrice:     bestBidPrice,
		BidSize:      bestBidQty,
		UpdateTimeMs: time.Now().UnixMilli(),
		UpdateID:     ticker.UpdateID,
	}
}

func convertFuturesEventToBinanceTickerMessage(ticker *binanceFutures.WsBookTickerEvent) container.TickerWrapper {
	bestAskPrice, _ := strconv.ParseFloat(ticker.BestAskPrice, 64)
	bestAskQty, _ := strconv.ParseFloat(ticker.BestAskQty, 64)
	bestBidPrice, _ := strconv.ParseFloat(ticker.BestBidPrice, 64)
	bestBidQty, _ := strconv.ParseFloat(ticker.BestBidQty, 64)
	return container.TickerWrapper{
		Exchange:     config.BinanceExchange,
		InstType:     config.SpotInstrument,
		InstID:       ticker.Symbol,
		Channel:      config.NoChannel,
		AskPrice:     bestAskPrice,
		AskSize:      bestAskQty,
		BidPrice:     bestBidPrice,
		BidSize:      bestBidQty,
		UpdateTimeMs: ticker.Time,
		UpdateID:     ticker.UpdateID,
	}
}
