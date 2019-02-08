package google

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func buildClient() *retryablehttp.Client {
	httpClient := retryablehttp.NewClient()

	// Stop debug logger
	httpClient.Logger = nil
	httpClient.RetryMax = 1
	httpClient.RetryWaitMin = 30 * time.Second
	httpClient.RetryWaitMax = 2 * time.Minute

	return httpClient
}

func buildRequest(langFrom, langTo, inputText string) (*retryablehttp.Request, error) {
	var URL *url.URL
	URL, err := url.Parse("https://translate.googleapis.com/translate_a/single")

	parameters := url.Values{}
	parameters.Add("client", "gtx") // Google translate extension
	parameters.Add("dt", "t")       // Translate text
	parameters.Add("hl", "en")      // Interface language
	parameters.Add("sl", langFrom)  // Source language or "auto"
	parameters.Add("tl", langTo)    // Target language
	parameters.Add("ie", "UTF-8")   // Input encoding
	parameters.Add("oe", "UTF-8")   // Output encoding
	parameters.Add("q", inputText)  // Source text

	URL.RawQuery = parameters.Encode()

	r, err := retryablehttp.NewRequest("GET", URL.String(), nil)
	if err != nil {
		return r, errors.Wrap(err, "Failed to create request")
	}

	return r, err
}

func (q *BatchTranslator) translateBatch(items []inputObject) {
	if len(items) < 1 {
		log.Debug("nothing to do")
		return
	}

	log.Infof("processing %d items", len(items))

	q.lastBatch = time.Now()

	go func() {
		// TODO: Add support for different language pairs
		err := q.translateItems(items)
		if err != nil {
			// Send error to all items
			for _, i := range items {
				i.outChan <- returnObject{
					text: "",
					err:  err,
				}
			}
		}
	}()
}

func (q *BatchTranslator) translateItems(items []inputObject) error {
	// Join every input separated by newline
	var reqText string
	for _, i := range items {
		reqText += i.req.Text + "\n"
	}

	// FIXME: Properly batch multiple languages
	r, err := buildRequest(items[0].req.From, items[0].req.To, reqText)
	if err != nil {
		return err
	}

	resp, err := q.client.Do(r)
	if err != nil {
		log.Fatal(err)

		return errors.Wrapf(err, "Failed to do request (%s delay)", q.batchDelay)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if contents, err := ioutil.ReadAll(resp.Body); err == nil {
			return fmt.Errorf("%s - %s", resp.Status, contents)
		}

		return fmt.Errorf("%s", resp.Status)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Failed to read response body")
	}

	response, err := decodeResponse(string(contents))
	if err != nil {
		log.Errorf("Unknown response: %q", string(contents))
		return errors.Wrap(err, "Failed to decode response json")
	}

	/* Google sometimes splits sentences in input to multiple output objects
	attempt to merge them back into one */
	if len(items) != len(response) {
		log.Debug("Response pair count don't match input")
		response = mergeOutput(items, response)
	}

	for i, pair := range response {
		if strings.TrimSpace(pair.input) != strings.TrimSpace(items[i].req.Text) {
			items[i].outChan <- returnObject{
				text: pair.input,
				err:  fmt.Errorf("mismatched input text! %q != %q", items[i].req.Text, pair.input),
			}

			continue
		}

		items[i].outChan <- returnObject{
			text: pair.output,
			err:  nil,
		}
	}

	return nil
}
