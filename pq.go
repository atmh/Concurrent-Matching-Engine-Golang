package main

import (
	"container/heap"
)

type Item struct {
	value Order
	index int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	isSell := pq[i].value.orderType == inputSell

	if pq[i].value.price != pq[j].value.price {
		if isSell {
			return pq[i].value.price < pq[j].value.price
		}
		return pq[i].value.price > pq[j].value.price
	}

	if pq[i].value.timestamp != pq[j].value.timestamp {
		return pq[i].value.timestamp < pq[j].value.timestamp
	}
	return pq[i].value.id < pq[j].value.id
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) update(item *Item, value Order, priority int) {
	item.value = value
	heap.Fix(pq, item.index)
}

func (pq *PriorityQueue) peek() interface{} {
	cur := *pq
	n := len(cur)
	if n > 0 {
		return cur[0]
	}
	return nil
}

func (pq *PriorityQueue) delete(id uint32) bool {
	deleted := false
	for _, item := range *pq {
		if item.value.id == id {
			heap.Remove(pq, item.index)
			deleted = true
			break
		}
	}
	return deleted
}

func (pq *PriorityQueue) isEmpty() bool {
	cur := *pq
	return len(cur) == 0
}

func (pq *PriorityQueue) len() int {
	cur := *pq
	return len(cur)
}