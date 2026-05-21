package message

import (
	"best-ticker/config"
	"testing"

	"github.com/drinkthere/okx/models/market"
)

func TestRequiresOrderBookSeqCheck(t *testing.T) {
	tests := []struct {
		name    string
		channel config.Channel
		want    bool
	}{
		{name: "books", channel: config.BooksChannel, want: true},
		{name: "books-l2-tbt", channel: config.BooksL2TbtChannel, want: true},
		{name: "books50-l2-tbt", channel: config.Books50L2TbtChannel, want: true},
		{name: "bbo-tbt", channel: config.BboTbtChannel, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requiresOrderBookSeqCheck(tt.channel); got != tt.want {
				t.Fatalf("requiresOrderBookSeqCheck(%s) = %t, want %t", tt.channel, got, tt.want)
			}
		})
	}
}

func TestAdvanceOrderBookSeqID(t *testing.T) {
	tests := []struct {
		name      string
		channel   config.Channel
		currSeqID int64
		book      market.OrderBookWs
		wantSeqID int64
		wantOK    bool
	}{
		{
			name:      "first snapshot",
			channel:   config.BooksL2TbtChannel,
			currSeqID: orderBookSeqUninitialized,
			book:      market.OrderBookWs{PrevSeqID: orderBookSeqUninitialized, SeqID: 100},
			wantSeqID: 100,
			wantOK:    true,
		},
		{
			name:      "zero seq id is valid",
			channel:   config.Books50L2TbtChannel,
			currSeqID: orderBookSeqUninitialized,
			book:      market.OrderBookWs{PrevSeqID: orderBookSeqUninitialized, SeqID: 0},
			wantSeqID: 0,
			wantOK:    true,
		},
		{
			name:      "normal update",
			channel:   config.BooksChannel,
			currSeqID: 100,
			book:      market.OrderBookWs{PrevSeqID: 100, SeqID: 101},
			wantSeqID: 101,
			wantOK:    true,
		},
		{
			name:      "heartbeat keeps seq",
			channel:   config.BooksL2TbtChannel,
			currSeqID: 101,
			book:      market.OrderBookWs{PrevSeqID: 101, SeqID: 101},
			wantSeqID: 101,
			wantOK:    true,
		},
		{
			name:      "sequence reset is accepted when prev matches current",
			channel:   config.BooksL2TbtChannel,
			currSeqID: 101,
			book:      market.OrderBookWs{PrevSeqID: 101, SeqID: 1},
			wantSeqID: 1,
			wantOK:    true,
		},
		{
			name:      "sequence gap rejects update",
			channel:   config.BooksL2TbtChannel,
			currSeqID: 101,
			book:      market.OrderBookWs{PrevSeqID: 100, SeqID: 102},
			wantSeqID: 101,
			wantOK:    false,
		},
		{
			name:      "bbo skips prev sequence check",
			channel:   config.BboTbtChannel,
			currSeqID: 101,
			book:      market.OrderBookWs{PrevSeqID: 0, SeqID: 102},
			wantSeqID: 102,
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSeqID, gotOK := advanceOrderBookSeqID(tt.channel, tt.currSeqID, &tt.book)
			if gotSeqID != tt.wantSeqID || gotOK != tt.wantOK {
				t.Fatalf("advanceOrderBookSeqID() = (%d, %t), want (%d, %t)", gotSeqID, gotOK, tt.wantSeqID, tt.wantOK)
			}
		})
	}
}
