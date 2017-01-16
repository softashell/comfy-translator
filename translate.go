package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	delay      = 1500 * time.Millisecond
	requestURL = "https://translate.google.com/"
)

var (
	client      = &http.Client{}
	lastRequest = time.Now()
)

func googleTranslate(text string, from string, to string) (string, error) {
	timePassed := time.Since(lastRequest)
	if timePassed < delay {
		sleep := delay - timePassed
		log.Printf("Throttling google translate request for %f seconds", sleep.Seconds())
		time.Sleep(sleep)
	}

	var URL *url.URL
	URL, err := url.Parse(requestURL)
	URL.Path += "/translate_a/single"

	parameters := url.Values{}
	parameters.Add("client", "gtx")
	parameters.Add("dt", "t")
	parameters.Add("sl", from)
	parameters.Add("tl", to)
	parameters.Add("ie", "UTF-8")
	parameters.Add("oe", "UTF-8")
	parameters.Add("q", text)

	// /translate_a/single?client=gtx&dt=t&sl=%hs&tl=%hs&ie=UTF-8&oe=UTF-8&q=%s
	URL.RawQuery = parameters.Encode()

	req, err := http.NewRequest("GET", URL.String(), nil)
	check(err)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko")
	req.Header.Set("Referer", requestURL) // TODO: Probably not needed ??

	lastRequest = time.Now()

	resp, err := client.Do(req)
	defer resp.Body.Close()
	check(err)

	if resp.StatusCode != http.StatusOK {
		contents, err := ioutil.ReadAll(resp.Body)
		check(err)
		return "", fmt.Errorf("%d %s", resp.StatusCode, contents)
	}

	// [[["It will be saved","助かるわい",,,3]],,"ja"]
	contents, err := ioutil.ReadAll(resp.Body)
	check(err)

	reg, err := regexp.Compile("\"(.+?)\"")
	check(err)

	var allStrings []string
	allStrings = reg.FindAllString(string(contents), 2)

	if len(allStrings) < 1 {
		return "", fmt.Errorf("Bad response %s", contents)
	}

	s := allStrings[0]
	s = strings.Trim(s, "\"")

	return s, nil
}
