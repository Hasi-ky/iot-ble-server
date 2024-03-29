package storage

import (
	"sync"
	"time"
)

//缓存点信息
//针对终端
type NodeCache struct {
	FirstMark   string       `json:"FirstMark"`  ///conbine with devmac + frame
	TimeStamp time.Time `json:"timeStamp"`
	MsgType   string    `json:"msgType"`
}
// CQueue is a concurrent unbounded queue which uses two-Lock concurrent queue
type CQueue struct {
	Head  *cnode
	Tail  *cnode
	Hlock sync.Mutex
	Tlock sync.Mutex
	sz    int
}

type cnode struct {
	Value NodeCache
	Next  *cnode
}

// NewCQueue returns an empty CQueue.
func NewCQueue() *CQueue {
	n := &cnode{}
	return &CQueue{Head: n, Tail: n}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *CQueue) Enqueue(v NodeCache) {
	n := &cnode{Value: v}
	q.Tlock.Lock()
	q.Tail.Next = n // Link node at the end of the linked list
	q.Tail = n      // Swing Tail to node
	q.sz++
	q.Tlock.Unlock()
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *CQueue) Dequeue() NodeCache {
	var t NodeCache
	q.Hlock.Lock()
	n := q.Head
	newHead := n.Next
	if newHead == nil {
		q.Hlock.Unlock()
		return t
	}
	v := newHead.Value
	newHead.Value = t
	q.Head = newHead
	q.sz--
	q.Hlock.Unlock()
	return v
}

//Get header of queue with no sync
func (q *CQueue) Peek() NodeCache {
	var t NodeCache
	newHead := q.Head.Next
	if newHead == nil {
		return t
	}
	return newHead.Value
}

func (q *CQueue) Len() int {
	return q.sz
}
