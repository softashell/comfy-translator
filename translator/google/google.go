package google

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-retryablehttp"

	"unicode"

	"gitgud.io/softashell/comfy-translator/config"
	"gitgud.io/softashell/comfy-translator/translator"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko"
	delay     = time.Second * 2
)

var (
	allStringRegex = regexp.MustCompile("\"(.+?)\",\"(.+?)\",?")
	garbageRegex   = regexp.MustCompile(`\s?_{2,3}(\s\d)?`)
)

type Translate struct {
	enabled     bool
	client      *retryablehttp.Client
	lastRequest time.Time
	mutex       *sync.Mutex
}

func New() *Translate {
	httpClient := retryablehttp.NewClient()

	// Stop debug logger
	httpClient.Logger = nil

	return &Translate{
		client:      httpClient,
		lastRequest: time.Now(),
		mutex:       &sync.Mutex{},
		enabled:     false,
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

func (t *Translate) Translate(req *translator.Request) (string, error) {
	start := time.Now()

	t.mutex.Lock()
	translator.CheckThrottle(t.lastRequest, delay)
	defer t.mutex.Unlock()

	var URL *url.URL
	URL, err := url.Parse("https://translate.google.com/translate_a/single")

	parameters := url.Values{}
	parameters.Add("client", "gtx")
	parameters.Add("dt", "t")
	parameters.Add("sl", req.From)
	parameters.Add("tl", req.To)
	parameters.Add("ie", "UTF-8")
	parameters.Add("oe", "UTF-8")
	parameters.Add("q", req.Text)

	// /translate_a/single?client=gtx&dt=t&sl=%hs&tl=%hs&ie=UTF-8&oe=UTF-8&q=%s
	URL.RawQuery = parameters.Encode()

	r, err := retryablehttp.NewRequest("GET", URL.String(), nil)
	if err != nil {
		log.Errorln("Failed to create request", err)
		return "", err
	}

	r.Header.Set("User-Agent", userAgent)

	t.lastRequest = time.Now()

	resp, err := t.client.Do(r)
	if err != nil {
		log.Errorln("Failed to do request", err)
		return "", err
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
		log.Error("Failed to read response body", err)
		return "", err
	}

	allStrings := allStringRegex.FindAllStringSubmatch(string(contents), -1)

	if len(allStrings) < 1 {
		return "", fmt.Errorf("Bad response %s", contents)
	}

	var out string
	for _, v := range allStrings {
		if len(v) < 3 {
			continue
		}

		out += v[1]
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
