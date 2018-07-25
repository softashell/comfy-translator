package google

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"unicode"

	"gitgud.io/softashell/comfy-translator/config"
	"gitgud.io/softashell/comfy-translator/translator"
)

const (
	userAgent    = "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko"
	defaultDelay = time.Second * 6
	// 4 works okay for a short while, 5 works for about 30 minutes
)

var (
	garbageRegex = regexp.MustCompile(`\s?_{2,3}(\s\d)?`)
)

type Translate struct {
	enabled     bool
	client      *retryablehttp.Client
	lastRequest time.Time
	mutex       *sync.Mutex

	errorCount uint
	delay      time.Duration
	delayMutex *sync.Mutex
}

func New() *Translate {
	httpClient := retryablehttp.NewClient()

	// Stop debug logger
	httpClient.Logger = nil
	httpClient.RetryMax = 1
	httpClient.RetryWaitMin = 30 * time.Second
	httpClient.RetryWaitMax = 2 * time.Minute

	return &Translate{
		client:      httpClient,
		lastRequest: time.Now(),
		mutex:       &sync.Mutex{},
		enabled:     false,

		errorCount: 0,
		delay:      defaultDelay,
		delayMutex: &sync.Mutex{},
	}
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

func (t *Translate) incError() {
	t.delayMutex.Lock()
	defer t.delayMutex.Unlock()

	t.errorCount++

	t.delay += defaultDelay * time.Duration(t.errorCount)
}

func (t *Translate) decError() {
	t.delayMutex.Lock()
	defer t.delayMutex.Unlock()

	if t.errorCount < 1 {
		return
	}

	t.errorCount++

	subDelay := defaultDelay * time.Duration(t.errorCount)

	if t.delay > defaultDelay {
		if t.delay < subDelay || t.delay-subDelay < defaultDelay {
			t.delay = defaultDelay
		} else {
			t.delay -= subDelay
		}
	}
}

func (t *Translate) Translate(req *translator.Request) (string, error) {
	start := time.Now()

	t.mutex.Lock()
	translator.CheckThrottle(t.lastRequest, t.delay+(time.Duration(rand.Intn(3000))*time.Millisecond))
	defer t.mutex.Unlock()

	var URL *url.URL
	URL, err := url.Parse("https://translate.googleapis.com/translate_a/single")

	parameters := url.Values{}
	parameters.Add("client", "gtx") // Google translate extension
	parameters.Add("dt", "t")       // Translate text
	parameters.Add("hl", "en")      // Interface language
	parameters.Add("sl", req.From)  // Source language or "auto"
	parameters.Add("tl", req.To)    // Target language
	parameters.Add("ie", "UTF-8")   // Input encoding
	parameters.Add("oe", "UTF-8")   // Output encoding
	parameters.Add("q", req.Text)   // Source text

	URL.RawQuery = parameters.Encode()

	r, err := retryablehttp.NewRequest("GET", URL.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, "Failed to create request")
	}

	r.Header.Set("User-Agent", userAgent)

	t.lastRequest = time.Now()

	resp, err := t.client.Do(r)
	if err != nil {
		t.incError()

		return "", errors.Wrapf(err, "Failed to do request (%d errors, %s delay)", t.errorCount, t.delay)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if contents, err := ioutil.ReadAll(resp.Body); err == nil {
			return "", fmt.Errorf("%s - %s", resp.Status, contents)
		}

		return "", fmt.Errorf("%s", resp.Status)
	}

	// [[["It will be saved","助かるわい",,,3]],,"ja"]
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "Failed to read response body")
	}

	response, err := decodeResponse(string(contents))
	if err != nil {
		log.Error("Unknown response: %q", string(contents))
		return "", errors.Wrap(err, "Failed to decode response json")
	}

	var out string
	for inText, outText := range response {
		if inText != req.Text {
			log.Warnf("mismatched input text! %q != %q", req.Text, inText)
		}
		out += outText
	}

	// Reduce delay between requests
	t.decError()

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
