package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	delay = time.Second
)

var (
	client              = &http.Client{}
	lastGoogleRequest   = time.Now()
	lastTransltrRequest = time.Now()
	lastHonyakuRequest  = time.Now()
)

func checkThrottle(lastReq time.Time) {
	timePassed := time.Since(lastReq)
	if timePassed < delay {
		sleep := delay - timePassed
		log.Debugf("Throttling request for %f seconds", sleep.Seconds())
		time.Sleep(sleep)
	}
}

func translateWithGoogle(req *translateRequest) (string, error) {
	start := time.Now()

	checkThrottle(lastGoogleRequest)

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
	check(err)

	r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko")

	lastGoogleRequest = time.Now()

	resp, err := client.Do(r)
	check(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%+v", resp)
	}

	// [[["It will be saved","助かるわい",,,3]],,"ja"]
	contents, err := ioutil.ReadAll(resp.Body)
	check(err)

	allStrings := regexp.MustCompile("\"(.+?)\",\"(.+?)\",?").FindAllStringSubmatch(string(contents), -1)

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
	out2 := regexp.MustCompile(`\s?_{2,3}(\s\d)?`).ReplaceAllString(out, "")
	if len(out) < 1 || (len(out2) < len(out)/2) {
		return "", fmt.Errorf("Bad response %q", contents)
	}

	out = out2

	// Replace escaped quotes
	out = strings.Replace(out, "\\\"", "\"", -1)

	log.WithFields(log.Fields{
		"time": time.Since(start),
	}).Debugf("Google: %q", out)

	return out, nil
}

func translateWithTransltr(req *translateRequest) (string, error) {
	start := time.Now()

	checkThrottle(lastTransltrRequest)

	// Convert json object to string
	jsonString, err := json.Marshal(req)
	if err != nil {
		log.Error("Failed to marshal JSON API request", err.Error())
	}

	lastTransltrRequest = time.Now()

	// Post the request
	resp, reply, errs := gorequest.New().Post("http://transltr.org/api/translate").Send(string(jsonString)).EndBytes()
	for _, err := range errs {
		log.WithFields(log.Fields{
			"response": resp,
			"reply":    reply,
		}).Error(err.Error())
		return "", err
	}

	var response translateResponse

	if err := json.Unmarshal(reply, &response); err != nil {
		log.Error("Failed to unmarshal JSON API response", err.Error())
		return "", err
	}

	var out string
	out = response.TranslationText

	// Seems to use google translate as backend as well so it will output the same garbage
	out2 := regexp.MustCompile(`\s?_{2,3}(\s\d)?`).ReplaceAllString(out, "")
	if len(out) < 1 || (len(out2) < len(out)/2) {
		// Output it anway since this is the last translation for now
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

	checkThrottle(lastHonyakuRequest)

	var URL *url.URL
	URL, err := url.Parse("http://honyaku.yahoo.co.jp/transtext")
	check(err)

	parameters := url.Values{}
	parameters.Add("both", "TH")
	parameters.Add("eid", "CR-JE")
	parameters.Add("text", req.Text)

	URL.RawQuery = parameters.Encode()

	r, err := http.NewRequest("GET", URL.String(), nil)
	check(err)

	r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko")

	lastHonyakuRequest = time.Now()

	resp, err := client.Do(r)
	check(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		contents, err := ioutil.ReadAll(resp.Body)
		check(err)
		return "", fmt.Errorf("%d %s", resp.StatusCode, contents)
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
