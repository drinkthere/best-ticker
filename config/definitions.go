package config

type (
	Exchange       string
	InstrumentType string
	Channel        string
	SortType       string
)

const (
	BinanceExchange = Exchange("Binance")
	OkxExchange     = Exchange("Okx")

	UnknownInstrument = InstrumentType("UNKNOWN")
	SpotInstrument    = InstrumentType("SPOT")
	FuturesInstrument = InstrumentType("FUTURES")

	NoChannel           = Channel("tickers")        // for statistic
	Books50L2TbtChannel = Channel("books50-l2-tbt") // tick-by-tick 50 depth levels, 10ms
	BooksL2TbtChannel   = Channel("books-l2-tbt")   // tick-by-tick 400 depth levels, 10ms,
	BboTbtChannel       = Channel("bbo-tbt")        // tick-by-tick 1 depth level, 10ms

	AscSortType  = SortType("asc")  // 价格从低到高排序
	DescSortType = SortType("desc") // 价格从高到低排序
)
