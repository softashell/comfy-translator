package main

import (
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

	ch, wait = q.Join(req)
	if wait != true {
		t.Error("We should wait here")
	}
	if ch == nil {
		t.Error("Didn't return channel")
	}
	wg := sync.WaitGroup{}
	go func(chan string) {
		wg.Add(1)
		defer wg.Done()

		for out := range ch {
			if out != "test" {
				t.Error("Unexpected output for first waiting function")
			}
		}

	}(ch)

	ch2, wait2 := q.Join(req)
	if wait2 != true {
		t.Error("We should wait here")
	}
	if ch2 == nil {
		t.Error("Didn't return channel")
	}

	go func(chan string) {
		wg.Add(1)
		defer wg.Done()

		for out := range ch2 {
			if out != "test" {
				t.Error("Unexpected output for second waiting function")
			}
		}
	}(ch2)

	if len(q.items) != 1 {
		t.Error("Queue does not contain one item")
	}

	if len(q.items[0].waiters) != 2 {
		t.Error("Does not have enough waiting jobs")
	}

	q.Push(req, "test")

	if len(q.items) != 0 {
		t.Error("Queue not empty")
	}

	wg.Wait()
}
