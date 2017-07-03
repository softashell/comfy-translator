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

	"gitgud.io/softashell/comfy-translator/translator"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko"
	delay     = time.Second / 2
)

var (
	allStringRegex = regexp.MustCompile("\"(.+?)\",\"(.+?)\",?")
	garbageRegex   = regexp.MustCompile(`\s?_{2,3}(\s\d)?`)
)

type Translate struct {
	client      *http.Client
	lastRequest time.Time
	mutex       *sync.Mutex
}

func New() Translate {
	log.Info("Starting google translate engine")

	return Translate{
		client:      &http.Client{Timeout: (10 * time.Second)},
		lastRequest: time.Now(),
		mutex:       &sync.Mutex{},
	}
}

func (t Translate) Name() string {
	return "Google"
}

func (t Translate) Translate(req *translator.Request) (string, error) {
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

	r, err := http.NewRequest("GET", URL.String(), nil)
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
		return "", fmt.Errorf("Bad response %q", out)
	}

	out = out2

	out = strings.Replace(out, "\\\\", "\\", -1)

	// Replace escaped quotes
	out = strings.Replace(out, "\\\"", "\"", -1)

	// Replace escaped newlines
	out = strings.Replace(out, "\\n", "\n", -1)

	log.WithFields(log.Fields{
		"time": time.Since(start),
	}).Debugf("Google: %q", out)

	return out, nil
}
