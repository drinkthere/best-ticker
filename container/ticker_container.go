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
	UpdateID     int64   // 更新ID 400900217
}

func (wrapper *TickerWrapper) updateTicker(message TickerWrapper) bool {
	if message.Exchange == config.BinanceExchange && message.InstType == config.SpotInstrument {
		// 币安现货没有消息更新时间（updateTimeMs），只有updateID
		if wrapper.UpdateID >= message.UpdateID {
			// message is expired
			return false
		}
	} else {
		if wrapper.UpdateTimeMs >= message.UpdateTimeMs {
			// message is expired
			return false
		}
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
	wrapper.UpdateID = message.UpdateID
	return true
}

type TickerComposite struct {
	Exchange       config.Exchange
	InstType       config.InstrumentType
	tickerWrappers map[string]TickerWrapper
	rwLock         *sync.RWMutex
}

func NewTickerComposite(exchange config.Exchange, instType config.InstrumentType) *TickerComposite {
	return &TickerComposite{
		Exchange:       exchange,
		InstType:       instType,
		tickerWrappers: map[string]TickerWrapper{},
		rwLock:         new(sync.RWMutex),
	}
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
