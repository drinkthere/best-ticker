package context

import (
	"best-ticker/config"
	"best-ticker/container"
)

type GlobalContext struct {
	InstrumentComposite           container.InstrumentComposite
	OkxFuturesTickerComposite     *container.TickerComposite
	OkxSpotTickerComposite        *container.TickerComposite
	BinanceFuturesTickerComposite *container.TickerComposite
	BinanceSpotTickerComposite    *container.TickerComposite

	OkxFuturesOrderBookCompositeWrapper *container.OrderBookCompositeWrapper
	OkxSpotOrderBookCompositeWrapper    *container.OrderBookCompositeWrapper

	OkxFuturesFastestSourceWrapper *container.FastestChannelSourceWrapper
	OkxSpotFastestSourceWrapper    *container.FastestChannelSourceWrapper

	OrderBookUpdateChan chan *container.OrderBookUpdate
	TickerUpdateChan    chan *container.TickerWrapper
}

func (context *GlobalContext) Init(globalConfig *config.Config) {
	// 初始化交易对数据
	context.initInstrumentComposite(globalConfig)

	// 初始化ticker数据
	context.initTickerComposite()

	// 初始化orderBook数据
	context.initOrderBookComposite(globalConfig)

	context.OrderBookUpdateChan = make(chan *container.OrderBookUpdate)
	context.TickerUpdateChan = make(chan *container.TickerWrapper)
}

func (context *GlobalContext) initInstrumentComposite(globalConfig *config.Config) {
	instrumentComposite := container.InstrumentComposite{}
	instrumentComposite.Init(globalConfig, globalConfig.Service)
	context.InstrumentComposite = instrumentComposite
}

func (context *GlobalContext) initTickerComposite() {
	context.OkxFuturesTickerComposite = container.NewTickerComposite(config.OkxExchange, config.FuturesInstrument)
	context.OkxSpotTickerComposite = container.NewTickerComposite(config.OkxExchange, config.SpotInstrument)

	context.BinanceFuturesTickerComposite = container.NewTickerComposite(config.BinanceExchange, config.FuturesInstrument)
	context.BinanceSpotTickerComposite = container.NewTickerComposite(config.BinanceExchange, config.SpotInstrument)
}

func (context *GlobalContext) initOrderBookComposite(globalConfig *config.Config) {
	futuresWrapper := &container.OrderBookCompositeWrapper{}
	futuresWrapper.Init(config.OkxExchange, config.FuturesInstrument, globalConfig.Sources)
	context.OkxFuturesOrderBookCompositeWrapper = futuresWrapper

	spotWrapper := &container.OrderBookCompositeWrapper{}
	spotWrapper.Init(config.OkxExchange, config.SpotInstrument, globalConfig.Sources)
	context.OkxSpotOrderBookCompositeWrapper = spotWrapper

	futuresFastWrapper := &container.FastestChannelSourceWrapper{}
	futuresFastWrapper.Init(config.OkxExchange, config.FuturesInstrument, context.InstrumentComposite.InstIDs)
	context.OkxFuturesFastestSourceWrapper = futuresFastWrapper

	spotFastWrapper := &container.FastestChannelSourceWrapper{}
	spotFastWrapper.Init(config.OkxExchange, config.SpotInstrument, context.InstrumentComposite.SpotInstIDs)
	context.OkxSpotFastestSourceWrapper = spotFastWrapper
}
