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

	OkxFuturesBooks50L2Composite container.OrderBookComposite
	OkxSpotBooks50L2Composite    container.OrderBookComposite

	OkxFuturesL2Composite container.OrderBookComposite
	OkxSpotL2Composite    container.OrderBookComposite
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
	okxFuturesBboComposite := container.OrderBookComposite{}
	okxFuturesBboComposite.Init(config.OkxExchange, config.FuturesInstrument, config.BboTbtChannel)
	context.OkxFuturesBboComposite = okxFuturesBboComposite

	okxSpotBboComposite := container.OrderBookComposite{}
	okxSpotBboComposite.Init(config.OkxExchange, config.SpotInstrument, config.BboTbtChannel)
	context.OkxSpotBboComposite = okxSpotBboComposite

	okxFuturesBooks50L2Composite := container.OrderBookComposite{}
	okxFuturesBooks50L2Composite.Init(config.OkxExchange, config.FuturesInstrument, config.Books50L2TbtChannel)
	context.OkxFuturesBooks50L2Composite = okxFuturesBooks50L2Composite

	okxSpotBooks50L2Composite := container.OrderBookComposite{}
	okxSpotBooks50L2Composite.Init(config.OkxExchange, config.SpotInstrument, config.Books50L2TbtChannel)
	context.OkxSpotBooks50L2Composite = okxSpotBooks50L2Composite

	okxFuturesL2Composite := container.OrderBookComposite{}
	okxFuturesL2Composite.Init(config.OkxExchange, config.FuturesInstrument, config.BooksL2TbtChannel)
	context.OkxFuturesL2Composite = okxFuturesL2Composite

	okxSpotL2Composite := container.OrderBookComposite{}
	okxSpotL2Composite.Init(config.OkxExchange, config.SpotInstrument, config.BooksL2TbtChannel)
	context.OkxSpotL2Composite = okxSpotL2Composite
}
