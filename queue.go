package main

import (
	"sync"

	"gitgud.io/softashell/comfy-translator/translator"
)

type Queue struct {
	items []queueObject

	lock *sync.Mutex
}

type queueObject struct {
	req translator.Request

	lock    *sync.Mutex
	count   int
	outChan chan string
}

func NewQueue() *Queue {
	q := new(Queue)

	q.lock = &sync.Mutex{}

	return q
}

// Join adds a new item to queue or returns true and a channel if you need to wait
func (q *Queue) Join(req translator.Request) (chan string, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if pos, found := q.findItem(req); found {
		q.items[pos].lock.Lock()
		q.items[pos].count++
		q.items[pos].lock.Unlock()

		return q.items[pos].outChan, true
	}

	i := queueObject{
		req:     req,
		outChan: make(chan string),
		lock:    &sync.Mutex{},
	}

	q.addItem(i)

	return nil, false
}

// Push sends output to all waiting threads and removes item from queue
func (q *Queue) Push(req translator.Request, response string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if pos, found := q.findItem(req); found {
		q.items[pos].lock.Lock()

		// Sends response to all listeners
		for i := 0; i < q.items[pos].count; i++ {
			q.items[pos].outChan <- response
		}

		// Close unused channel
		close(q.items[pos].outChan)

		q.items[pos].lock.Unlock()

		q.removeItem(pos)
	}
}

func (q *Queue) findItem(req translator.Request) (int, bool) {
	for i, t := range q.items {
		if t.req == req {
			return i, true
		}
	}
	return -1, false
}

func (q *Queue) addItem(t queueObject) {
	q.items = append(q.items, t)
}

func (q *Queue) removeItem(i int) {
	q.items = append(q.items[:i], q.items[i+1:]...)
}
