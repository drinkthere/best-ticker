package message

import (
	"best-ticker/client"
	"best-ticker/config"
	"best-ticker/container"
	"best-ticker/context"
	"best-ticker/utils"
	"best-ticker/utils/logger"
	"github.com/drinkthere/okx"
	"github.com/drinkthere/okx/events"
	"github.com/drinkthere/okx/events/public"
	"github.com/drinkthere/okx/models/market"
	wsRequestPublic "github.com/drinkthere/okx/requests/ws/public"
	"math/rand"
	"strings"
	"time"
)

func StartOkxTradesWs(cfg *config.Config, globalContext *context.GlobalContext) {
	for _, source := range cfg.Sources {
		// 循环不同的IP，监听对应的tickers
		startOkxFuturesTrades(&source.OkxConfig, globalContext, source.Colo, source.IP)
		logger.Info("[FTradesWebSocket] Start Listen Okx Futures Tickers, colo:%s, ip:%s", source.Colo, source.IP)
		startOkxSpotTrades(&source.OkxConfig, globalContext, source.Colo, source.IP)
		logger.Info("[STradesWebSocket] Start Listen Okx Spot Tickers, colo:%s, ip:%s", source.Colo, source.IP)
		time.Sleep(1 * time.Second)
	}
}

func startOkxFuturesTrades(cfg *config.OkxConfig, globalContext *context.GlobalContext, colo string, localIP string) {

	go func() {
		defer func() {
			logger.Warn("[FTradesWebSocket] Okx Futures Ticker Listening Exited.")
		}()
		tradesChan := make(chan *public.Trades)
		r := rand.New(rand.NewSource(2))

		for {
		ReConnect:
			errChan := make(chan *events.Error)
			subChan := make(chan *events.Subscribe)
			uSubChan := make(chan *events.Unsubscribe)
			loginCh := make(chan *events.Login)
			successCh := make(chan *events.Success)

			var okxClient = client.OkxClient{}
			okxClient.Init(cfg, colo, localIP)

			okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
			for _, instID := range globalContext.InstrumentComposite.InstIDs {
				err := okxClient.Client.Ws.Public.Trades(wsRequestPublic.Trades{
					InstID: instID,
				}, tradesChan)

				if err != nil {
					logger.Error("[FTradesWebSocket] Fail To Listen Futures Trades For %s, %s", instID, err.Error())
					logger.Warn("[FTradesWebSocket] Will Reconnect Futures-Trades-WebSocket After 5 Second")
					time.Sleep(time.Second * 5)
					goto ReConnect
				} else {
					logger.Info("[FTradesWebSocket] Futures Trades WebSocket Has Established For %s", instID)
				}

			}

			for {
				select {
				case sub := <-subChan:
					channel, _ := sub.Arg.Get("channel")
					logger.Info("[FTradesWebSocket] Futures Subscribe \t%s", channel)
				case err := <-errChan:
					if strings.Contains(err.Msg, "i/o timeout") {
						logger.Warn("[FTradesWebSocket] Error occurred %s, Will reconnect after 1 second.", err.Msg)
						time.Sleep(time.Second * 1)
						goto ReConnect
					}
					logger.Error("[FTradesWebSocket] Futures Occur Some Error \t%+v", err)
					for _, datum := range err.Data {
						logger.Error("[FTradesWebSocket] Futures Error Data \t\t%+v", datum)
					}
				case s := <-successCh:
					logger.Info("[FTradesWebSocket] Futures Receive Success: %+v", s)
				case t := <-tradesChan:
					if r.Int31n(10000) < 5 {
						logger.Info("[FTradesWebSocket] trade is %+v", t.Trades[0])
					}
					tickerMsg := convertTradesToTickersMsg(globalContext, t, okx.SwapInstrument)
					result := globalContext.OkxFuturesTickerComposite.UpdateTicker(tickerMsg)
					if result {
						logger.Debug("swap result is %t", result)
						logger.Debug("FUTURES|CHANNEL|Trade")
						globalContext.TickerUpdateChan <- &tickerMsg
					}
				case b := <-okxClient.Client.Ws.DoneChan:
					logger.Info("[FTradesWebSocket] Futures End\t%v", b)
					// 暂停一秒再跳出，避免异常时频繁发起重连
					logger.Warn("[FTradesWebSocket] Will Reconnect Futures-Trades-WebSocket After 1 Second")
					time.Sleep(time.Second * 1)
					goto ReConnect
				}
			}
		}
	}()
}

func startOkxSpotTrades(cfg *config.OkxConfig, globalContext *context.GlobalContext, colo string, localIP string) {

	go func() {
		defer func() {
			logger.Warn("[STradeWebSocket] Okx Spot Ticker Listening Exited.")
		}()
		r := rand.New(rand.NewSource(2))
		tradesChan := make(chan *public.Trades)

		for {
		ReConnect:
			errChan := make(chan *events.Error)
			subChan := make(chan *events.Subscribe)
			uSubChan := make(chan *events.Unsubscribe)
			loginCh := make(chan *events.Login)
			successCh := make(chan *events.Success)

			var okxClient = client.OkxClient{}
			okxClient.Init(cfg, colo, localIP)
			okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
			for _, instID := range globalContext.InstrumentComposite.SpotInstIDs {
				err := okxClient.Client.Ws.Public.Trades(wsRequestPublic.Trades{
					InstID: instID,
				}, tradesChan)

				if err != nil {
					logger.Error("[STradeWebSocket] Fail To Listen Spot Trade for %s, %s", instID, err.Error())
					logger.Warn("[STradeWebSocket] Will Reconnect Spot-Trade-WebSocket After 5 Second")
					time.Sleep(time.Second * 5)
					goto ReConnect
				}
				logger.Info("[STradeWebSocket] Spot Trade WebSocket Has Established For %s", instID)
			}

			for {
				select {
				case sub := <-subChan:
					channel, _ := sub.Arg.Get("channel")
					logger.Info("[STradeWebSocket] Spot Subscribe \t%s", channel)
				case err := <-errChan:
					if strings.Contains(err.Msg, "i/o timeout") {
						logger.Warn("[STradeWebSocket] Error occurred %s, Will reconnect after 1 second.", err.Msg)
						time.Sleep(time.Second * 1)
						goto ReConnect
					}
					logger.Error("[STradeWebSocket] Spot Occur Some Error \t%+v", err)
					for _, datum := range err.Data {
						logger.Error("[STradeWebSocket] Spot Error Data \t\t%+v", datum)
					}
				case s := <-successCh:
					logger.Info("[STradeWebSocket] Spot Receive Success: %+v", s)
				case t := <-tradesChan:
					if r.Int31n(10000) < 5 {
						logger.Info("[STradeWebSocket] trade is %+v", t.Trades[0])
					}
					tickerMsg := convertTradesToTickersMsg(globalContext, t, okx.SpotInstrument)
					result := globalContext.OkxSpotTickerComposite.UpdateTicker(tickerMsg)
					if result {
						logger.Debug("spot result is %t", result)
						logger.Debug("SPOT|CHANNEL|Trade")
						globalContext.TickerUpdateChan <- &tickerMsg
					}
				case b := <-okxClient.Client.Ws.DoneChan:
					logger.Info("[STradeWebSocket] Spot End\t%v", b)
					// 暂停一秒再跳出，避免异常时频繁发起重连
					logger.Warn("[STradeWebSocket] Will Reconnect Spot-Trade-WebSocket After 1 Second")
					time.Sleep(time.Second * 1)
					goto ReConnect
				}
			}
		}
	}()
}

func convertTradesToTickersMsg(globalContext *context.GlobalContext, tradesEvent *public.Trades, instType okx.InstrumentType) container.TickerWrapper {

	// 获取最新的trade
	var latestTrade *market.Trade
	var updateTs time.Time
	for _, trade := range tradesEvent.Trades {
		currUpdateTs := time.Time(trade.TS)
		if currUpdateTs.After(updateTs) {
			updateTs = currUpdateTs
			latestTrade = trade
		}
	}
	instID := latestTrade.InstID
	tickInfo := globalContext.InstrumentComposite.SwapTickMap[instID]
	if instType == okx.SpotInstrument {
		tickInfo = globalContext.InstrumentComposite.SpotTickMap[instID]
	}

	bidPx, bidSz, askPx, askSz := 0.0, 0.0, 0.0, 0.0
	px := float64(latestTrade.Px)
	sz := float64(latestTrade.Sz)
	if latestTrade.Side == okx.TradeBuySide {
		bidPx = px
		bidSz = sz
		askPx = px + tickInfo.TickSiz
		askSz = tickInfo.MinSz
	} else {
		bidPx = px - tickInfo.TickSiz
		bidSz = tickInfo.MinSz
		askPx = px
		askSz = sz
	}

	return container.TickerWrapper{
		Exchange:     config.OkxExchange,
		InstType:     utils.ConvertToStdInstType(config.OkxExchange, string(instType)),
		InstID:       instID,
		Channel:      config.NoChannel,
		AskPrice:     askPx,
		AskSize:      askSz,
		BidPrice:     bidPx,
		BidSize:      bidSz,
		UpdateTimeMs: updateTs.UnixMilli(),
		IsFromTrade:  true,
	}
}
