package translator

import (
	"time"

	"github.com/ewhal/nyaa/util/log"
)

const (
	delay = time.Second / 2
)

type Request struct {
	Text string `json:"text"`
	From string `json:"from"`
	To   string `json:"to"`
}

type Response struct {
	Text            string `json:"text"`
	From            string `json:"from"`
	To              string `json:"to"`
	TranslationText string `json:"translationText"`
}

type Translator interface {
	Name() string
	Translate(*Request) (string, error)
}

func CheckThrottle(lastReq time.Time, delay time.Duration) {
	timePassed := time.Since(lastReq)
	if timePassed < delay {
		sleep := delay - timePassed
		log.Debugf("Throttling request for %f seconds", sleep.Seconds())
		time.Sleep(sleep)
	}
}
