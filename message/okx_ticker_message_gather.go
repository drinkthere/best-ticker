package message

import (
	"best-ticker/config"
	"best-ticker/container"
	"best-ticker/context"
	"best-ticker/utils"
	"best-ticker/utils/logger"
	"github.com/drinkthere/okx/events/public"
	"github.com/drinkthere/okx/models/market"
	"math/rand"
	"time"
)

func convertToOkxTickerMessage(ticker *market.Ticker) container.TickerWrapper {
	return container.TickerWrapper{
		Exchange:     config.OkxExchange,
		InstType:     utils.ConvertToStdInstType(config.OkxExchange, string(ticker.InstType)),
		InstID:       ticker.InstID,
		Channel:      config.NoChannel,
		AskPrice:     float64(ticker.AskPx),
		AskSize:      float64(ticker.AskSz),
		BidPrice:     float64(ticker.BidPx),
		BidSize:      float64(ticker.BidSz),
		UpdateTimeMs: time.Time(ticker.TS).UnixMilli(),
	}
}

func StartGatherOkxFuturesTicker(tickChan chan *public.Tickers, globalContext *context.GlobalContext) {

	r := rand.New(rand.NewSource(2))
	go func() {
		defer func() {
			logger.Warn("[GatherFTicker] Okx Futures Ticker Gather Exited.")
		}()
		for {
			s := <-tickChan
			for _, t := range s.Tickers {
				tickerMsg := convertToOkxTickerMessage(t)
				result := globalContext.OkxFuturesTickerComposite.UpdateTicker(tickerMsg)
				if result {
					logger.Debug("FUTURES|CHANNEL|Ticker")
					globalContext.TickerUpdateChan <- &tickerMsg
				}
			}
			if r.Int31n(10000) < 5 && len(s.Tickers) > 0 {
				logger.Info("[GatherFTicker] Receive Okx Futures Ticker %+v\n", s.Tickers[0])
			}
		}
	}()

	logger.Info("[GatherFTicker] Start Gather Okx Futures Ticker")
}

func StartGatherOkxSpotTicker(tickChan chan *public.Tickers, globalContext *context.GlobalContext) {

	r := rand.New(rand.NewSource(2))
	go func() {
		defer func() {
			logger.Warn("[GatherFTicker] Okx Spot Ticker Gather Exited.")
		}()
		for {
			s := <-tickChan
			for _, t := range s.Tickers {
				tickerMsg := convertToOkxTickerMessage(t)
				result := globalContext.OkxSpotTickerComposite.UpdateTicker(tickerMsg)
				if result {
					logger.Debug("SPOT|CHANNEL|Ticker")
					globalContext.TickerUpdateChan <- &tickerMsg
				}
			}
			if r.Int31n(10000) < 5 && len(s.Tickers) > 0 {
				logger.Info("[GatherSTicker] Receive Okx Spot Ticker %+v\n", s.Tickers[0])
			}
		}
	}()

	logger.Info("[GatherSTicker] Start Gather Okx Spot Ticker")
}
