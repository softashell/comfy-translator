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
	waiters []waiterObject
}

type waiterObject struct {
	outChan chan string
}

func NewQueue() *Queue {
	q := new(Queue)

	q.lock = &sync.Mutex{}

	return q
}

// Join adds a new item to queue or returns true and a channel if you need to wait
func (q *Queue) Join(req translator.Request) (chan string, bool) {

	if pos, found := q.findItem(req); found {
		var w waiterObject

		w.outChan = make(chan string)

		q.items[pos].lock.Lock()
		q.items[pos].waiters = append(q.items[pos].waiters, w)
		q.items[pos].lock.Unlock()

		return w.outChan, true
	}

	i := queueObject{}
	i.req = req

	q.addItem(i)

	return nil, false
}

// Push sends output to all waiting threads and removes item from queue
func (q *Queue) Push(req translator.Request, response string) {
	if pos, found := q.findItem(req); found {
		q.items[pos].lock.Lock()

		for _, w := range q.items[pos].waiters {
			w.outChan <- response

			close(w.outChan)
		}

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
	q.lock.Lock()

	t.lock = &sync.Mutex{}

	q.items = append(q.items, t)

	q.lock.Unlock()
}

func (q *Queue) removeItem(i int) {
	q.lock.Lock()

	q.items = append(q.items[:i], q.items[i+1:]...)

	q.lock.Unlock()
}
