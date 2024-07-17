package container

import (
	"best-ticker/utils"
)

type InstrumentComposite struct {
	InstIDs []string // 永续合约支持多个交易对，如：BTC-USDT-SWAP, ETH-USDT-SWAP

	SpotInstIDs []string // 现货，支持多个交易对，如：BTC-USDT, ETH-USDT
}

func (composite *InstrumentComposite) Init(instIDs []string) {
	composite.InstIDs = []string{}

	for _, instID := range instIDs {
		composite.InstIDs = append(composite.InstIDs, instID)
		composite.SpotInstIDs = append(composite.SpotInstIDs, utils.ConvertToOkxSpotInstID(instID))
	}
}
