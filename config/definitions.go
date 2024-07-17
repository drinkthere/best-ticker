package config

type (
	Exchange       string
	InstrumentType string
	Channel        string
)

const (
	BinanceExchange = Exchange("Binance")
	OkxExchange     = Exchange("Okx")

	UnknownInstrument = InstrumentType("UNKNOWN")
	SpotInstrument    = InstrumentType("SPOT")
	FuturesInstrument = InstrumentType("FUTURES")

	NoChannel           = Channel("tickers")        // for statistic
	BooksChannel        = Channel("books")          // 400 depth levels
	Books5Channel       = Channel("books5")         //  5 depth levels
	Books50L2TbtChannel = Channel("books50-l2-tbt") // tick-by-tick 50 depth levels
	BooksL2TbtChannel   = Channel("books-l2-tbt")   // tick-by-tick 400 depth levels
	BooksBboTbtChannel  = Channel("bbo-tbt")        // tick-by-tick 1 depth level
)
