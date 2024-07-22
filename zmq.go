package main

import (
	"best-ticker/config"
	"best-ticker/container"
	"best-ticker/context"
	"best-ticker/protocol/pb"
	"best-ticker/utils/logger"
	zmq "github.com/pebbe/zmq4"
	"google.golang.org/protobuf/proto"
	"os"
)

func StartZmq() {
	logger.Info("Start Ticker ZMQ")
	startTickerZmq(&globalConfig, &globalContext)

	logger.Info("Start OrderBook ZMQ")
	startOrderBookZmq(&globalConfig, &globalContext)
}

func startTickerZmq(cfg *config.Config, globalContext *context.GlobalContext) {
	go func() {
		ctx, err := zmq.NewContext()
		if err != nil {
			logger.Fatal("[ZMQ] Failed to create context, error: %s", err.Error())
			os.Exit(1)
		}
		pub, err := ctx.NewSocket(zmq.PUB)
		if err != nil {
			ctx.Term()
			logger.Fatal("[ZMQ] Failed to create PUB socket, error: %s", err.Error())
			os.Exit(2)
		}

		err = pub.Bind(cfg.TickerZMQIPC)
		if err != nil {
			ctx.Term()
			logger.Fatal("[ZMQ] Failed to bind IPC %s, error: %s", cfg.TickerZMQIPC, err.Error())
			os.Exit(3)
		}

		defer pub.Close()
		defer ctx.Term()

		logger.Info("ticker zmq publisher started")
		for {
			select {
			case t := <-globalContext.TickerUpdateChan:
				md := &pb.OkxTicker{
					InstID:   t.InstID,
					InstType: string(t.InstType),
					BestBid:  t.BidPrice,
					BestAsk:  t.AskPrice,
					EventTs:  t.UpdateTimeMs,
					BidSz:    t.BidSize,
					AskSz:    t.AskSize,
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

func startOrderBookZmq(cfg *config.Config, globalContext *context.GlobalContext) {
	go func() {
		ctx, err := zmq.NewContext()
		if err != nil {
			logger.Fatal("[ZMQ] Failed to create context, error: %s", err.Error())
			os.Exit(1)
		}
		pub, err := ctx.NewSocket(zmq.PUB)
		if err != nil {
			ctx.Term()
			logger.Fatal("[ZMQ] Failed to create PUB socket, error: %s", err.Error())
			os.Exit(2)
		}

		err = pub.Bind(cfg.OrderBookZMQIPC)
		if err != nil {
			ctx.Term()
			logger.Fatal("[ZMQ] Failed to bind IPC %s, error: %s", cfg.OrderBookZMQIPC, err.Error())
			os.Exit(3)
		}

		defer pub.Close()
		defer ctx.Term()

		logger.Info("orderBook zmq publisher started")
		for {
			select {
			case ob := <-globalContext.OrderBookUpdateChan:
				var orderBook *pb.OkxOrderBook
				if ob.InstType == config.FuturesInstrument {
					bboOb := globalContext.OkxFuturesBboComposite.GetOrderBook(ob.InstID)
					l250Ob := globalContext.OkxFuturesBooks50L2Composite.GetOrderBook(ob.InstID)
					l2Ob := globalContext.OkxFuturesL2Composite.GetOrderBook(ob.InstID)
					orderBook = genOrderBooks(10, bboOb, l250Ob, l2Ob)
				} else if ob.InstType == config.SpotInstrument {
					bboOb := globalContext.OkxSpotBboComposite.GetOrderBook(ob.InstID)
					l250Ob := globalContext.OkxSpotBooks50L2Composite.GetOrderBook(ob.InstID)
					l2Ob := globalContext.OkxSpotL2Composite.GetOrderBook(ob.InstID)
					orderBook = genOrderBooks(10, bboOb, l250Ob, l2Ob)
				}
				if orderBook == nil {
					continue
				}
				orderBook.InstID = ob.InstID
				orderBook.InstType = string(ob.InstType)
				data, err := proto.Marshal(orderBook)
				if err != nil {
					logger.Error("[ZMQ] Error marshaling OrderBook: %v", err)
					continue
				}
				_, err = pub.Send(string(data), 0)
				if err != nil {
					logger.Error("[ZMQ] Error sending OrderBook: %v", err)
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
		} else {
			ob = mergeOrderBook(limit, bboOb, l2Ob)
			ob.UpdateTimeMs = bboOb.UpdateTime()
		}
	} else {
		// l250Ob的数据新
		if bboOb.UpdateTime() <= l250Ob.UpdateTime() {
			ob = l250Ob.LimitDepth(limit)
			ob.UpdateTimeMs = l250Ob.UpdateTime()
		} else {
			ob = mergeOrderBook(limit, bboOb, l250Ob)
			ob.UpdateTimeMs = bboOb.UpdateTime()
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
