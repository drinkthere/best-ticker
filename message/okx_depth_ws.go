package message

import (
	"best-ticker/client"
	"best-ticker/config"
	"best-ticker/context"
	"best-ticker/utils/logger"
	"github.com/drinkthere/okx/events"
	"github.com/drinkthere/okx/events/public"
	wsRequestPublic "github.com/drinkthere/okx/requests/ws/public"
	"time"
)

func StartOkxDepthWs(cfg *config.Config, globalContext *context.GlobalContext,
	okxFuturesTickerChan chan *public.OrderBook, okxSpotTickerChan chan *public.OrderBook) {

	// 循环不同的IP，监听不同的depth channel
	startOkxFuturesDepths(&cfg.OkxConf, globalContext, false, "", okxFuturesTickerChan)
	logger.Info("[FDepthWebSocket] Start Listen Okx Futures Depth Channel")
	startOkxSpotDepths(&cfg.OkxConf, globalContext, false, "", okxSpotTickerChan)
	logger.Info("[SDepthWebSocket] Start Listen Okx Spot Depth Channel")
}

func startOkxFuturesDepths(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	isColo bool, localIP string, depthChan chan *public.OrderBook) {

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

			okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
			for _, instID := range globalContext.InstrumentComposite.InstIDs {
				err := okxClient.Client.Ws.Public.OrderBook(wsRequestPublic.OrderBook{
					InstID:  instID,
					Channel: string(config.BooksBboTbtChannel),
				}, depthChan)

				if err != nil {
					logger.Fatal("[FDepthWebSocket] Fail To Listen Futures Depth For %s, %s", instID, err.Error())
				} else {
					logger.Info("[FDepthWebSocket] Futures Depth WebSocket Has Established For %s", instID)
				}

			}

			for {
				select {
				case sub := <-subChan:
					channel, _ := sub.Arg.Get("channel")
					logger.Info("[FDepthWebSocket] Futures Subscribe \t%s", channel)
				case err := <-errChan:
					logger.Error("[FDepthWebSocket] Futures Occur Some Error \t%+v", err)
					for _, datum := range err.Data {
						logger.Error("[FDepthWebSocket] Futures Error Data \t\t%+v", datum)
					}
				case s := <-successCh:
					logger.Info("[FDepthWebSocket] Futures Receive Success: %+v", s)
				case b := <-okxClient.Client.Ws.DoneChan:
					logger.Info("[FDepthWebSocket] Futures End\t%v", b)
					// 暂停一秒再跳出，避免异常时频繁发起重连
					logger.Warn("[FDepthWebSocket] Will Reconnect Futures-WebSocket After 1 Second")
					time.Sleep(time.Second * 1)
					goto ReConnect
				}
			}
		}
	}()
}

func startOkxSpotDepths(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	isColo bool, localIP string, tickerChan chan *public.OrderBook) {
	/*
		go func() {
			defer func() {
				logger.Warn("[STickerWebSocket] Okx Spot Ticker Listening Exited.")
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
				okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
				for _, instID := range globalContext.InstrumentComposite.SpotInstIDs {
					err := okxClient.Client.Ws.Public.Tickers(wsRequestPublic.Tickers{
						InstID: instID,
					}, tickerChan)

					if err != nil {
						logger.Fatal("[STickerWebSocket] Fail To Listen Spot Ticker for %s, %s", instID, err.Error())
					}
					logger.Info("[STickerWebSocket] Spot Ticker WebSocket Has Established For %s", instID)
				}

				for {
					select {
					case sub := <-subChan:
						channel, _ := sub.Arg.Get("channel")
						logger.Info("[STickerWebSocket] Spot Subscribe \t%s", channel)
					case err := <-errChan:
						logger.Error("[STickerWebSocket] Spot Occur Some Error \t%+v", err)
						for _, datum := range err.Data {
							logger.Error("[STickerWebSocket] Spot Error Data \t\t%+v", datum)
						}
					case s := <-successCh:
						logger.Info("[STickerWebSocket] Spot Receive Success: %+v", s)
					case b := <-okxClient.Client.Ws.DoneChan:
						logger.Info("[STickerWebSocket] Spot End\t%v", b)
						// 暂停一秒再跳出，避免异常时频繁发起重连
						logger.Warn("[STickerWebSocket] Will Reconnect Spot-WebSocket After 1 Second")
						time.Sleep(time.Second * 1)
						goto ReConnect
					}
				}
			}
		}()

	*/
}
