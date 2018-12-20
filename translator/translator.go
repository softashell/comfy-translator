package translator

import (
	"fmt"
	"time"

	"gitgud.io/softashell/comfy-translator/config"
	log "github.com/sirupsen/logrus"
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
	Start(c config.TranslatorConfig) error
	Enabled() bool
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

type BadTranslationError struct {
	Input  string
	Output string
}

func (e BadTranslationError) Error() string {
	return fmt.Sprintf("garbage translation: %q => %q", e.Input, e.Output)
}
