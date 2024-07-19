package message

import (
	"best-ticker/client"
	"best-ticker/config"
	"best-ticker/container"
	"best-ticker/context"
	"best-ticker/utils/logger"
	"github.com/drinkthere/okx/events"
	"github.com/drinkthere/okx/events/public"
	"github.com/drinkthere/okx/models/market"
	wsRequestPublic "github.com/drinkthere/okx/requests/ws/public"
	"math/rand"
	"time"
)

func StartOkxDepthWs(cfg *config.Config, globalContext *context.GlobalContext) {
	// 循环不同的IP，监听不同的depth channel
	startOkxFuturesDepths(&cfg.OkxConfig, globalContext, false, "", config.BboTbtChannel)
	logger.Info("[FDepthWebSocket] Start Listen Okx Futures Depth Channel")

	startOkxFuturesDepths(&cfg.OkxConfig, globalContext, true, "192.168.14.38", config.Books50L2TbtChannel)
	logger.Info("[FDepthWebSocket] Start Listen Okx Futures Depth Channel")

	startOkxSpotDepths(&cfg.OkxConfig, globalContext, false, "", config.BboTbtChannel)
	logger.Info("[SDepthWebSocket] Start Listen Okx Spot Depth Channel")

	startOkxSpotDepths(&cfg.OkxConfig, globalContext, true, "192.168.14.38", config.Books50L2TbtChannel)
	logger.Info("[SDepthWebSocket] Start Listen Okx Spot Depth Channel")
}

func startOkxFuturesDepths(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	isColo bool, localIP string, subCh config.Channel) {

	r := rand.New(rand.NewSource(2))
	depthChan := make(chan *public.OrderBook)
	go func() {
		defer func() {
			logger.Warn("[FDepthWebSocket] Okx Futures Channel Listening Exited.")
		}()
		for {
		ReConnect:
			errChan := make(chan *events.Error)
			subChan := make(chan *events.Subscribe)
			uSubChan := make(chan *events.Unsubscribe)
			loginCh := make(chan *events.Login)
			successCh := make(chan *events.Success)

			var okxClient = client.OkxClient{}
			okxClient.Init(cfg, isColo, localIP)
			if subCh == config.BooksL2TbtChannel || subCh == config.Books50L2TbtChannel {
				err := okxClient.Client.Ws.Private.Login()
				if err != nil {
					logger.Fatal("[FDepthWebSocket] Fail To Login, Error: %s", err.Error())
				}
			}

			okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
			currCh := string(subCh)
		ReSubscribe:
			for _, instID := range globalContext.InstrumentComposite.InstIDs {
				err := okxClient.Client.Ws.Public.OrderBook(wsRequestPublic.OrderBook{
					InstID:  instID,
					Channel: currCh,
				}, depthChan)
				if err != nil {
					logger.Fatal("[FDepthWebSocket] Fail To Listen Futures Depth For %s %s, %s, %s", localIP, instID, currCh, err.Error())
				} else {
					logger.Info("[FDepthWebSocket] Futures Depth WebSocket Has Established For %s %s %s", localIP, instID, currCh)
				}
			}
			for {
				select {
				case sub := <-subChan:
					instID, _ := sub.Arg.Get("instID")
					logger.Info("[FDepthWebSocket] Futures Subscribe %s %s %s", localIP, instID, currCh)
				case usub := <-uSubChan:
					instID, _ := usub.Arg.Get("instID")
					logger.Info("[FDepthWebSocket] Futures Unsubscribe %s %s %s", localIP, instID, currCh)
					time.Sleep(time.Second * 30)
					goto ReSubscribe
				case err := <-errChan:
					logger.Error("[FDepthWebSocket] Futures Occur Some Error %s %s %+v", localIP, currCh, err)
				case s := <-successCh:
					logger.Info("[FDepthWebSocket] Futures Receive Success: %s %s %+v", localIP, currCh, s)
				case s := <-depthChan:
					for _, b := range s.Books {
						// update orderbook
						chRaw, _ := s.Arg.Get("channel")
						ch := config.Channel(chRaw.(string))
						instIDRaw, _ := s.Arg.Get("instId")
						instID := instIDRaw.(string)
						action := s.Action

						instType := config.FuturesInstrument
						obMsg := convertToObMsg(instType, ch, instID, action, b)
						result := updateOrderBook(instType, ch, obMsg, globalContext)
						if !result {
							okxClient.Client.Ws.Public.UOrderBook(wsRequestPublic.OrderBook{
								InstID:  instID,
								Channel: currCh,
							})
						} else {
							if r.Int31n(10000) < 5 {
								orderBook := getOrderBook(instID, instType, ch, globalContext)
								logger.Info("[GatherFDepth] orderBooks.bids is %+v, orderBooks.asks is %+v", orderBook.BestBid(), orderBook.BestAsk())
							}
							checkToUpdateTicker(instID, instType, ch, globalContext)
						}
					}
				case b := <-okxClient.Client.Ws.DoneChan:
					logger.Info("[FDepthWebSocket] Futures End %s %s %v", localIP, currCh, b)
					// 暂停一秒再跳出，避免异常时频繁发起重连
					logger.Warn("[FDepthWebSocket] Will Reconnect Futures-Depth-WebSocket After 1 Second")
					time.Sleep(time.Second * 1)
					goto ReConnect
				}

			}
		}
	}()
}

func startOkxSpotDepths(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	isColo bool, localIP string, subCh config.Channel) {
	go func() {
		defer func() {
			logger.Warn("[SDepthWebSocket] Okx Spot Depth Listening Exited.")
		}()

		r := rand.New(rand.NewSource(2))
		depthChan := make(chan *public.OrderBook)
		for {
		ReConnect:
			errChan := make(chan *events.Error)
			subChan := make(chan *events.Subscribe)
			uSubChan := make(chan *events.Unsubscribe)
			loginCh := make(chan *events.Login)
			successCh := make(chan *events.Success)

			var okxClient = client.OkxClient{}
			okxClient.Init(cfg, isColo, localIP)
			if subCh == config.BooksL2TbtChannel || subCh == config.Books50L2TbtChannel {
				err := okxClient.Client.Ws.Private.Login()
				if err != nil {
					logger.Fatal("[SDepthWebSocket] Fail To Login, Error: %s", err.Error())
				}
			}
			okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
			currCh := string(subCh)
		ReSubscribe:
			for _, instID := range globalContext.InstrumentComposite.SpotInstIDs {
				err := okxClient.Client.Ws.Public.OrderBook(wsRequestPublic.OrderBook{
					InstID:  instID,
					Channel: currCh,
				}, depthChan)

				if err != nil {
					logger.Fatal("[SDepthWebSocket] Fail To Listen Futures Depth For %s %s %s %s", localIP, instID, currCh, err.Error())
				} else {
					logger.Info("[SDepthWebSocket] Futures Depth WebSocket Has Established For %s %s %s", localIP, instID, currCh)
				}
			}

			for {
				select {
				case sub := <-subChan:
					instID, _ := sub.Arg.Get("instID")
					logger.Info("[SDepthWebSocket] Spot Subscribe %s %s %s", localIP, instID, currCh)
				case usub := <-uSubChan:
					instID, _ := usub.Arg.Get("instID")
					logger.Info("[SDepthWebSocket] Futures Unsubscribe %s %s %s", localIP, instID, currCh)
					time.Sleep(time.Second * 30)
					goto ReSubscribe
				case err := <-errChan:
					logger.Error("[SDepthWebSocket] Spot Occur Some Error %s %s %+v", localIP, currCh, err)
				case s := <-successCh:
					logger.Info("[SDepthWebSocket] Spot Receive Success: %s %s %+v", localIP, currCh, s)
				case s := <-depthChan:
					for _, b := range s.Books {
						// update orderbook
						chRaw, _ := s.Arg.Get("channel")
						ch := config.Channel(chRaw.(string))
						instIDRaw, _ := s.Arg.Get("instId")
						instID := instIDRaw.(string)
						action := s.Action

						instType := config.SpotInstrument
						obMsg := convertToObMsg(instType, ch, instID, action, b)
						result := updateOrderBook(instType, ch, obMsg, globalContext)
						if !result {
							okxClient.Client.Ws.Public.UOrderBook(wsRequestPublic.OrderBook{
								InstID:  instID,
								Channel: currCh,
							})
						} else {
							if r.Int31n(10000) < 5 {
								orderBook := getOrderBook(instID, instType, ch, globalContext)
								logger.Info("[GatherFDepth] orderBooks.bids is %+v, orderBooks.asks is %+v", orderBook.BestBid(), orderBook.BestAsk())
							}
							checkToUpdateTicker(instID, instType, ch, globalContext)
						}
					}
				case b := <-okxClient.Client.Ws.DoneChan:
					logger.Info("[SDepthWebSocket] Spot End %s %s %v", localIP, currCh, b)
					// 暂停一秒再跳出，避免异常时频繁发起重连
					logger.Warn("[SDepthWebSocket] Will Reconnect Spot-Depth-WebSocket After 1 Second")
					time.Sleep(time.Second * 1)
					goto ReConnect
				}
			}
		}
	}()
}

func checkToUpdateTicker(instID string, instType config.InstrumentType, ch config.Channel, globalContext *context.GlobalContext) {
	// b.TS > ticker.TS && ticker changed, update ticker
	ticker := getTicker(instID, instType, globalContext)
	orderBook := getOrderBook(instID, instType, ch, globalContext)
	if orderBook == nil || ticker == nil {
		logger.Info("orderBook is nil or ticker is nil")
		return
	}

	obUpdateTime := orderBook.UpdateTime()
	if obUpdateTime > ticker.UpdateTimeMs {
		bestBid := orderBook.BestBid()
		bestAsk := orderBook.BestAsk()
		if bestBid.DepthPrice != ticker.AskPrice || bestAsk.DepthPrice != ticker.BidPrice {
			tickerMsg := convertDepthToOkxTickerMessage(config.FuturesInstrument, instID, ch, bestBid, bestAsk, obUpdateTime)
			result := updateTicker(instType, tickerMsg, globalContext)
			if result {
				// todo 更新zmq消息

			}
		}
	}
}

func convertToObMsg(instType config.InstrumentType, channel config.Channel, instID string,
	action string, msg *market.OrderBookWs) container.OrderBookMsg {
	return container.OrderBookMsg{
		Exchange:     config.OkxExchange,
		InstType:     instType,
		Channel:      channel,
		InstID:       instID,
		Action:       action,
		OrderBookMsg: msg,
	}
}

func convertDepthToOkxTickerMessage(instType config.InstrumentType, instID string, channel config.Channel,
	bestBid market.OrderBookEntity, bestAsk market.OrderBookEntity, obUpdateTime int64) container.TickerWrapper {
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

func getTicker(instID string, instType config.InstrumentType, globalContext *context.GlobalContext) *container.TickerWrapper {
	var ticker *container.TickerWrapper
	if instType == config.FuturesInstrument {
		ticker = globalContext.OkxFuturesTickerComposite.GetTicker(instID)
	} else {
		ticker = globalContext.OkxSpotTickerComposite.GetTicker(instID)
	}
	return ticker
}

func updateTicker(instType config.InstrumentType, tickerMsg container.TickerWrapper, globalContext *context.GlobalContext) bool {
	var result bool
	if instType == config.FuturesInstrument {
		result = globalContext.OkxFuturesTickerComposite.UpdateTicker(tickerMsg)
	} else {
		result = globalContext.OkxSpotTickerComposite.UpdateTicker(tickerMsg)
	}
	return result
}

func getOrderBook(instID string, instType config.InstrumentType, ch config.Channel, globalContext *context.GlobalContext) *container.OrderBook {

	var orderBook *container.OrderBook
	if instType == config.FuturesInstrument {
		if ch == config.BboTbtChannel {
			orderBook = globalContext.OkxFuturesBboComposite.GetOrderBook(instID)
		} else if ch == config.Books50L2TbtChannel {
			orderBook = globalContext.OkxFuturesBooks50L2Composite.GetOrderBook(instID)
		} else if ch == config.BooksL2TbtChannel {
			orderBook = globalContext.OkxFuturesL2Composite.GetOrderBook(instID)
		}
	} else {
		if ch == config.BboTbtChannel {
			orderBook = globalContext.OkxSpotBboComposite.GetOrderBook(instID)
		} else if ch == config.Books50L2TbtChannel {
			orderBook = globalContext.OkxSpotBooks50L2Composite.GetOrderBook(instID)
		} else if ch == config.BooksL2TbtChannel {
			orderBook = globalContext.OkxSpotL2Composite.GetOrderBook(instID)
		}
	}
	return orderBook
}

func updateOrderBook(instType config.InstrumentType, ch config.Channel, obMsg container.OrderBookMsg, globalContext *context.GlobalContext) bool {
	if instType == config.FuturesInstrument {
		if ch == config.BboTbtChannel {
			return globalContext.OkxFuturesBboComposite.UpdateOrderBook(obMsg)
		} else if ch == config.Books50L2TbtChannel {
			return globalContext.OkxFuturesBooks50L2Composite.UpdateOrderBook(obMsg)
		} else if ch == config.BooksL2TbtChannel {
			return globalContext.OkxFuturesL2Composite.UpdateOrderBook(obMsg)
		}
	} else {
		if ch == config.BboTbtChannel {
			return globalContext.OkxSpotBboComposite.UpdateOrderBook(obMsg)
		} else if ch == config.Books50L2TbtChannel {
			return globalContext.OkxSpotBooks50L2Composite.UpdateOrderBook(obMsg)
		} else if ch == config.BooksL2TbtChannel {
			return globalContext.OkxSpotL2Composite.UpdateOrderBook(obMsg)
		}
	}
	return true
}
