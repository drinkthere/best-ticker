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
	for _, source := range cfg.Sources {
		for _, channel := range source.Channels {
			// 循环不同的IP，监听不同的depth channel
			startOkxFuturesDepths(&source.OkxConfig, globalContext, source.Colo, source.IP, channel)
			logger.Info("[FDepthWebSocket] Start Listen Okx Futures Depth Channel, isColo:%t, ip:%s, channel=%s", source.Colo, source.IP, channel)

			startOkxSpotDepths(&source.OkxConfig, globalContext, source.Colo, source.IP, channel)
			logger.Info("[SDepthWebSocket] Start Listen Okx Spot Depth Channel, isColo:%t, ip:%s, channel=%s", source.Colo, source.IP, channel)

			time.Sleep(1 * time.Second)
		}
	}
}

func startOkxFuturesDepths(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	isColo bool, localIP string, channel config.Channel) {

	r := rand.New(rand.NewSource(2))
	depthChan := make(chan *public.OrderBook)
	unsubscribing := false
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
			if channel == config.BooksL2TbtChannel || channel == config.Books50L2TbtChannel {
				err := okxClient.Client.Ws.Private.Login()
				if err != nil {
					logger.Fatal("[FDepthWebSocket] Fail To Login, Error: %s", err.Error())
				}
			}

			okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
			currCh := string(channel)
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
					instID, _ := sub.Arg.Get("instId")
					logger.Info("[FDepthWebSocket] Futures Subscribe %s %s %s", localIP, instID, currCh)
				case usub := <-uSubChan:
					instID, _ := usub.Arg.Get("instId")
					logger.Warn("[FDepthWebSocket] Futures Unsubscribe %s %s %s", localIP, instID, currCh)
					time.Sleep(time.Minute * 1)
					logger.Info("unsubscribing is %t", unsubscribing)
					unsubscribing = false
					goto ReSubscribe
				case err := <-errChan:
					logger.Error("[FDepthWebSocket] Futures Occur Some Error %s %s %+v", localIP, currCh, err)
				case s := <-successCh:
					logger.Info("[FDepthWebSocket] Futures Receive Success: %s %s %+v", localIP, currCh, s)
				case s := <-depthChan:
					for _, b := range s.Books {
						// update orderbook
						instIDRaw, _ := s.Arg.Get("instId")
						instID := instIDRaw.(string)
						action := s.Action

						instType := config.FuturesInstrument
						obMsg := convertToObMsg(localIP, isColo, channel, instID, instType, action, b)
						result := updateOrderBook(obMsg, globalContext)
						if !result {
							if unsubscribing {
								continue
							} else {
								unsubscribing = true
								okxClient.Client.Ws.Public.UOrderBook(wsRequestPublic.OrderBook{
									InstID:  instID,
									Channel: currCh,
								})
							}
						} else {
							if r.Int31n(10000) < 5 {
								orderBook := getOrderBook(localIP, isColo, channel, instID, instType, globalContext)
								logger.Info("[GatherFDepth] %s %s orderBooks.bids is %+v, orderBooks.asks is %+v", instType, channel, orderBook.BestBid(), orderBook.BestAsk())
							}
							checkToUpdateTicker(localIP, isColo, channel, instID, instType, globalContext)
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
	isColo bool, localIP string, channel config.Channel) {

	r := rand.New(rand.NewSource(2))
	depthChan := make(chan *public.OrderBook)
	unsubscribing := false
	go func() {
		defer func() {
			logger.Warn("[SDepthWebSocket] Okx Spot Depth Listening Exited.")
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
			if channel == config.BooksL2TbtChannel || channel == config.Books50L2TbtChannel {
				err := okxClient.Client.Ws.Private.Login()
				if err != nil {
					logger.Fatal("[SDepthWebSocket] Fail To Login, Error: %s", err.Error())
				}
			}
			okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
			currCh := string(channel)
		ReSubscribe:
			for _, instID := range globalContext.InstrumentComposite.SpotInstIDs {
				err := okxClient.Client.Ws.Public.OrderBook(wsRequestPublic.OrderBook{
					InstID:  instID,
					Channel: currCh,
				}, depthChan)

				if err != nil {
					logger.Fatal("[SDepthWebSocket] Fail To Listen Spot Depth For %s %s %s %s", localIP, instID, currCh, err.Error())
				} else {
					logger.Info("[SDepthWebSocket] Spot Depth WebSocket Has Established For %s %s %s", localIP, instID, currCh)
				}
			}

			for {
				select {
				case sub := <-subChan:
					instID, _ := sub.Arg.Get("instId")
					logger.Info("[SDepthWebSocket] Spot Subscribe %s %s %s", localIP, instID, currCh)
				case usub := <-uSubChan:
					instID, _ := usub.Arg.Get("instId")
					logger.Warn("[SDepthWebSocket] Spot Unsubscribe %s %s %s", localIP, instID, currCh)
					time.Sleep(time.Minute * 1)
					logger.Info("unsubscribing is %t", unsubscribing)
					unsubscribing = false
					goto ReSubscribe
				case err := <-errChan:
					logger.Error("[SDepthWebSocket] Spot Occur Some Error %s %s %+v", localIP, currCh, err)
				case s := <-successCh:
					logger.Info("[SDepthWebSocket] Spot Receive Success: %s %s %+v", localIP, currCh, s)
				case s := <-depthChan:
					for _, b := range s.Books {
						// update orderbook
						instIDRaw, _ := s.Arg.Get("instId")
						instID := instIDRaw.(string)
						action := s.Action

						instType := config.SpotInstrument
						obMsg := convertToObMsg(localIP, isColo, channel, instID, instType, action, b)
						result := updateOrderBook(obMsg, globalContext)
						if !result {
							if unsubscribing {
								continue
							} else {
								unsubscribing = true
								okxClient.Client.Ws.Public.UOrderBook(wsRequestPublic.OrderBook{
									InstID:  instID,
									Channel: currCh,
								})
							}
						} else {
							if r.Int31n(10000) < 5 {
								orderBook := getOrderBook(localIP, isColo, channel, instID, instType, globalContext)
								logger.Info("[GatherFDepth] %s %s orderBooks.bids is %+v, orderBooks.asks is %+v", instType, channel, orderBook.BestBid(), orderBook.BestAsk())
							}
							checkToUpdateTicker(localIP, isColo, channel, instID, instType, globalContext)
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

func checkToUpdateTicker(localIP string, isColo bool, channel config.Channel, instID string, instType config.InstrumentType, globalContext *context.GlobalContext) {
	// b.TS > ticker.TS && ticker changed, update ticker
	ticker := getTicker(instID, instType, globalContext)
	orderBook := getOrderBook(localIP, isColo, channel, instID, instType, globalContext)
	if orderBook == nil || ticker == nil {
		logger.Info("orderBook is nil or ticker is nil")
		return
	}

	obUpdateTime := orderBook.UpdateTime()
	if obUpdateTime > ticker.UpdateTimeMs {
		bestBid := orderBook.BestBid()
		bestAsk := orderBook.BestAsk()

		if bestBid.DepthPrice != ticker.BidPrice || bestAsk.DepthPrice != ticker.AskPrice {
			tickerMsg := convertDepthToOkxTickerMessage(instType, instID, channel, bestBid, bestAsk, obUpdateTime)
			result := updateTicker(instType, tickerMsg, globalContext)
			if result {
				logger.Debug("%s|CHANNEL|%s", instType, channel)
				globalContext.TickerUpdateChan <- &tickerMsg
			}
		}
	}
}

func convertToObMsg(ip string, colo bool, channel config.Channel, instID string, instType config.InstrumentType,
	action string, msg *market.OrderBookWs) container.OrderBookMsg {
	return container.OrderBookMsg{
		IP:           ip,
		Colo:         colo,
		Channel:      channel,
		Exchange:     config.OkxExchange,
		InstID:       instID,
		InstType:     instType,
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
	} else if instType == config.SpotInstrument {
		result = globalContext.OkxSpotTickerComposite.UpdateTicker(tickerMsg)
	}
	return result
}

func getOrderBook(localIP string, isColo bool, channel config.Channel, instID string, instType config.InstrumentType, globalContext *context.GlobalContext) *container.OrderBook {

	var orderBook *container.OrderBook
	if instType == config.FuturesInstrument {
		futuresOrderBookComposite := globalContext.OkxFuturesOrderBookCompositeWrapper.GetOrderBookComposite(localIP, isColo, channel)
		orderBook = futuresOrderBookComposite.GetOrderBook(instID)
	} else {
		spotOrderBookComposite := globalContext.OkxSpotOrderBookCompositeWrapper.GetOrderBookComposite(localIP, isColo, channel)
		orderBook = spotOrderBookComposite.GetOrderBook(instID)
	}
	return orderBook
}

func updateOrderBook(obMsg container.OrderBookMsg, globalContext *context.GlobalContext) bool {
	if obMsg.InstType == config.FuturesInstrument {
		futuresOrderBookComposite := globalContext.OkxFuturesOrderBookCompositeWrapper.GetOrderBookComposite(obMsg.IP, obMsg.Colo, obMsg.Channel)
		updateResult := futuresOrderBookComposite.UpdateOrderBook(obMsg)
		if updateResult {
			// 更新当前channel最快的信息
			globalContext.OkxFuturesFastestSourceWrapper.UpdateFastestOrderBookSource(obMsg.Channel, obMsg.InstID, obMsg.IP, obMsg.Colo)
			globalContext.OrderBookUpdateChan <- &container.OrderBookUpdate{
				Channel:  obMsg.Channel,
				InstID:   obMsg.InstID,
				InstType: obMsg.InstType,
			}
			return true
		}

	} else if obMsg.InstType == config.SpotInstrument {
		spotOrderBookComposite := globalContext.OkxSpotOrderBookCompositeWrapper.GetOrderBookComposite(obMsg.IP, obMsg.Colo, obMsg.Channel)
		updateResult := spotOrderBookComposite.UpdateOrderBook(obMsg)
		if updateResult {
			globalContext.OkxSpotFastestSourceWrapper.UpdateFastestOrderBookSource(obMsg.Channel, obMsg.InstID, obMsg.IP, obMsg.Colo)

			globalContext.OrderBookUpdateChan <- &container.OrderBookUpdate{
				Channel:  obMsg.Channel,
				InstID:   obMsg.InstID,
				InstType: obMsg.InstType,
			}
			return true
		}
	}
	return false
}
