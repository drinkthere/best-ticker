package message

import (
	"best-ticker/config"
	"best-ticker/context"
	"best-ticker/utils/logger"
	binanceSpot "github.com/dictxwang/go-binance"
	binanceFutures "github.com/dictxwang/go-binance/futures"
	"math/rand"
	"time"
)

func StartBinanceMarketWs(cfg *config.Config, globalContext *context.GlobalContext,
	binanceFuturesTickerChan chan *binanceFutures.WsBookTickerEvent, binanceSpotTickerChan chan *binanceSpot.WsBookTickerEvent) {
	binanceWs := BinanceMarketWebSocket{}
	binanceWs.Init(binanceSpotTickerChan, binanceFuturesTickerChan)

	for _, source := range cfg.Sources {
		// 循环不同的IP，监听对应的tickers
		binanceWs.StartBinanceMarketWs(globalContext, source.Colo, source.IP)
		logger.Info("[BinanceTickerWebSocket] Start Listen Binance Futures and Spot Tickers, isColo:%t, ip:%s", source.Colo, source.IP)
		time.Sleep(1 * time.Second)
	}

}

type BinanceMarketWebSocket struct {
	spotTickerChan    chan *binanceSpot.WsBookTickerEvent
	futuresTickerChan chan *binanceFutures.WsBookTickerEvent
}

func (ws *BinanceMarketWebSocket) Init(
	spotTickerChan chan *binanceSpot.WsBookTickerEvent,
	futuresTickerChan chan *binanceFutures.WsBookTickerEvent) {

	ws.spotTickerChan = spotTickerChan
	ws.futuresTickerChan = futuresTickerChan
}

func (ws *BinanceMarketWebSocket) StartBinanceMarketWs(globalContext *context.GlobalContext, colo bool, ip string) {

	batchSize := 30
	instIDs := globalContext.InstrumentComposite.InstIDs

	instIDsSize := len(instIDs)
	for i := 0; i < instIDsSize; i += batchSize {
		end := i + batchSize
		if end > instIDsSize {
			end = instIDsSize
		}
		batch := instIDs[i:end]

		// 初始化现货行情监控
		innerMargin := innerBinanceSpotWebSocket{
			instIDs:    batch,
			tickerChan: ws.spotTickerChan,
			isStopped:  true,
			randGen:    rand.New(rand.NewSource(2)),
		}
		innerMargin.subscribeBookTickers(ip)

		// // 初始化合约行情监控
		// innerFutures := innerBinanceFuturesWebSocket{
		// 	instIDs:    batch,
		// 	tickerChan: ws.futuresTickerChan,
		// 	isStopped:  true,
		// 	randGen:    rand.New(rand.NewSource(2)),
		// }
		// innerFutures.subscribeBookTickers(ip, colo)
	}
	logger.Info("[BSTickerWebSocket] Start Listen Binance Spot Tickers")
	logger.Info("[BFTickerWebSocket] Start Listen Binance Futures Tickers")
}

type innerBinanceSpotWebSocket struct {
	instIDs    []string
	tickerChan chan *binanceSpot.WsBookTickerEvent
	isStopped  bool
	stopChan   chan struct{}
	randGen    *rand.Rand
}

func (ws *innerBinanceSpotWebSocket) handleTickerEvent(event *binanceSpot.WsBookTickerEvent) {
	if ws.randGen.Int31n(10000) < 2 {
		logger.Info("[BSTickerWebSocket] Binance Spot Event: %+v", event)
	}
	ws.tickerChan <- event
}

func (ws *innerBinanceSpotWebSocket) handleError(err error) {
	// 出错断开连接，再重连
	logger.Error("[BSTickerWebSocket] Binance Spot Handle Error And Reconnect Ws: %s", err.Error())
	ws.stopChan <- struct{}{}
	ws.isStopped = true
}

func (ws *innerBinanceSpotWebSocket) subscribeBookTickers(ip string) {

	go func() {
		defer func() {
			logger.Warn("[BSTickerWebSocket] Binance Spot BookTicker Listening Exited.")
		}()
		for {
			if !ws.isStopped {
				time.Sleep(time.Second * 1)
				continue
			}
			var stopChan chan struct{}
			var err error
			if ip == "" {
				_, stopChan, err = binanceSpot.WsCombinedBookTickerServe(ws.instIDs, ws.handleTickerEvent, ws.handleError)
			} else {
				_, stopChan, err = binanceSpot.WsCombinedBookTickerServeWithIP(ip, ws.instIDs, ws.handleTickerEvent, ws.handleError)
			}
			if err != nil {
				logger.Error("[BSTickerWebSocket] Subscribe Binance Margin Book Tickers Error: %s", err.Error())
				time.Sleep(time.Second * 1)
				continue
			}
			logger.Info("[BSTickerWebSocket] Subscribe Binance Margin Book Tickers: %d", len(ws.instIDs))
			// 重置channel和时间
			ws.stopChan = stopChan
			ws.isStopped = false
		}
	}()
}

type innerBinanceFuturesWebSocket struct {
	instIDs    []string
	tickerChan chan *binanceFutures.WsBookTickerEvent
	isStopped  bool
	stopChan   chan struct{}
	randGen    *rand.Rand
}

func (ws *innerBinanceFuturesWebSocket) handleTickerEvent(event *binanceFutures.WsBookTickerEvent) {
	if ws.randGen.Int31n(10000) < 2 {
		logger.Info("[BFTickerWebSocket] Binance Futures Event: %+v", event)
	}
	ws.tickerChan <- event
}

func (ws *innerBinanceFuturesWebSocket) handleError(err error) {
	// 出错断开连接，再重连
	logger.Error("[BFTickerWebSocket] Binance Futures Handle Error And Reconnect Ws: %s", err.Error())
	ws.stopChan <- struct{}{}
	ws.isStopped = true
}

func (ws *innerBinanceFuturesWebSocket) subscribeBookTickers(ip string, colo bool) {

	go func() {
		defer func() {
			logger.Warn("[BFTickerWebSocket] Binance Futures Flash Ticker Listening Exited.")
		}()
		for {
			if !ws.isStopped {
				time.Sleep(time.Second * 1)
				continue
			}
			if colo {
				binanceFutures.UseIntranet = true
			} else {
				binanceFutures.UseIntranet = false
			}
			var stopChan chan struct{}
			var err error
			if ip == "" {
				_, stopChan, err = binanceFutures.WsCombinedBookTickerServe(ws.instIDs, ws.handleTickerEvent, ws.handleError)
			} else {
				_, stopChan, err = binanceFutures.WsCombinedBookTickerServeWithIP(ip, ws.instIDs, ws.handleTickerEvent, ws.handleError)
			}

			if err != nil {
				logger.Error("[BFTickerWebSocket] Subscribe Binance Futures Book Tickers Error: %s", err.Error())
				time.Sleep(time.Second * 1)
				continue
			}
			logger.Info("[BFTickerWebSocket] Subscribe Binance Futures Book Tickers: %d", len(ws.instIDs))
			// 重置channel和时间
			ws.stopChan = stopChan
			ws.isStopped = false
		}
	}()
}
