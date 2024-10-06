package message

import (
	"best-ticker/client"
	"best-ticker/config"
	"best-ticker/context"
	"best-ticker/utils/logger"
	"github.com/drinkthere/okx"
	"github.com/drinkthere/okx/events"
	"github.com/drinkthere/okx/events/public"
	"github.com/drinkthere/okx/models/market"
	wsRequestPublic "github.com/drinkthere/okx/requests/ws/public"
	"math/rand"
	"time"
)

func StartOkxTradesWs(cfg *config.Config, globalContext *context.GlobalContext,
	okxFuturesTickerChan chan *public.Tickers, okxSpotTickerChan chan *public.Tickers) {
	for _, source := range cfg.Sources {
		// 循环不同的IP，监听对应的tickers
		startOkxFuturesTrades(&source.OkxConfig, globalContext, source.Colo, source.IP, okxFuturesTickerChan)
		logger.Info("[FTradesWebSocket] Start Listen Okx Futures Tickers, isColo:%t, ip:%s", source.Colo, source.IP)
		startOkxSpotTrades(&source.OkxConfig, globalContext, source.Colo, source.IP, okxSpotTickerChan)
		logger.Info("[STradesWebSocket] Start Listen Okx Spot Tickers, isColo:%t, ip:%s", source.Colo, source.IP)
		time.Sleep(1 * time.Second)
	}
}

func startOkxFuturesTrades(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	isColo bool, localIP string, tickerChan chan *public.Tickers) {

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
			okxClient.Init(cfg, isColo, localIP)

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
					logger.Error("[FTradesWebSocket] Futures Occur Some Error \t%+v", err)
					for _, datum := range err.Data {
						logger.Error("[FTradesWebSocket] Futures Error Data \t\t%+v", datum)
					}
				case s := <-successCh:
					logger.Info("[FTradesWebSocket] Futures Receive Success: %+v", s)
				case t := <-tradesChan:
					tickers := convertTradesToTickersMsg(globalContext, t, okx.FuturesInstrument)
					if r.Int31n(10000) < 5 {
						logger.Info("[FTradesWebSocket] ticker is %+v", tickers.Tickers[0])
					}
					tickerChan <- tickers
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

func startOkxSpotTrades(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	isColo bool, localIP string, tickerChan chan *public.Tickers) {

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
			okxClient.Init(cfg, isColo, localIP)
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
					logger.Error("[STradeWebSocket] Spot Occur Some Error \t%+v", err)
					for _, datum := range err.Data {
						logger.Error("[STradeWebSocket] Spot Error Data \t\t%+v", datum)
					}
				case s := <-successCh:
					logger.Info("[STradeWebSocket] Spot Receive Success: %+v", s)
				case t := <-tradesChan:
					tickers := convertTradesToTickersMsg(globalContext, t, okx.SpotInstrument)
					if r.Int31n(10000) < 5 {
						logger.Info("[STradeWebSocket] ticker is %+v", tickers.Tickers[0])
					}
					tickerChan <- tickers
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

func convertTradesToTickersMsg(globalContext *context.GlobalContext, tradesEvent *public.Trades, instType okx.InstrumentType) *public.Tickers {
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

	ticker := &market.Ticker{
		InstID:   instID,
		AskPx:    okx.JSONFloat64(askPx),
		AskSz:    okx.JSONFloat64(askSz),
		BidPx:    okx.JSONFloat64(bidPx),
		BidSz:    okx.JSONFloat64(bidSz),
		InstType: instType,
		TS:       latestTrade.TS,
	}

	// 只根据最新的trade来设置ticker，所以这里长度为1
	tickers := &public.Tickers{
		Arg:     tradesEvent.Arg,
		Tickers: make([]*market.Ticker, 0, 1),
	}
	tickers.Arg.Set("channel", "tickers")
	tickers.Tickers = append(tickers.Tickers, ticker)

	return tickers
}
