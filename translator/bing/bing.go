package bing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"

	log "github.com/Sirupsen/logrus"

	"gitgud.io/softashell/comfy-translator/config"
	"gitgud.io/softashell/comfy-translator/translator"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.10136"
	delay     = time.Second * 20 // TODO: Needs tweaking

	translatorURL = "http://www.bing.com/translator"
	translatorAPI = "http://www.bing.com/translator/api/Translate/TranslateArray"
)

type Translate struct {
	enabled     bool
	client      *http.Client
	lastRequest time.Time
	mutex       *sync.Mutex

	requests int

	cookieExpiration time.Time
}

type translateArrayRequest struct {
	Text string `json:"text"`
}

type translateResponse struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Items []struct {
		Text          string `json:"text"`
		WordAlignment string `json:"wordAlignment"`
	} `json:"items"`
}

func New() *Translate {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: (10 * time.Second),
	}

	return &Translate{
		client:      client,
		lastRequest: time.Now(),
		mutex:       &sync.Mutex{},

		cookieExpiration: time.Now(),
	}
}

func (t *Translate) Name() string {
	return "Bing"
}

func (t *Translate) Start(c config.TranslatorConfig) error {
	err := t.getCookies()
	if err != nil {
		return err
	}

	t.enabled = true

	return nil
}

func (t *Translate) Enabled() bool {
	return t.enabled
}

func (t *Translate) Translate(req *translator.Request) (string, error) {
	log.Debugf("Translating %q from %q to %q", req.Text, req.From, req.To)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	translator.CheckThrottle(t.lastRequest, delay)

	if time.Now().After(t.cookieExpiration) || t.requests > 3 {
		err := t.getCookies()
		if err != nil {
			return "", err
		}
	}

	translator.CheckThrottle(t.lastRequest, delay)

	var URL *url.URL
	URL, err := url.Parse(translatorAPI)

	parameters := url.Values{}
	parameters.Add("from", req.From)
	parameters.Add("to", req.To)

	URL.RawQuery = parameters.Encode()

	var tr []translateArrayRequest
	tr = append(tr, translateArrayRequest{
		Text: req.Text,
	})

	// Convert json object to string
	jsonString, err := json.Marshal(tr)
	if err != nil {
		log.Errorln("Failed to marshal JSON API request", err)
		return "", err
	}

	r, err := http.NewRequest("POST", URL.String(), bytes.NewBuffer(jsonString))
	if err != nil {
		log.Errorln("Failed to create request", err)
		return "", err
	}

	r.Header.Set("User-Agent", userAgent)
	r.Header.Set("Content-Type", "application/json")
	r.Close = true

	t.lastRequest = time.Now()
	t.requests++

	resp, err := t.client.Do(r)
	if err != nil {
		log.Errorln("Failed to do request", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", resp.Status)
	}

	//{"from":"en","to":"ja","items":[{"text":"ハローワールド！","wordAlignment":""}]}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln("Failed to read response body", err)
		return "", err
	}

	var response translateResponse
	if err := json.Unmarshal(contents, &response); err != nil {
		log.Errorln("Failed to unmarshal JSON API response", err)
		return "", err
	}

	if len(response.Items) < 1 {
		return "", fmt.Errorf("Empty response")
	} else if len(response.Items) > 1 {
		// Never has happened before, maybe I should panic if it does to avoid responses not handled properly
		log.Warningf("More than one item in response: %s", string(contents))
	}

	var out string

	for i := range response.Items {
		out += response.Items[i].Text
	}

	return out, nil
}

func (t *Translate) getCookies() error {
	log.Print("Getting bing cookies")

	var URL *url.URL
	URL, err := url.Parse(translatorURL)
	if err != nil {
		panic(err)
	}

	r, err := http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		log.Errorln("Failed to create request", err)
		return err
	}

	r.Header.Set("User-Agent", userAgent)

	t.lastRequest = time.Now()

	resp, err := t.client.Do(r)
	if err != nil {
		log.Errorln("Failed to do request", err)

		t.cookieExpiration = time.Now()

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}

	t.cookieExpiration = time.Now().Add(time.Minute * 10)

	return nil
}
