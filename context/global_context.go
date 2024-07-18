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

	OkxFuturesBooksComposite container.OrderBookComposite
	OkxSpotBooksComposite    container.OrderBookComposite

	OkxFuturesBooks5Composite container.OrderBookComposite
	OkxSpotBooks5Composite    container.OrderBookComposite
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

	okxFuturesBooksComposite := container.OrderBookComposite{}
	okxFuturesBooksComposite.Init(config.OkxExchange, config.FuturesInstrument, config.BooksChannel)
	context.OkxFuturesBooksComposite = okxFuturesBooksComposite

	okxSpotBooksComposite := container.OrderBookComposite{}
	okxSpotBooksComposite.Init(config.OkxExchange, config.SpotInstrument, config.BooksChannel)
	context.OkxSpotBooksComposite = okxSpotBooksComposite

	okxFuturesBooks5Composite := container.OrderBookComposite{}
	okxFuturesBooks5Composite.Init(config.OkxExchange, config.FuturesInstrument, config.Books5Channel)
	context.OkxFuturesBooks5Composite = okxFuturesBooks5Composite

	okxSpotBooks5Composite := container.OrderBookComposite{}
	okxSpotBooks5Composite.Init(config.OkxExchange, config.SpotInstrument, config.Books5Channel)
	context.OkxSpotBooks5Composite = okxSpotBooks5Composite
}
