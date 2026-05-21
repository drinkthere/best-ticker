package message

import (
	"best-ticker/config"

	"github.com/drinkthere/okx/models/market"
)

const (
	orderBookSeqUninitialized = int64(-1)
	orderBookSeqResubscribing = int64(-2)
)

func requiresOrderBookSeqCheck(channel config.Channel) bool {
	switch channel {
	case config.BooksChannel, config.BooksL2TbtChannel, config.Books50L2TbtChannel:
		return true
	default:
		return false
	}
}

func advanceOrderBookSeqID(channel config.Channel, currSeqID int64, book *market.OrderBookWs) (int64, bool) {
	if !requiresOrderBookSeqCheck(channel) {
		return book.SeqID, true
	}
	if currSeqID != book.PrevSeqID {
		return currSeqID, false
	}
	return book.SeqID, true
}
