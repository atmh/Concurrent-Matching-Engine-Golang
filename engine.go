package main

import "C"
import (
	"container/heap"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Engine struct {
	channelMap ChannelMap
}

type ChannelMap struct {
	chanMap map[string]*Pair
}

type Order struct {
	id          uint32
	price       uint32
	count       uint32
	instrument  string
	timestamp   int64
	orderType   inputType
	executionId uint32
	clientDone 	chan interface{}
}

type Pair struct {
	first  interface{}
	second interface{}
}

func (e *Engine) accept(ctx context.Context, conn net.Conn, masterChan chan *Pair) {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go handleConn(ctx, conn, masterChan)
}

func handleConn(ctx context.Context, conn net.Conn, masterChan chan *Pair) {
	idToOrderMap := make(map[uint32]*Order)

	for {
		clientDone := make(chan interface{})
		select {
		case <-ctx.Done():
			close(clientDone)
			return
		default:
			in, err := readInput(conn)
			if err != nil {
				if err != io.EOF {
					_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				}
				close(clientDone)
				return
			}
			clientReceiverChan := make(chan *Pair)

			if in.orderType == inputCancel {
				if order, ok := idToOrderMap[in.orderId]; ok {
					in.instrument = order.instrument
					orderType := order.orderType
					getRoutineAndSubmit(in, masterChan, clientReceiverChan, true, orderType, clientDone)
				} else {
					outputOrderDeleted(in, false, GetCurrentTimestamp())
				}
				continue
			}

			getRoutineAndSubmit(in, masterChan, clientReceiverChan, false, in.orderType, clientDone)
			o := Order{id: in.orderId, price: in.price, count: in.count, instrument: in.instrument, orderType: in.orderType, 
				clientDone: clientDone}
			idToOrderMap[in.orderId] = &o
			// read from chnl
		}
		<- clientDone
	}
}

func getRoutineAndSubmit(in input, masterChan chan *Pair, clientReceiverChan chan *Pair, isCancel bool, 
	orderType inputType, clientDone chan interface{}) {
	if isCancel {
		o := Order{id: in.orderId, orderType: in.orderType, instrument: in.instrument, clientDone: clientDone}
		pair := Pair{first: &o, second: clientReceiverChan}
		masterChan <- &pair
		masterReponse := <-clientReceiverChan
		if orderType == inputBuy {
			masterReponse.second.(chan *Order) <- &o //send order to sell go routine
		} else {
			masterReponse.first.(chan *Order) <- &o //send order to buy go routine
		}
		return
	}
	routineReceiver := getRoutineFromMaster(masterChan, in, clientReceiverChan, clientDone)
	submitOrderToRoutine(in, routineReceiver, clientDone)
}

func getRoutineFromMaster(masterChan chan *Pair, in input, clientReceiverChan chan *Pair, clientDone chan interface{}) chan *Order {
	o := Order{id: in.orderId, price: in.price, count: in.count,
		instrument: in.instrument, orderType: in.orderType, clientDone: clientDone}
	pair := Pair{first: &o, second: clientReceiverChan}
	masterChan <- &pair
	masterResponse := <-clientReceiverChan

	if in.orderType == inputBuy {
		return masterResponse.first.(chan *Order)
	}
	return masterResponse.second.(chan *Order)
}

func submitOrderToRoutine(in input, routineReceiver chan *Order, clientDone chan interface{}) {
	o := Order{id: in.orderId, price: in.price, count: in.count,
		instrument: in.instrument, orderType: in.orderType, clientDone: clientDone}
	routineReceiver <- &o
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}

// creates buy, sell and adder goroutines
func (book *ChannelMap) addInstrument(ctx context.Context, instru string) {
	sellRoutineReceiver := make(chan *Order)
	buyRoutineReceiver := make(chan *Order)

	commChan := make(chan *Order)

	routineReceiverPair := Pair{first: buyRoutineReceiver, second: sellRoutineReceiver}
	book.chanMap[instru] = &routineReceiverPair

	sellBook := make(PriorityQueue, 0)
	heap.Init(&sellBook)
	go initInstrumentRoutine(ctx, buyRoutineReceiver, &sellBook, commChan)

	buyBook := make(PriorityQueue, 0)
	heap.Init(&buyBook)
	go initInstrumentRoutine(ctx, sellRoutineReceiver, &buyBook, commChan)
}

func initInstrumentRoutine(ctx context.Context, routineReceiver chan *Order, routineBook *PriorityQueue, commChan chan *Order) {
	for {
		select {
		case <-ctx.Done():
			return
		case o := <-routineReceiver:
			if o.orderType != inputCancel {
				match(o, routineBook, commChan)
				continue
			}

			select {
			case orderToAdd := <- commChan:
				timestamp := GetCurrentTimestamp()
				orderToAdd.timestamp = timestamp
				heap.Push(routineBook, &Item{value: *orderToAdd})
				in := input{orderType: orderToAdd.orderType, orderId: orderToAdd.id,
					price: orderToAdd.price, count: orderToAdd.count, instrument: orderToAdd.instrument}
				outputOrderAdded(in, timestamp)
				close(orderToAdd.clientDone)
			default:
			}

			in := input{orderType: o.orderType, orderId: o.id}
			if routineBook.delete(o.id) {
				outputOrderDeleted(in, true, GetCurrentTimestamp())
			} else {
				outputOrderDeleted(in, false, GetCurrentTimestamp())
			}
			close(o.clientDone)
		case orderToAdd := <- commChan:
			timestamp := GetCurrentTimestamp()
			orderToAdd.timestamp = timestamp
			heap.Push(routineBook, &Item{value: *orderToAdd})
			in := input{orderType: orderToAdd.orderType, orderId: orderToAdd.id,
				price: orderToAdd.price, count: orderToAdd.count, instrument: orderToAdd.instrument}
			outputOrderAdded(in, timestamp)
			close(orderToAdd.clientDone)
		}
	}
}

func match(o *Order, book *PriorityQueue, commChan chan *Order) {
	// matching here. If matched successfully, delete from book and send id to delete to deleteOrderChan.
	// else add order into adderInputChannel
	for o.count != 0 {
		//CAS to add to book
		if book.isEmpty() {
			select {
			case commChan <- o:
				return
			case orderToAdd := <- commChan:
				timestamp := GetCurrentTimestamp()
				orderToAdd.timestamp = timestamp
				heap.Push(book, &Item{value: *orderToAdd})
				in := input{orderType: orderToAdd.orderType, orderId: orderToAdd.id,
					price: orderToAdd.price, count: orderToAdd.count, instrument: orderToAdd.instrument}
				outputOrderAdded(in, timestamp)
				close(orderToAdd.clientDone)
				match(o, book, commChan)
				return
			}

		}

		// must peak here and pop later
		restingOrder := book.peek().(*Item).value
		if (o.orderType == 'S' && restingOrder.price < o.price) || (o.orderType == 'B' && restingOrder.price > o.price) {
			//CAS to add to book
			select {
			case commChan <- o:
				return
			case orderToAdd := <- commChan:
				timestamp := GetCurrentTimestamp()
				orderToAdd.timestamp = timestamp
				heap.Push(book, &Item{value: *orderToAdd})
				in := input{orderType: orderToAdd.orderType, orderId: orderToAdd.id,
					price: orderToAdd.price, count: orderToAdd.count, instrument: orderToAdd.instrument}
				outputOrderAdded(in, timestamp)
				close(orderToAdd.clientDone)
				match(o, book, commChan)
				return
			}
		}

		var minCount uint32
		if (o.count < restingOrder.count) {
			minCount = o.count
		} else {
			minCount = restingOrder.count
		}
		
		restingOrder.executionId += 1

		outputOrderExecuted(restingOrder.id, o.id, restingOrder.executionId,
			restingOrder.price, minCount, GetCurrentTimestamp())

		heap.Pop(book)
		if o.count >= restingOrder.count {
			o.count -= restingOrder.count
		} else {
			restingOrder.count -= o.count
			heap.Push(book, &Item{value: restingOrder})
			o.count = 0
		}
	}

	close(o.clientDone)
}
