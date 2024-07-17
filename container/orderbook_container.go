package container

import (
	"best-ticker/config"
	"container/heap"
	"github.com/drinkthere/okx/models/market"
	"sync"
	"time"
)

type PriceLevelHeap []market.OrderBookEntity

func (h PriceLevelHeap) Len() int           { return len(h) }
func (h PriceLevelHeap) Less(i, j int) bool { return h[i].DepthPrice < h[j].DepthPrice }
func (h PriceLevelHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PriceLevelHeap) Push(x interface{}) {
	*h = append(*h, x.(market.OrderBookEntity))
}

func (h *PriceLevelHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type OrderBook struct {
	bids                    PriceLevelHeap
	asks                    PriceLevelHeap
	removeCrossedLevels     bool
	receivedInitialSnapshot bool
	mu                      *sync.RWMutex
	updateTimeMs            int64
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		bids:                    make(PriceLevelHeap, 0),
		asks:                    make(PriceLevelHeap, 0),
		removeCrossedLevels:     true,
		receivedInitialSnapshot: false,
		mu:                      &sync.RWMutex{},
		updateTimeMs:            0,
	}
}

func (ob *OrderBook) Update(isSnapshot bool, bookChange *market.OrderBookWs) bool {
	if isSnapshot {
		ob.bids = ob.bids[:0]
		ob.asks = ob.asks[:0]
		ob.receivedInitialSnapshot = true
	}

	if ob.receivedInitialSnapshot {
		applyPriceLevelChanges(&ob.asks, bookChange.Asks)
		applyPriceLevelChanges(&ob.bids, bookChange.Bids)
	}

	ob.updateTimeMs = time.Time(bookChange.TS).UnixMilli()
	return true
}

func (ob *OrderBook) BestBid() *market.OrderBookEntity {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	if len(ob.bids) == 0 {
		return nil
	}
	return &ob.bids[0]
}

func (ob *OrderBook) BestAsk() *market.OrderBookEntity {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	if len(ob.asks) == 0 {
		return nil
	}
	return &ob.asks[0]
}

func (ob *OrderBook) UpdateTime() int64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.updateTimeMs
}

func applyPriceLevelChanges(tree *PriceLevelHeap, priceLevelChanges []*market.OrderBookEntity) {
	for _, priceLevel := range priceLevelChanges {
		index, found := findPriceLevel(*tree, priceLevel.DepthPrice)
		if found && priceLevel.OrderNumbers == 0 {
			*tree = removeFromHeap(*tree, index)
		} else if found {
			(*tree)[index].OrderNumbers = priceLevel.OrderNumbers
			heap.Fix(tree, index)
		} else if priceLevel.OrderNumbers != 0 {
			*tree = append(*tree, *priceLevel)
			heap.Init(tree)
		}
	}
}

func findPriceLevel(heap PriceLevelHeap, price float64) (int, bool) {
	for i, level := range heap {
		if level.DepthPrice == price {
			return i, true
		}
	}
	return -1, false
}

func removeFromHeap(heap PriceLevelHeap, index int) PriceLevelHeap {
	heap[index] = heap[len(heap)-1]
	return heap[:len(heap)-1]
}

type OrderBookMsg struct {
	Exchange     config.Exchange
	InstType     config.InstrumentType
	Channel      config.Channel
	InstID       string
	IsSnapshot   bool
	OrderBookMsg *market.OrderBookWs
}

type OrderBookComposite struct {
	Exchange          config.Exchange
	InstType          config.InstrumentType
	Channel           config.Channel
	orderBookWrappers map[string]OrderBook
	rwLock            *sync.RWMutex
}

func (composite *OrderBookComposite) Init(exchange config.Exchange, instType config.InstrumentType, channel config.Channel) {
	composite.Exchange = exchange
	composite.InstType = instType
	composite.Channel = channel
	composite.orderBookWrappers = map[string]OrderBook{}
	composite.rwLock = new(sync.RWMutex)
}

func (composite *OrderBookComposite) GetOrderBook(instID string) *OrderBook {
	composite.rwLock.RLock()
	ob, has := composite.orderBookWrappers[instID]
	composite.rwLock.RUnlock()

	if has {
		return &ob
	} else {
		return nil
	}
}

func (composite *OrderBookComposite) UpdateOrderBook(message OrderBookMsg) bool {
	composite.rwLock.Lock()
	defer composite.rwLock.Unlock()

	updateResult := false
	if composite.Exchange != message.Exchange || composite.InstType != message.InstType || composite.Channel != message.Channel {
		return updateResult
	}

	orderBook, has := composite.orderBookWrappers[message.InstID]

	if !has {
		ob := NewOrderBook()
		updateResult = ob.Update(message.IsSnapshot, message.OrderBookMsg)
		composite.orderBookWrappers[message.InstID] = *ob
	} else {
		updateResult = orderBook.Update(message.IsSnapshot, message.OrderBookMsg)
		composite.orderBookWrappers[message.InstID] = orderBook
	}
	return updateResult
}
