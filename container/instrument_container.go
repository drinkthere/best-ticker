package container

import (
	"best-ticker/client"
	"best-ticker/config"
	"best-ticker/utils"
	"best-ticker/utils/logger"
	"github.com/drinkthere/okx"
	requests "github.com/drinkthere/okx/requests/rest/public"
)

type TickInfo struct {
	TickSiz float64
	LotSz   float64
	MinSz   float64
}

type InstrumentComposite struct {
	InstIDs []string // 永续合约支持多个交易对，如：BTC-USDT-SWAP, ETH-USDT-SWAP

	SpotInstIDs []string // 现货，支持多个交易对，如：BTC-USDT, ETH-USDT

	SpotTickMap map[string]*TickInfo // okx spot tick info
	SwapTickMap map[string]*TickInfo // okx swap tick info
}

func (composite *InstrumentComposite) Init(instIDs []string, exchange config.Exchange) {
	composite.InstIDs = []string{}

	for _, instID := range instIDs {
		composite.InstIDs = append(composite.InstIDs, instID)
		composite.SpotInstIDs = append(composite.SpotInstIDs, utils.ConvertToOkxSpotInstID(instID))
	}

	if exchange == config.OkxExchange {
		var okxClient = client.OkxClient{}
		okxClient.Init(&config.OkxConfig{}, false, "")
		composite.SpotTickMap = getTickMap(&okxClient, composite.SpotInstIDs, okx.SpotInstrument)
		composite.SwapTickMap = getTickMap(&okxClient, composite.InstIDs, okx.SwapInstrument)
	} else {
		composite.SpotTickMap = make(map[string]*TickInfo)
		composite.SwapTickMap = make(map[string]*TickInfo)
	}
}

func getTickMap(okxClient *client.OkxClient, instIDs []string, instrumentType okx.InstrumentType) map[string]*TickInfo {
	tickMap := make(map[string]*TickInfo)

	resp, err := okxClient.Client.Rest.PublicData.GetInstruments(requests.GetInstruments{
		InstType: instrumentType,
	})

	if err != nil {
		logger.Fatal("[InitInstrumentComposite] Get Spot tickers Failed")
	}

	if resp.Code != 0 {
		logger.Fatal("[InitInstrumentComposite] Get Spot tickers Failed, Resp is %+v", resp)
	}

	for _, inst := range resp.Instruments {
		if utils.InArray(inst.InstID, instIDs) {
			tickInfo := &TickInfo{
				TickSiz: float64(inst.TickSz),
				LotSz:   float64(inst.LotSz),
				MinSz:   float64(inst.MinSz),
			}
			tickMap[inst.InstID] = tickInfo
		}
	}
	return tickMap
}
