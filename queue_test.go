package main

import (
	"fmt"
	"sync"
	"testing"

	"gitgud.io/softashell/comfy-translator/translator"
)

func TestQueue(t *testing.T) {
	q := NewQueue()

	req := translator.Request{
		Text: "test",
	}

	if len(q.items) != 0 {
		t.Error("Queue not empty?")
	}

	ch, wait := q.Join(req)
	if ch != nil {
		t.Error("Returned unexpected channel")
	}
	if wait == true {
		t.Error("We shouldn't wait here")
	}

	if len(q.items) != 1 {
		t.Error("Queue does not contain one item")
	}

	wg := sync.WaitGroup{}

	joinWait(t, q, req, &wg, "test")
	joinWait(t, q, req, &wg, "test")
	joinWait(t, q, req, &wg, "test")

	if q.items[0].count != 3 {
		t.Error("Does not have enough waiting jobs")
	}

	q.Push(req, "test")

	if len(q.items) != 0 {
		t.Error("Queue not empty")
	}

	wg.Wait()
}

func joinWait(t *testing.T, q *Queue, req translator.Request, wg *sync.WaitGroup, expecting string) {
	ch, wait := q.Join(req)
	if wait != true {
		t.Error("We should wait here")
	}
	if ch == nil {
		t.Error("Didn't return channel")
	}
	go func(chan string) {
		wg.Add(1)
		defer wg.Done()

		out := <-ch
		fmt.Println("got", out)
		if out != expecting {
			t.Error("Unexpected output for waiting function")
		}
	}(ch)
}
