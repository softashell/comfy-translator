package google

import (
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"unicode"

	"gitgud.io/softashell/comfy-translator/config"
	"gitgud.io/softashell/comfy-translator/translator"
)

const (
	userAgent    = "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko"
	defaultDelay = time.Second * 8
	// 4 works okay for a short while, 5 works for about 30 minutes
	// 6 works ok for since line requests
	// 8 seems safe for batch requests without token (~12 lines, ~600 chars each request)
)

var (
	garbageRegex = regexp.MustCompile(`\s?_{2,3}(\s\d)?`)
)

type Translate struct {
	enabled     bool
	client      *retryablehttp.Client
	lastRequest time.Time
	mutex       *sync.Mutex

	delay time.Duration
	batch *BatchTranslator
}

func New() *Translate {
	t := &Translate{
		lastRequest: time.Now(),
		mutex:       &sync.Mutex{},
		enabled:     false,

		delay: defaultDelay,
		batch: NewBatchTranslator(3000, defaultDelay),
	}

	return t
}

func (t *Translate) Name() string {
	return "Google"
}

func (t *Translate) Start(c config.TranslatorConfig) error {
	t.enabled = true

	return nil
}

func (t *Translate) Enabled() bool {
	return t.enabled
}

func (t *Translate) Translate(req *translator.Request) (string, error) {
	start := time.Now()

	t.lastRequest = time.Now()

	out, err := t.batch.Join(req)
	if err != nil {
		return "", errors.Wrap(err, "Failed to process request")
	}

	// Delete garbage output which often leaves the output empty, fix your shit google tbh
	out2 := garbageRegex.ReplaceAllString(out, "")
	if len(out) < 1 || (len(out2) < len(out)/2) {
		return "", translator.BadTranslationError{
			Input:  req.Text,
			Output: out,
		}
	}

	out = cleanText(out2)

	if IsTranslationGarbage(out) {
		return "", translator.BadTranslationError{
			Input:  req.Text,
			Output: out,
		}
	}

	log.WithFields(log.Fields{
		"time": time.Since(start),
	}).Debugf("Google: %q", out)

	return out, nil
}

func cleanText(text string) string {
	text = strings.Replace(text, "\\\\", "\\", -1)

	// Replace escaped quotes and newlines
	text = strings.Replace(text, "\\\"", "\"", -1)
	text = strings.Replace(text, "\\n", "\n", -1)

	// Replace raw characters
	text = strings.Replace(text, "\\u0026", "＆", -1)
	text = strings.Replace(text, "\\u003c", "<", -1)
	text = strings.Replace(text, "\\u003e", ">", -1)
	text = strings.Replace(text, "’", "'", -1)

	// Fix oddly placed '
	text = strings.Replace(text, " ' ", "'", -1)
	text = strings.Replace(text, " '", "'", -1)

	return text
}

func IsTranslationGarbage(text string) bool {
	text = strings.ToLower(text)
	if strings.Contains(text, "powered by discuz") || strings.Contains(text, "powered by translate") {
		return true
	}

	var rest int
	var japanese int

	for _, r := range text {
		if unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) || unicode.Is(unicode.Han, r) {
			japanese++
			continue
		}

		if unicode.IsSpace(r) {
			continue
		}

		rest++
	}

	if japanese > rest {
		return true
	}

	return false
}
