package message

import (
	"best-ticker/client"
	"best-ticker/config"
	"best-ticker/context"
	"best-ticker/utils/logger"
	"github.com/drinkthere/okx/events"
	"github.com/drinkthere/okx/events/public"
	wsRequestPublic "github.com/drinkthere/okx/requests/ws/public"
	"strings"
	"time"
)

func StartOkxTickerWs(cfg *config.Config, globalContext *context.GlobalContext,
	okxFuturesTickerChan chan *public.Tickers, okxSpotTickerChan chan *public.Tickers) {
	for _, source := range cfg.Sources {
		// 循环不同的IP，监听对应的tickers
		startOkxFuturesTickers(&source.OkxConfig, globalContext, source.Colo, source.IP, okxFuturesTickerChan)
		logger.Info("[FTickerWebSocket] Start Listen Okx Futures Tickers, colo:%s, ip:%s", source.Colo, source.IP)
		startOkxSpotTickers(&source.OkxConfig, globalContext, source.Colo, source.IP, okxSpotTickerChan)
		logger.Info("[STickerWebSocket] Start Listen Okx Spot Tickers, colo:%s, ip:%s", source.Colo, source.IP)
		time.Sleep(1 * time.Second)
	}
}

func startOkxFuturesTickers(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	colo string, localIP string, tickerChan chan *public.Tickers) {

	go func() {
		defer func() {
			logger.Warn("[FTickerWebSocket] Okx Futures Ticker Listening Exited.")
		}()
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
				err := okxClient.Client.Ws.Public.Tickers(wsRequestPublic.Tickers{
					InstID: instID,
				}, tickerChan)

				if err != nil {
					logger.Error("[FTickerWebSocket] Fail To Listen Futures Ticker For %s, %s", instID, err.Error())
					logger.Warn("[FTickerWebSocket] Will Reconnect Futures-WebSocket After 5 Second")
					time.Sleep(time.Second * 5)
					goto ReConnect
				} else {
					logger.Info("[FTickerWebSocket] Futures Ticker WebSocket Has Established For %s", instID)
				}

			}

			for {
				select {
				case sub := <-subChan:
					channel, _ := sub.Arg.Get("channel")
					logger.Info("[FTickerWebSocket] Futures Subscribe \t%s", channel)
				case err := <-errChan:
					if strings.Contains(err.Msg, "i/o timeout") {
						logger.Warn("[FTickerWebSocket] Error occurred %s, Will reconnect after 1 second.", err.Msg)
						time.Sleep(time.Second * 1)
						goto ReConnect
					}
					logger.Error("[FTickerWebSocket] Futures Occur Some Error \t%+v", err)
					for _, datum := range err.Data {
						logger.Error("[FTickerWebSocket] Futures Error Data \t\t%+v", datum)
					}
				case s := <-successCh:
					logger.Info("[FTickerWebSocket] Futures Receive Success: %+v", s)
				case b := <-okxClient.Client.Ws.DoneChan:
					logger.Info("[FTickerWebSocket] Futures End\t%v", b)
					// 暂停一秒再跳出，避免异常时频繁发起重连
					logger.Warn("[FTickerWebSocket] Will Reconnect Futures-WebSocket After 1 Second")
					time.Sleep(time.Second * 1)
					goto ReConnect
				}
			}
		}
	}()
}

func startOkxSpotTickers(cfg *config.OkxConfig, globalContext *context.GlobalContext,
	colo string, localIP string, tickerChan chan *public.Tickers) {

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
			okxClient.Init(cfg, colo, localIP)
			okxClient.Client.Ws.SetChannels(errChan, subChan, uSubChan, loginCh, successCh)
			for _, instID := range globalContext.InstrumentComposite.SpotInstIDs {
				err := okxClient.Client.Ws.Public.Tickers(wsRequestPublic.Tickers{
					InstID: instID,
				}, tickerChan)

				if err != nil {
					logger.Error("[STickerWebSocket] Fail To Listen Spot Ticker for %s, %s", instID, err.Error())
					logger.Warn("[STickerWebSocket] Will Reconnect Spot-WebSocket After 5 Second")
					time.Sleep(time.Second * 5)
					goto ReConnect
				}
				logger.Info("[STickerWebSocket] Spot Ticker WebSocket Has Established For %s", instID)
			}

			for {
				select {
				case sub := <-subChan:
					channel, _ := sub.Arg.Get("channel")
					logger.Info("[STickerWebSocket] Spot Subscribe \t%s", channel)
				case err := <-errChan:
					if strings.Contains(err.Msg, "i/o timeout") {
						logger.Warn("[STickerWebSocket] Error occurred %s, Will reconnect after 1 second.", err.Msg)
						time.Sleep(time.Second * 1)
						goto ReConnect
					}
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
}
