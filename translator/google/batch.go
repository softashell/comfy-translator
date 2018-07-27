package google

import (
	"math/rand"
	"runtime"
	"time"

	"gitgud.io/softashell/comfy-translator/translator"
	"github.com/hashicorp/go-retryablehttp"

	log "github.com/Sirupsen/logrus"
)

type inputChannel chan inputObject
type outputChannel chan returnObject

type batchTranslator struct {
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

func NewBatchTranslator(length int, delay time.Duration) *batchTranslator {
	q := &batchTranslator{
		inCh:       make(inputChannel, runtime.NumCPU()),
		maxLength:  length,
		batchDelay: delay,
		lastBatch:  time.Now(),
		client:     buildClient(),
	}

	go q.worker()

	return q
}

func (q *batchTranslator) worker() {
	var timePassed time.Duration

	for {
		var items []inputObject
		var totalLength int

		// Add some random delay to requests
		delay := q.batchDelay + time.Duration(rand.Intn(4000))*time.Millisecond

	ReadChannel:
		for timePassed < delay && totalLength < q.maxLength {
			timePassed = time.Since(q.lastBatch)
			select {
			case item := <-q.inCh:
				items = append(items, item)
				totalLength += len(item.req.Text)
			default:
				break ReadChannel
			}
		}

		if len(items) < 1 {
			time.Sleep(time.Second)
			continue
		}

		if timePassed < delay {
			sleep := q.batchDelay - timePassed
			log.Debugf("Throttling batch processing for %.1f seconds, Requests: %d Total request length: %d", sleep.Seconds(), len(items), totalLength)
			time.Sleep(sleep)
		}

		q.translateBatch(items)
	}
}

func (q *batchTranslator) Join(req *translator.Request) (string, error) {
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
