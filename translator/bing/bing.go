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

	"golang.org/x/net/publicsuffix"

	log "github.com/Sirupsen/logrus"
	cookiejar "github.com/juju/persistent-cookiejar"

	"gitgud.io/softashell/comfy-translator/translator"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.10136"
	delay     = time.Second * 15

	translatorURL = "https://www.bing.com/translator"
	translatorAPI = "https://www.bing.com/translator/api/Translate/TranslateArray"
)

type Translate struct {
	client           *http.Client
	lastRequest      time.Time
	mutex            *sync.Mutex
	cookieExpiration time.Time
	jar              *cookiejar.Jar
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
	log.Info("Starting bing translation engine")

	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
		Filename:         cookiejar.DefaultCookieFile(),
	})
	if err != nil {
		log.Fatal(err)
	}

	cookieExpiration := getExpiration(jar)

	client := &http.Client{
		Jar:     jar,
		Timeout: (10 * time.Second),
	}

	return &Translate{
		client:           client,
		lastRequest:      time.Now(),
		mutex:            &sync.Mutex{},
		cookieExpiration: cookieExpiration,
		jar:              jar,
	}
}

func (t Translate) Name() string {
	return "Bing"
}

func (t Translate) Translate(req *translator.Request) (string, error) {
	if time.Now().After(t.cookieExpiration) {
		err := t.getCookies()
		if err != nil {
			return "", err
		}
	}

	log.Debugf("Translating %q from %q to %q", req.Text, req.From, req.To)

	t.mutex.Lock()
	defer t.mutex.Unlock()

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
		log.Warning("More than one item in response: %s", string(contents))
	}

	return response.Items[0].Text, nil
}

func (t Translate) getCookies() error {
	log.Debug("Getting bing cookies")

	t.mutex.Lock()
	defer t.mutex.Unlock()

	translator.CheckThrottle(t.lastRequest, delay)

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
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.Status)
	}

	t.cookieExpiration = getExpiration(t.jar)

	err = t.jar.Save()
	if err != nil {
		log.Error(err)
	}

	return nil
}

func getExpiration(jar *cookiejar.Jar) time.Time {
	var URL *url.URL
	URL, err := url.Parse("http://www.bing.com/translator")
	if err != nil {
		panic(err)
	}

	if len(jar.Cookies(URL)) > 3 {
		for _, c := range jar.Cookies(URL) {
			if c.Name == "MUID" {
				return c.Expires
			}
		}
	}

	return time.Now()
}
