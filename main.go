package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/negroni"
	"gopkg.in/tylerb/graceful.v1"

	"gitgud.io/softashell/comfy-translator/translator"
	"gitgud.io/softashell/comfy-translator/translator/bing"
	"gitgud.io/softashell/comfy-translator/translator/google"
)

var (
	cache       *Cache
	translators []translator.Translator
)

func main() {
	m := http.NewServeMux()

	m.HandleFunc("/api/translate", translateHandler)

	n := negroni.New()
	//n.Use(negroni.NewLogger())
	n.UseHandler(m)

	var err error

	cache, err = NewCache()
	if err != nil {
		log.Fatalf("Failed to initialize translation cache: %v", err)
	}
	defer cache.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	debug := os.Getenv("DEBUG")
	if len(debug) > 0 {
		log.SetLevel(log.DebugLevel)
	}

	listenAddr := fmt.Sprintf("127.0.0.1:%s", port)

	log.WithFields(log.Fields{
		"addr": listenAddr,
	}).Info("Starting comfy translator")

	startTranslators()

	err = graceful.RunWithErr(listenAddr, 60*time.Second, n)
	if err != nil {
		log.Fatal(err)
	}
}

func translateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req translator.Request

	switch r.Method {
	case "POST":
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			panic(err)
		}
		if err := r.Body.Close(); err != nil {
			panic(err)
		}

		if err := json.Unmarshal(body, &req); err != nil {
			log.Errorf("%+v %s", req, string(body))

			w.WriteHeader(http.StatusUnprocessableEntity)

			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}

			return
		}
	case "GET":
		if err := r.ParseForm(); err != nil {
			panic(err)
		}

		req.Text = r.Form.Get("text")
		req.From = r.Form.Get("from")
		req.To = r.Form.Get("to")

		fmt.Println(req.Text)
	default:
		err := fmt.Errorf("Bad request method")
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	if len(req.Text) < 1 || len(req.From) < 1 || len(req.To) < 1 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		err := fmt.Errorf("Empty arguments")
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	if req.From != "ja" || req.To != "en" {
		w.WriteHeader(http.StatusBadRequest)
		err := fmt.Errorf("Unsupported languages")
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)

		}
		return
	}

	response := translator.Response{
		TranslationText: translate(req),
		From:            req.From,
		To:              req.To,
		Text:            req.Text,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func startTranslators() {
	translators = append(translators,
		google.New(),

		// FIXME: Bing starts refusing connection pretty randomly and I can't tell what it doesn't like
		bing.New(),
	)

	for _, t := range translators {
		err := cache.createBucket(t.Name())
		if err != nil {
			log.Errorf("Failed to create missing bucket %q", t.Name())
		}
	}
}

func translate(req translator.Request) string {
	if len(strings.TrimSpace(req.Text)) < 1 {
		return req.Text
	}

	start := time.Now()

	var err error
	var out, source string

	// TODO: Old generic cache, probably should migrate contents to Google bucket since that's where most of it came from
	found, out := cache.Get("translations", req.Text)
	if !found {
		for _, t := range translators {
			source = t.Name()

			log.Debugf("Translating with %s", source)

			found, out = cache.Get(source, req.Text)
			if found {
				source = source + "(cache)"
				break
			}

			out, err = t.Translate(&req)
			if err != nil {
				log.Warnf("%s: %s", source, err)
				continue
			}

			if len(out) > 0 {
				err = cache.Put(source, req.Text, out)
				if err != nil {
					log.WithFields(log.Fields{
						"err": err,
					}).Errorf("Failed to save result to %s cache", source)
				}
			}

			break
		}

		if len(out) < 1 {
			// TODO: Return original text or try to handle error in handler
			log.Errorf("All services failed to translate %q", req.Text)
		}
	} else {
		source = "Cache"
	}

	log.WithFields(log.Fields{
		"time":   time.Since(start),
		"source": source,
	}).Infof("%q -> %q", req.Text, out)

	return out

}
