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
)

type translateRequest struct {
	Text string `json:"text"`
	From string `json:"from"`
	To   string `json:"to"`
}

type translateResponse struct {
	Text            string `json:"text"`
	From            string `json:"from"`
	To              string `json:"to"`
	TranslationText string `json:"translationText"`
}

var cache *Cache

func main() {
	log.SetLevel(log.DebugLevel)

	m := http.NewServeMux()

	m.HandleFunc("/api/translate", translateHandler)

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.UseHandler(m)

	cache = NewCache()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	listenAddr := fmt.Sprintf("127.0.0.1:%s", port)

	log.WithFields(log.Fields{
		"addr": listenAddr,
	}).Info("Starting comfy translator")

	err := graceful.RunWithErr(listenAddr, 60*time.Second, n)
	if err != nil {
		log.Fatal(err)
	}
}

func translateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req translateRequest

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
			log.Printf("%+v %s", req, string(body))

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

	response := translateResponse{}
	response.TranslationText = translate(req)
	response.From = req.From
	response.To = req.To
	response.Text = req.Text

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func translate(req translateRequest) string {
	if len(strings.TrimSpace(req.Text)) < 1 {
		return req.Text
	}

	log.Printf("Input: %q", req.Text)

	found, out := cache.Get(req.Text)

	var err error

	if !found {
		out, err = translateWithGoogle(&req)
		if err != nil {
			log.Warning("Google:", err)
			out, err = translateWithTransltr(&req)
		}

		if err != nil {
			log.Warning("Transltr:", err)

			// Not going to cache this since it's probably garbage as well
			out, err = translateWithHonyaku(&req)
			if err != nil {
				log.Warning("Honyaku:", err)
			}
		} else if len(out) > 0 {
			cache.Put(req.Text, out)
		}
	}

	return out
}
