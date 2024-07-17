package message

import (
	"best-ticker/config"
	"best-ticker/container"
	"best-ticker/context"
	"best-ticker/utils/logger"
	"github.com/drinkthere/okx/events/public"
	"github.com/drinkthere/okx/models/market"
	"math/rand"
)

func convertToObMsg(instType config.InstrumentType, channel config.Channel, instID string,
	isSnapshot bool, msg *market.OrderBookWs) container.OrderBookMsg {
	return container.OrderBookMsg{
		Exchange:     config.OkxExchange,
		InstType:     instType,
		Channel:      channel,
		InstID:       instID,
		IsSnapshot:   true,
		OrderBookMsg: msg,
	}
}

func convertDepthToOkxTickerMessage(instType config.InstrumentType, instID string, channel config.Channel,
	bestBid *market.OrderBookEntity, bestAsk *market.OrderBookEntity, obUpdateTime int64) container.TickerWrapper {
	return container.TickerWrapper{
		Exchange:     config.OkxExchange,
		InstType:     instType,
		InstID:       instID,
		Channel:      channel,
		AskPrice:     bestAsk.DepthPrice,
		AskSize:      bestAsk.Size,
		BidPrice:     bestBid.DepthPrice,
		BidSize:      bestBid.Size,
		UpdateTimeMs: obUpdateTime,
	}
}

func StartGatherOkxFuturesDepthWs(depthChan chan *public.OrderBook, globalContext *context.GlobalContext) {

	r := rand.New(rand.NewSource(2))
	go func() {
		defer func() {
			logger.Warn("[GatherFDepth] Okx Futures Ticker Gather Exited.")
		}()
		for {
			s := <-depthChan
			for _, b := range s.Books {
				// update orderbook
				chRaw, _ := s.Arg.Get("channel")
				ch := config.Channel(chRaw.(string))
				instIDRaw, _ := s.Arg.Get("instId")
				instID := instIDRaw.(string)

				isSnapshot := false
				if s.Action == "snapshot" {
					isSnapshot = true
				}
				obMsg := convertToObMsg(config.FuturesInstrument, ch, instID, isSnapshot, b)
				globalContext.OkxFuturesBboComposite.UpdateOrderBook(obMsg)

				// b.TS > ticker.TS && ticker changed, update ticker
				ticker := globalContext.OkxFuturesTickerComposite.GetTicker(instID)
				orderBook := globalContext.OkxFuturesBboComposite.GetOrderBook(instID)

				obUpdateTime := orderBook.UpdateTime()
				if obUpdateTime > ticker.UpdateTimeMs {
					bestBid := orderBook.BestBid()
					bestAsk := orderBook.BestAsk()
					if bestBid.DepthPrice != ticker.AskPrice || bestAsk.DepthPrice != ticker.BidPrice {
						tickerMsg := convertDepthToOkxTickerMessage(config.FuturesInstrument, instID, ch, bestBid, bestAsk, obUpdateTime)
						result := globalContext.OkxFuturesTickerComposite.UpdateTicker(tickerMsg)
						if result {
							// todo 更新zmq消息
							if r.Int31n(10000) < 10000 {
								newT := globalContext.OkxFuturesTickerComposite.GetTicker(instID)
								logger.Info("[GatherFDepth] Okx Ticker is %+v\n", newT)
							}
						}
					}
				}
			}
		}
	}()

	logger.Info("[GatherFTicker] Start Gather Okx Futures Ticker")
}

func StartGatherOkxSpotDepthWs(tickChan chan *public.OrderBook, globalContext *context.GlobalContext) {
	//
	//r := rand.New(rand.NewSource(2))
	//go func() {
	//	defer func() {
	//		logger.Warn("[GatherFTicker] Okx Spot Ticker Gather Exited.")
	//	}()
	//	for {
	//		s := <-tickChan
	//		for _, t := range s.Tickers {
	//			tickerMsg := convertToOkxTickerMessage(t)
	//			result := globalContext.OkxSpotTickerComposite.UpdateTicker(tickerMsg)
	//			if result {
	//				// todo 更新zmq消息
	//			}
	//		}
	//		if r.Int31n(10000) < 5 && len(s.Tickers) > 0 {
	//			logger.Info("[GatherSTicker] Receive Okx Spot Ticker %+v\n", s.Tickers[0])
	//		}
	//	}
	//}()
	//
	//logger.Info("[GatherSTicker] Start Gather Okx Spot Ticker")
}
