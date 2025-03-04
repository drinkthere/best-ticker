package container

import (
	"best-ticker/config"
	"best-ticker/protocol/pb"
	"best-ticker/utils/logger"
	"fmt"
	"github.com/drinkthere/okx/models/market"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type OrderBookUpdate struct {
	InstID   string
	InstType config.InstrumentType
	Channel  config.Channel
}
type OrderBook struct {
	bids         []*market.OrderBookEntity
	asks         []*market.OrderBookEntity
	mu           *sync.RWMutex
	updateTimeMs int64
	seqID        int64
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		bids:         make([]*market.OrderBookEntity, 0),
		asks:         make([]*market.OrderBookEntity, 0),
		mu:           &sync.RWMutex{},
		updateTimeMs: 0,
		seqID:        0,
	}
}

func (ob *OrderBook) update(channel config.Channel, action string, bookChange *market.OrderBookWs) bool {
	result := true
	if action == "snapshot" {
		ob.reset(bookChange)
		return result
	} else if action == "update" {
		result = ob.handleOrderBookMessage(bookChange)
	} else if channel == config.BboTbtChannel {
		ob.reset(bookChange)
		return result
	}
	return result
}

func (ob *OrderBook) handleOrderBookMessage(bookChange *market.OrderBookWs) bool {
	// logger.Info("prevSeqId=%d, seqId=%d", bookChange.PrevSeqID, bookChange.SeqID)
	bids := parseOrders(bookChange.Bids, config.DescSortType)
	asks := parseOrders(bookChange.Asks, config.AscSortType)
	if len(bids) == 0 && len(asks) == 0 {
		// data not update, just update seqID
		ob.updateTimeMs = time.Time(bookChange.TS).UnixMilli()
		ob.seqID = bookChange.SeqID
		return true
	}

	ob.handleDeltas(bids, asks)
	checkSum := bookChange.Checksum
	if checkSum != 0 {
		maxItems := 25
		var payloadArray []string
		for i := 0; i < maxItems; i++ {
			if i < len(ob.bids) {
				payloadArray = append(payloadArray, strconv.FormatFloat(ob.bids[i].DepthPrice, 'f', -1, 64))
				payloadArray = append(payloadArray, strconv.FormatFloat(ob.bids[i].Size, 'f', -1, 64))
			}
			if i < len(ob.asks) {
				payloadArray = append(payloadArray, strconv.FormatFloat(ob.asks[i].DepthPrice, 'f', -1, 64))
				payloadArray = append(payloadArray, strconv.FormatFloat(ob.asks[i].Size, 'f', -1, 64))
			}
		}

		payload := strings.Join(payloadArray, ":")
		localChecksum := crc32.ChecksumIEEE([]byte(payload))

		if int32(localChecksum) != checkSum {
			logger.Warn("[Depth Update] Checksum does not match. checksum=%d", checkSum)
			return false
		}
	}

	ob.updateTimeMs = time.Time(bookChange.TS).UnixMilli()
	ob.seqID = bookChange.SeqID
	return true
}

func (ob *OrderBook) handleDeltas(bids []*market.OrderBookEntity, asks []*market.OrderBookEntity) {
	//for _, obid := range ob.bids {
	//	logger.Info("obid=%f, size=%f", obid.DepthPrice, obid.Size)
	//}
	//for _, oask := range ob.asks {
	//	logger.Info("oask=%f, size=%f", oask.DepthPrice, oask.Size)
	//}
	//for _, bid := range bids {
	//	logger.Info("bid=%f, size=%f", bid.DepthPrice, bid.Size)
	//}
	//for _, ask := range asks {
	//	logger.Info("ask=%f, size=%f", ask.DepthPrice, ask.Size)
	//}

	// 处理 bids 变更
	for _, newBid := range bids {
		found := false
		for i, bid := range ob.bids {
			if bid.DepthPrice == newBid.DepthPrice {
				// 找到了相同价格的 bid
				found = true
				if newBid.Size == 0 {
					// 删除此深度
					//logger.Info("Delete bid price=%f, size=%f", bid.DepthPrice, bid.Size)
					ob.bids = append(ob.bids[:i], ob.bids[i+1:]...)
				} else {
					// 更新 Size
					bid.Size = newBid.Size
				}
				break
			} else if bid.DepthPrice < newBid.DepthPrice {
				// 在正确的位置插入新的 bid
				ob.bids = append(ob.bids[:i], append([]*market.OrderBookEntity{newBid}, ob.bids[i:]...)...)
				found = true
				break
			}
		}
		if !found && newBid.Size != 0 {
			// 未找到相同价格的 bid,追加到末尾
			ob.bids = append(ob.bids, newBid)
		}
	}

	// 处理 asks 变更
	for _, newAsk := range asks {
		found := false
		for i, ask := range ob.asks {
			if ask.DepthPrice == newAsk.DepthPrice {
				// 找到了相同价格的 ask
				found = true
				if newAsk.Size == 0 {
					// 删除此深度
					//logger.Info("Delete ask price=%f, size=%f", ask.DepthPrice, ask.Size)
					ob.asks = append(ob.asks[:i], ob.asks[i+1:]...)
				} else {
					// 更新 Size
					ask.Size = newAsk.Size
				}
				break
			} else if ask.DepthPrice > newAsk.DepthPrice {
				// 在正确的位置插入新的 ask
				ob.asks = append(ob.asks[:i], append([]*market.OrderBookEntity{newAsk}, ob.asks[i:]...)...)
				found = true
				break
			}
		}
		if !found && newAsk.Size != 0 {
			// 未找到相同价格的 ask,追加到末尾
			ob.asks = append(ob.asks, newAsk)
		}
	}
}

func (ob *OrderBook) reset(bookChange *market.OrderBookWs) {
	ob.bids = parseOrders(bookChange.Bids, config.DescSortType)
	ob.asks = parseOrders(bookChange.Asks, config.AscSortType)
	ob.updateTimeMs = time.Time(bookChange.TS).UnixMilli()
	ob.seqID = bookChange.SeqID
}

func parseOrders(orders []*market.OrderBookEntity, sortType config.SortType) []*market.OrderBookEntity {
	if len(orders) == 0 {
		return orders
	}
	sort.Slice(orders, func(i, j int) bool {
		switch sortType {
		case config.AscSortType:
			return orders[i].DepthPrice < orders[j].DepthPrice
		case config.DescSortType:
			return orders[i].DepthPrice > orders[j].DepthPrice
		default:
			// 处理无效的 sortType
			return false
		}
	})
	return orders
}

func (ob *OrderBook) BestBid() market.OrderBookEntity {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	if len(ob.bids) == 0 {
		return market.OrderBookEntity{}
	}
	return *ob.bids[0]
}

func (ob *OrderBook) BestAsk() market.OrderBookEntity {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	if len(ob.asks) == 0 {
		return market.OrderBookEntity{}
	}
	return *ob.asks[0]
}

func (ob *OrderBook) UpdateTime() int64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.updateTimeMs
}

func (ob *OrderBook) SeqID() int64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.seqID
}

func (ob *OrderBook) AsksLength() int {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return len(ob.asks)
}

func (ob *OrderBook) BidsLength() int {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return len(ob.bids)
}

type PublicOrderBook struct {
	Bids         []*market.OrderBookEntity
	Asks         []*market.OrderBookEntity
	UpdateTimeMs int64
}

func (ob *OrderBook) LimitDepth(limit int) *pb.OkxOrderBook {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	newOb := &pb.OkxOrderBook{
		Bids:         make([]*pb.OkxOrder, 0, limit),
		Asks:         make([]*pb.OkxOrder, 0, limit),
		UpdateTimeMs: ob.updateTimeMs,
	}

	// 限制买卖盘的深度并深拷贝订单项
	if len(ob.bids) > limit {
		for _, entity := range ob.bids[:limit] {
			newEntity := pb.OkxOrder{
				Price: entity.DepthPrice,
				Size:  entity.Size,
			}
			newOb.Bids = append(newOb.Bids, &newEntity)
		}
	} else {
		for _, entity := range ob.bids {
			newEntity := pb.OkxOrder{
				Price: entity.DepthPrice,
				Size:  entity.Size,
			}
			newOb.Bids = append(newOb.Bids, &newEntity)
		}
	}

	if len(ob.asks) > limit {
		for _, entity := range ob.asks[:limit] {
			newEntity := pb.OkxOrder{
				Price: entity.DepthPrice,
				Size:  entity.Size,
			}
			newOb.Asks = append(newOb.Asks, &newEntity)
		}
	} else {
		for _, entity := range ob.asks {
			newEntity := pb.OkxOrder{
				Price: entity.DepthPrice,
				Size:  entity.Size,
			}
			newOb.Asks = append(newOb.Asks, &newEntity)
		}
	}

	return newOb
}

type OrderBookMsg struct {
	IP           string
	Colo         string
	Channel      config.Channel
	Exchange     config.Exchange
	InstID       string
	InstType     config.InstrumentType
	Action       string
	OrderBookMsg *market.OrderBookWs
}

type OrderBookComposite struct {
	Exchange      config.Exchange
	InstType      config.InstrumentType
	Channel       config.Channel
	OrderBooksMap map[string]OrderBook
	rwLock        *sync.RWMutex
}

func NewOrderBookComposite(exchange config.Exchange, instType config.InstrumentType, channel config.Channel) *OrderBookComposite {
	return &OrderBookComposite{
		Exchange:      exchange,
		InstType:      instType,
		Channel:       channel,
		OrderBooksMap: map[string]OrderBook{},
		rwLock:        &sync.RWMutex{},
	}
}

func (composite *OrderBookComposite) GetOrderBook(instID string) *OrderBook {
	composite.rwLock.RLock()
	ob, has := composite.OrderBooksMap[instID]
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
	orderBook, has := composite.OrderBooksMap[message.InstID]
	if !has {
		logger.Warn("orderBook is not exists. %s %t %s %s %s", message.IP, message.Colo, message.InstID, message.InstType, message.Channel)
		ob := NewOrderBook()
		updateResult = ob.update(composite.Channel, message.Action, message.OrderBookMsg)
		composite.OrderBooksMap[message.InstID] = *ob
	} else {
		updateResult = orderBook.update(composite.Channel, message.Action, message.OrderBookMsg)
		composite.OrderBooksMap[message.InstID] = orderBook
	}
	return updateResult
}

type OrderBookCompositeWrapper struct {
	Exchange              config.Exchange
	InstType              config.InstrumentType
	OrderBookCompositeMap map[string]*OrderBookComposite
	rwLock                *sync.RWMutex
}

func (ocw *OrderBookCompositeWrapper) Init(exchange config.Exchange, instType config.InstrumentType, sources []config.Source) {
	ocw.Exchange = exchange
	ocw.InstType = instType
	ocw.OrderBookCompositeMap = map[string]*OrderBookComposite{}
	ocw.rwLock = new(sync.RWMutex)

	ocw.rwLock.RLock()
	defer ocw.rwLock.RUnlock()
	for _, source := range sources {
		for _, channel := range source.Channels {
			key := genOrderBookCompositeKey(source.IP, source.Colo, channel)
			ocw.OrderBookCompositeMap[key] = NewOrderBookComposite(exchange, instType, channel)
		}
	}
}

func (ocw *OrderBookCompositeWrapper) GetOrderBookComposite(ip string, colo string, channel config.Channel) *OrderBookComposite {
	key := genOrderBookCompositeKey(ip, colo, channel)
	ocw.rwLock.RLock()
	obc, has := ocw.OrderBookCompositeMap[key]
	ocw.rwLock.RUnlock()

	if has {
		return obc
	} else {
		return nil
	}
}

func genOrderBookCompositeKey(ip string, colo string, channel config.Channel) string {
	return fmt.Sprintf("%s_%s_%s", ip, colo, channel)
}

type FastestChannelSource struct {
	IP       string
	Colo     string
	UpdateTs int64
}

type FastestChannelSourceWrapper struct {
	Exchange   config.Exchange
	InstType   config.InstrumentType
	FastestMap map[string]FastestChannelSource
	RwLock     *sync.RWMutex
}

func (f *FastestChannelSourceWrapper) Init(exchange config.Exchange, instType config.InstrumentType, instIDs []string) {
	f.Exchange = exchange
	f.InstType = instType
	f.FastestMap = map[string]FastestChannelSource{}
	f.RwLock = new(sync.RWMutex)

	f.RwLock.RLock()
	defer f.RwLock.RUnlock()
	channels := []config.Channel{config.BboTbtChannel, config.BooksL2TbtChannel, config.Books50L2TbtChannel}
	for _, channel := range channels {
		for _, instID := range instIDs {
			key := genFastestOrderBookKey(channel, instID)
			f.FastestMap[key] = FastestChannelSource{}
		}
	}
}

func (f *FastestChannelSourceWrapper) GetFastestOrderBookSource(channel config.Channel, instID string) FastestChannelSource {
	key := genFastestOrderBookKey(channel, instID)
	f.RwLock.RLock()
	obc, has := f.FastestMap[key]
	f.RwLock.RUnlock()

	if has {
		return obc
	} else {
		return FastestChannelSource{}
	}
}

func (f *FastestChannelSourceWrapper) GetFastestOrderBookSourceNoLock(channel config.Channel, instID string) FastestChannelSource {
	key := genFastestOrderBookKey(channel, instID)
	obc, has := f.FastestMap[key]

	if has {
		return obc
	} else {
		return FastestChannelSource{}
	}
}
func (f *FastestChannelSourceWrapper) UpdateFastestOrderBookSource(channel config.Channel, instID string, ip string, colo string, updateTs int64) {
	key := genFastestOrderBookKey(channel, instID)
	f.RwLock.Lock()
	defer f.RwLock.Unlock()
	f.FastestMap[key] = FastestChannelSource{
		IP:       ip,
		Colo:     colo,
		UpdateTs: updateTs,
	}
}

func (f *FastestChannelSourceWrapper) UpdateFastestOrderBookSourceNoLock(channel config.Channel, instID string, ip string, colo string, updateTs int64) {
	key := genFastestOrderBookKey(channel, instID)

	f.FastestMap[key] = FastestChannelSource{
		IP:       ip,
		Colo:     colo,
		UpdateTs: updateTs,
	}
}

func genFastestOrderBookKey(channel config.Channel, instID string) string {
	return fmt.Sprintf("%s_%s", channel, instID)
}
