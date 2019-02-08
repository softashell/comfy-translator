package google

import (
	"math/rand"
	"runtime"
	"time"

	"gitgud.io/softashell/comfy-translator/translator"
	"github.com/hashicorp/go-retryablehttp"

	log "github.com/sirupsen/logrus"
)

type inputChannel chan inputObject
type outputChannel chan returnObject

type BatchTranslator struct {
	inCh inputChannel

	// Delay between requests
	batchDelay time.Duration

	// Last time batch was sent
	lastBatch time.Time

	// Max characters in request
	maxLength int

	// http client
	client *retryablehttp.Client
}

type inputObject struct {
	req     *translator.Request
	outChan outputChannel
}

type returnObject struct {
	text string
	err  error
}

func NewBatchTranslator(length int, delay time.Duration) *BatchTranslator {
	q := &BatchTranslator{
		inCh:       make(inputChannel, runtime.NumCPU()),
		maxLength:  length,
		batchDelay: delay,
		lastBatch:  time.Now(),
		client:     buildClient(),
	}

	go q.worker()

	return q
}

func (q *BatchTranslator) worker() {
	var timePassed time.Duration

	for {
		var items []inputObject
		var totalLength int
		var totalCount int

		delay := q.batchDelay + time.Duration(rand.Intn(4000))*time.Millisecond

	ReadChannel:
		for (timePassed < delay || totalCount == 0) && totalLength < q.maxLength {
			timePassed = time.Since(q.lastBatch)
			select {
			case item := <-q.inCh:
				items = append(items, item)
				totalLength += len(item.req.Text)
				totalCount++
			case <-time.After(delay):
				timePassed = time.Since(q.lastBatch)
				break ReadChannel
			}
		}

		if timePassed < delay {
			sleep := q.batchDelay - timePassed
			log.Debugf("Throttling batch processing for %.1f seconds, Requests: %d Total request length: %d", sleep.Seconds(), len(items), totalLength)
			time.Sleep(sleep)
		}

		q.translateBatch(items)
	}
}

func (q *BatchTranslator) Join(req *translator.Request) (string, error) {
	outCh := make(outputChannel)

	i := inputObject{
		req:     req,
		outChan: outCh,
	}

	// Add request to queue
	q.inCh <- i

	out := <-outCh

	close(outCh)

	// Wait for response
	return out.text, out.err
}
