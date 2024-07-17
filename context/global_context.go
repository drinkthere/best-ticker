package context

import (
	"best-ticker/config"
	"best-ticker/container"
)

type GlobalContext struct {
	InstrumentComposite       container.InstrumentComposite
	OkxFuturesTickerComposite container.TickerComposite
	OkxSpotTickerComposite    container.TickerComposite

	OkxFuturesBboComposite container.OrderBookComposite
	OkxSpotBboComposite    container.OrderBookComposite
}

func (context *GlobalContext) Init(globalConfig *config.Config) {
	// 初始化交易对数据
	context.initInstrumentComposite(globalConfig)

	// 初始化ticker数据
	context.initTickerComposite()

	// 初始化orderBook数据
	context.initOrderBookComposite()
}

func (context *GlobalContext) initInstrumentComposite(globalConfig *config.Config) {
	instrumentComposite := container.InstrumentComposite{}
	instrumentComposite.Init(globalConfig.InstIDs)
	context.InstrumentComposite = instrumentComposite
}

func (context *GlobalContext) initTickerComposite() {
	okxFuturesComposite := container.TickerComposite{}
	okxFuturesComposite.Init(config.OkxExchange, config.FuturesInstrument)
	context.OkxFuturesTickerComposite = okxFuturesComposite

	okxSpotComposite := container.TickerComposite{}
	okxSpotComposite.Init(config.OkxExchange, config.SpotInstrument)
	context.OkxSpotTickerComposite = okxSpotComposite
}

func (context *GlobalContext) initOrderBookComposite() {
	okxFuturesComposite := container.OrderBookComposite{}
	okxFuturesComposite.Init(config.OkxExchange, config.FuturesInstrument, config.BooksBboTbtChannel)
	context.OkxFuturesBboComposite = okxFuturesComposite

	okxSpotComposite := container.OrderBookComposite{}
	okxSpotComposite.Init(config.OkxExchange, config.SpotInstrument, config.BooksBboTbtChannel)
	context.OkxSpotBboComposite = okxSpotComposite
}
