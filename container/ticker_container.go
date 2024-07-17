package container

import (
	"best-ticker/config"
	"sync"
)

type TickerWrapper struct {
	Exchange config.Exchange
	InstType config.InstrumentType
	InstID   string
	Channel  config.Channel // Okx 深度数据的channel

	BidPrice     float64 // 买1价
	BidSize      float64 // 买1量
	AskPrice     float64 // 卖1价
	AskSize      float64 // 卖1量
	UpdateTimeMs int64   //更新时间（微秒）
}

func (wrapper *TickerWrapper) updateTicker(message TickerWrapper) bool {
	if wrapper.UpdateTimeMs >= message.UpdateTimeMs {
		// message is expired
		return false
	}
	if message.AskPrice > 0.0 && message.AskSize > 0.0 {
		wrapper.AskPrice = message.AskPrice
		wrapper.AskSize = message.AskSize
	}

	if message.BidPrice > 0.0 && message.BidSize > 0.0 {
		wrapper.BidPrice = message.BidPrice
		wrapper.BidSize = message.BidSize
	}

	wrapper.UpdateTimeMs = message.UpdateTimeMs
	return true
}

type TickerComposite struct {
	Exchange       config.Exchange
	InstType       config.InstrumentType
	tickerWrappers map[string]TickerWrapper
	rwLock         *sync.RWMutex
}

func (composite *TickerComposite) Init(exchange config.Exchange, instType config.InstrumentType) {
	composite.Exchange = exchange
	composite.InstType = instType
	composite.tickerWrappers = map[string]TickerWrapper{}
	composite.rwLock = new(sync.RWMutex)
}

func (composite *TickerComposite) GetTicker(instID string) *TickerWrapper {
	composite.rwLock.RLock()
	wrapper, has := composite.tickerWrappers[instID]
	composite.rwLock.RUnlock()

	if has {
		return &wrapper
	} else {
		return nil
	}
}

func (composite *TickerComposite) UpdateTicker(message TickerWrapper) bool {
	composite.rwLock.Lock()
	defer composite.rwLock.Unlock()

	updateResult := false
	if composite.Exchange != message.Exchange || composite.InstType != message.InstType {
		return updateResult
	}
	wrapper, has := composite.tickerWrappers[message.InstID]

	if !has {
		updateResult = wrapper.updateTicker(message)
		composite.tickerWrappers[message.InstID] = wrapper
	} else {
		updateResult = wrapper.updateTicker(message)
		composite.tickerWrappers[message.InstID] = wrapper
	}
	return updateResult
}
