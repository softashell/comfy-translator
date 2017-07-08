package yandex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"strings"

	"gitgud.io/softashell/comfy-translator/config"

	"gitgud.io/softashell/comfy-translator/translator"
	log "github.com/Sirupsen/logrus"
)

const (
	apiURL = "https://translate.yandex.net/api/v1.5/tr.json/translate"
	delay  = time.Second
)

type Translate struct {
	client      *http.Client
	lastRequest time.Time
	mutex       *sync.Mutex

	apiKey string
}

type yandexResponse struct {
	Code int      `json:"code"`
	Lang string   `json:"lang"`
	Text []string `json:"text"`
}

func New() *Translate {
	return &Translate{
		client:      &http.Client{Timeout: (10 * time.Second)},
		lastRequest: time.Now(),
		mutex:       &sync.Mutex{},
	}
}

func (t Translate) Name() string {
	return "Yandex"
}

func (t Translate) Start(c config.TranslatorConfig) error {
	t.apiKey = c.Key
	if len(t.apiKey) < 1 || !strings.HasPrefix(t.apiKey, "trnsl.") {
		return fmt.Errorf("%s: Invalid api key provided, edit comfy-translator.toml to disable or change key", t.Name())
	}

	return nil
}

func (t Translate) Translate(req *translator.Request) (string, error) {
	start := time.Now()

	t.mutex.Lock()
	translator.CheckThrottle(t.lastRequest, delay)
	defer t.mutex.Unlock()

	var URL *url.URL
	URL, err := url.Parse(apiURL)

	parameters := url.Values{}
	parameters.Add("key", t.apiKey)
	parameters.Add("text", req.Text)
	parameters.Add("lang", req.From+"-"+req.To)
	parameters.Add("format", "plain")

	// https://tech.yandex.com/translate/doc/dg/reference/translate-docpage
	URL.RawQuery = parameters.Encode()

	r, err := http.NewRequest("POST", URL.String(), nil)
	if err != nil {
		log.Errorln("Failed to create request", err)
		return "", err
	}

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

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body", err)
		return "", err
	}

	var response yandexResponse
	if err := json.Unmarshal(contents, &response); err != nil {
		log.Errorln("Failed to unmarshal JSON API response", err)
		return "", err
	}

	if len(response.Text) < 1 {
		return "", fmt.Errorf("Empty response")
	} else if len(response.Text) > 1 {
		log.Warning("More than one item in response: %s", string(contents))
	}

	var out string

	for i := range response.Text {
		out += response.Text[i]
	}

	log.WithFields(log.Fields{
		"time": time.Since(start),
	}).Debugf("Yandex: %q", out)

	return out, nil
}
