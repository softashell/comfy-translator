package main

import (
	"net/http"
	"sync"
	"time"
)

var (
	client = &http.Client{Timeout: (4 * time.Second)}

	lastTransltrRequest = time.Now()
	lastHonyakuRequest  = time.Now()

	transltrMutex = &sync.Mutex{}
	honyakuMutex  = &sync.Mutex{}
)

/*
func translateWithTransltr(req *translateRequest) (string, error) {
	start := time.Now()

	transltrMutex.Lock()
	checkThrottle(lastTransltrRequest)
	defer transltrMutex.Unlock()

	// Convert json object to string
	jsonString, err := json.Marshal(req)
	if err != nil {
		log.Errorln("Failed to marshal JSON API request", err)
		return "", err
	}

	r, err := http.NewRequest("POST", "http://transltr.org/api/translate", bytes.NewBuffer(jsonString))
	if err != nil {
		log.Errorln("Failed to create request", err)
		return "", err
	}

	r.Header.Set("Content-Type", "application/json")

	lastTransltrRequest = time.Now()

	resp, err := client.Do(r)
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
		log.Errorln("Failed to read response body", err)
		return "", err
	}

	var response translateResponse
	if err := json.Unmarshal(contents, &response); err != nil {
		log.Errorln("Failed to unmarshal JSON API response", err)
		return "", err
	}

	out := response.TranslationText

	// Seems to use google translate as backend as well so it will mostly output the same garbage
	out2 := regexp.MustCompile(`\s?_{2,3}(\s\d)?`).ReplaceAllString(out, "")
	if len(out) < 1 || (len(out2) < len(out)/2) {
		return "", fmt.Errorf("Garbage translation %q", out)
	}

	out = out2

	log.WithFields(log.Fields{
		"time": time.Since(start),
	}).Debugf("Transltr: %q", out)

	return out, nil
}

func translateWithHonyaku(req *translateRequest) (string, error) {
	start := time.Now()

	honyakuMutex.Lock()
	checkThrottle(lastHonyakuRequest)
	defer honyakuMutex.Unlock()

	var URL *url.URL
	URL, err := url.Parse("http://honyaku.yahoo.co.jp/transtext")
	check(err)

	parameters := url.Values{}
	parameters.Add("both", "TH")
	parameters.Add("eid", "CR-JE")
	parameters.Add("text", req.Text)

	URL.RawQuery = parameters.Encode()

	r, err := http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		log.Errorln("Failed to create request", err)
		return "", err
	}

	r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko")

	lastHonyakuRequest = time.Now()

	resp, err := client.Do(r)
	if err != nil {
		log.Errorln("Failed to do request", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return "", fmt.Errorf("Can't open response document %s", err)
	}

	out := doc.Find("#transafter").Text()
	out = strings.TrimSpace(out)

	if len(out) < 1 {
		return "", fmt.Errorf("Bad response %q", out)
	}

	log.WithFields(log.Fields{
		"time": time.Since(start),
	}).Debugf("Honyaku: %q", out)

	return out, nil
}

*/
