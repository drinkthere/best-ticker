package main

import (
	"best-ticker/config"
	"best-ticker/container"
	"best-ticker/context"
	"best-ticker/protocol/pb"
	"best-ticker/utils/logger"
	zmq "github.com/pebbe/zmq4"
	"google.golang.org/protobuf/proto"
)

func StartZmq() {
	if globalConfig.Service == config.OkxExchange {
		logger.Info("Start Okx Ticker ZMQ")
		startOkxTickerZmq(&globalConfig, &globalContext)

		logger.Info("Start Okx OrderBook ZMQ")
		startOkxOrderBookZmq(&globalConfig, &globalContext)
	} else if globalConfig.Service == config.BinanceExchange {
		logger.Info("Start Binance Ticker ZMQ")
		startBinanceTickerZmq(&globalConfig, &globalContext)
	}

}

func startOkxTickerZmq(cfg *config.Config, globalContext *context.GlobalContext) {
	go func() {
		ctx, err := zmq.NewContext()
		if err != nil {
			logger.Fatal("[OKXZMQ] Failed to create context, error: %s", err.Error())
		}
		pub, err := ctx.NewSocket(zmq.PUB)
		if err != nil {
			ctx.Term()
			logger.Fatal("[OKXZMQ] Failed to create PUB socket, error: %s", err.Error())
		}

		err = pub.Bind(cfg.OkxTickerZMQIPC)
		if err != nil {
			ctx.Term()
			logger.Fatal("[OKXZMQ] Failed to bind IPC %s, error: %s", cfg.OkxTickerZMQIPC, err.Error())
		}

		defer pub.Close()
		defer ctx.Term()

		logger.Info("Okx ticker zmq publisher started")
		for {
			select {
			case t := <-globalContext.TickerUpdateChan:
				md := &pb.OkxTicker{
					InstID:      t.InstID,
					InstType:    string(t.InstType),
					BestBid:     t.BidPrice,
					BestAsk:     t.AskPrice,
					EventTs:     t.UpdateTimeMs,
					BidSz:       t.BidSize,
					AskSz:       t.AskSize,
					IsFromTrade: t.IsFromTrade,
				}

				data, err := proto.Marshal(md)
				if err != nil {
					logger.Error("[OKXZMQ] Error marshaling Ticker: %v", err)
					continue
				}
				_, err = pub.Send(string(data), 0)
				if err != nil {
					logger.Error("[OKXZMQ] Error sending Ticker: %v", err)
					continue
				}
			}
		}
	}()
}

func startOkxOrderBookZmq(cfg *config.Config, globalContext *context.GlobalContext) {
	go func() {
		ctx, err := zmq.NewContext()
		if err != nil {
			logger.Fatal("[OKXZMQ] Failed to create context, error: %s", err.Error())
		}
		pub, err := ctx.NewSocket(zmq.PUB)
		if err != nil {
			ctx.Term()
			logger.Fatal("[OKXZMQ] Failed to create PUB socket, error: %s", err.Error())
		}

		err = pub.Bind(cfg.OkxOrderBookZMQIPC)
		if err != nil {
			ctx.Term()
			logger.Fatal("[OKXZMQ] Failed to bind IPC %s, error: %s", cfg.OkxOrderBookZMQIPC, err.Error())
		}

		defer pub.Close()
		defer ctx.Term()

		logger.Info("Okx orderBook zmq publisher started")
		for {
			select {
			case ob := <-globalContext.OrderBookUpdateChan:
				var orderBook *pb.OkxOrderBook
				if ob.InstType == config.FuturesInstrument {
					bboSource := globalContext.OkxFuturesFastestSourceWrapper.GetFastestOrderBookSource(config.BboTbtChannel, ob.InstID)
					bboComposite := globalContext.OkxFuturesOrderBookCompositeWrapper.GetOrderBookComposite(bboSource.IP, bboSource.Colo, config.BboTbtChannel)
					if bboComposite == nil {
						continue
					}
					bboOb := bboComposite.GetOrderBook(ob.InstID)

					l250Source := globalContext.OkxFuturesFastestSourceWrapper.GetFastestOrderBookSource(config.Books50L2TbtChannel, ob.InstID)
					l250Composite := globalContext.OkxFuturesOrderBookCompositeWrapper.GetOrderBookComposite(l250Source.IP, l250Source.Colo, config.Books50L2TbtChannel)
					if l250Composite == nil {
						continue
					}
					l250Ob := l250Composite.GetOrderBook(ob.InstID)

					l2Source := globalContext.OkxFuturesFastestSourceWrapper.GetFastestOrderBookSource(config.BooksL2TbtChannel, ob.InstID)
					l2Composite := globalContext.OkxFuturesOrderBookCompositeWrapper.GetOrderBookComposite(l2Source.IP, l2Source.Colo, config.BooksL2TbtChannel)
					if l2Composite == nil {
						continue
					}
					l2Ob := l2Composite.GetOrderBook(ob.InstID)

					orderBook = genOrderBooks(10, bboOb, l250Ob, l2Ob)
				} else if ob.InstType == config.SpotInstrument {
					bboSource := globalContext.OkxSpotFastestSourceWrapper.GetFastestOrderBookSource(config.BboTbtChannel, ob.InstID)
					bboComposite := globalContext.OkxSpotOrderBookCompositeWrapper.GetOrderBookComposite(bboSource.IP, bboSource.Colo, config.BboTbtChannel)
					if bboComposite == nil {
						continue
					}
					bboOb := bboComposite.GetOrderBook(ob.InstID)

					l250Source := globalContext.OkxSpotFastestSourceWrapper.GetFastestOrderBookSource(config.Books50L2TbtChannel, ob.InstID)
					l250Composite := globalContext.OkxSpotOrderBookCompositeWrapper.GetOrderBookComposite(l250Source.IP, l250Source.Colo, config.Books50L2TbtChannel)
					if l250Composite == nil {
						continue
					}
					l250Ob := l250Composite.GetOrderBook(ob.InstID)

					l2Source := globalContext.OkxSpotFastestSourceWrapper.GetFastestOrderBookSource(config.BooksL2TbtChannel, ob.InstID)
					l2Composite := globalContext.OkxSpotOrderBookCompositeWrapper.GetOrderBookComposite(l2Source.IP, l2Source.Colo, config.BooksL2TbtChannel)
					if l2Composite == nil {
						continue
					}
					l2Ob := l2Composite.GetOrderBook(ob.InstID)
					orderBook = genOrderBooks(10, bboOb, l250Ob, l2Ob)
				}
				if orderBook == nil {
					continue
				}
				orderBook.InstID = ob.InstID
				orderBook.InstType = string(ob.InstType)
				data, err := proto.Marshal(orderBook)
				if err != nil {
					logger.Error("[OKXZMQ] Error marshaling OrderBook: %v", err)
					continue
				}
				_, err = pub.Send(string(data), 0)
				if err != nil {
					logger.Error("[OKXZMQ] Error sending OrderBook: %v", err)
					continue
				}
			}
		}

	}()
}

func genOrderBooks(limit int, bboOb, l250Ob, l2Ob *container.OrderBook) *pb.OkxOrderBook {
	if bboOb == nil || l250Ob == nil || l2Ob == nil {
		return nil
	}
	var ob *pb.OkxOrderBook
	if l250Ob.UpdateTime() <= l2Ob.UpdateTime() {
		// l2Ob的数据新
		if bboOb.UpdateTime() <= l2Ob.UpdateTime() {
			ob = l2Ob.LimitDepth(limit)
			ob.UpdateTimeMs = l2Ob.UpdateTime()
			ob.SeqID = l2Ob.SeqID()
		} else {
			ob = mergeOrderBook(limit, bboOb, l2Ob)
			ob.UpdateTimeMs = bboOb.UpdateTime()
			ob.SeqID = bboOb.SeqID()
		}
	} else {
		// l250Ob的数据新
		if bboOb.UpdateTime() <= l250Ob.UpdateTime() {
			ob = l250Ob.LimitDepth(limit)
			ob.UpdateTimeMs = l250Ob.UpdateTime()
			ob.SeqID = l250Ob.SeqID()
		} else {
			ob = mergeOrderBook(limit, bboOb, l250Ob)
			ob.UpdateTimeMs = bboOb.UpdateTime()
			ob.SeqID = bboOb.SeqID()
		}
	}
	return ob
}

func mergeOrderBook(limit int, bboOb, ob *container.OrderBook) *pb.OkxOrderBook {
	bestBid := bboOb.BestBid()
	bestAsk := bboOb.BestAsk()

	limitedOb := ob.LimitDepth(limit)

	// 1. 删除orderbook中，asks中DepthPrice比bestAsk.DepthPrice小的项，并将bestAsk作为第一项插入asks中
	newAsks := make([]*pb.OkxOrder, 0, limit+1)
	newAsks = append(newAsks, &pb.OkxOrder{Price: bestAsk.DepthPrice, Size: bestAsk.Size})
	for _, ask := range limitedOb.Asks {
		if ask.Price > bestAsk.DepthPrice {
			newAsks = append(newAsks, ask)
		}
	}
	limitedOb.Asks = newAsks

	// 2. 删除orderbook中，bids中DepthPrice比bestBid.DepthPrice 大的项，并将bestBid作为第一项插入bids中
	newBids := make([]*pb.OkxOrder, 0, limit+1)
	newBids = append(newBids, &pb.OkxOrder{Price: bestBid.DepthPrice, Size: bestBid.Size})
	for _, bid := range limitedOb.Bids {
		if bid.Price < bestBid.DepthPrice {
			newBids = append(newBids, bid)
		}
	}
	limitedOb.Bids = newBids
	return limitedOb
}

func startBinanceTickerZmq(cfg *config.Config, globalContext *context.GlobalContext) {
	go func() {
		ctx, err := zmq.NewContext()
		if err != nil {
			logger.Fatal("[BNZMQ] Failed to create context, error: %s", err.Error())
		}
		pub, err := ctx.NewSocket(zmq.PUB)
		if err != nil {
			ctx.Term()
			logger.Fatal("[BNZMQ] Failed to create PUB socket, error: %s", err.Error())
		}

		err = pub.Bind(cfg.BinanceTickerZMQIPC)
		if err != nil {
			ctx.Term()
			logger.Fatal("[BNZMQ] Failed to bind IPC %s, error: %s", cfg.BinanceTickerZMQIPC, err.Error())
		}

		defer pub.Close()
		defer ctx.Term()

		logger.Info("Binance ticker zmq publisher started")
		for {
			select {
			case t := <-globalContext.TickerUpdateChan:
				md := &pb.BinanceTicker{
					InstID:   t.InstID,
					InstType: string(t.InstType),
					BestBid:  t.BidPrice,
					BestAsk:  t.AskPrice,
					EventTs:  t.UpdateTimeMs,
					BidSz:    t.BidSize,
					AskSz:    t.AskSize,
					UpdateID: t.UpdateID,
				}

				data, err := proto.Marshal(md)
				if err != nil {
					logger.Error("[ZMQ] Error marshaling Ticker: %v", err)
					continue
				}
				_, err = pub.Send(string(data), 0)
				if err != nil {
					logger.Error("[ZMQ] Error sending Ticker: %v", err)
					continue
				}
			}
		}
	}()
}
