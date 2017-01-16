package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/negroni"
	"gopkg.in/tylerb/graceful.v1"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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
	log.Printf("Starting!")

	m := http.NewServeMux()

	m.HandleFunc("/api/translate", translateHandler)

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.UseHandler(m)

	cache = NewCache()

	err := graceful.RunWithErr("localhost:7001", 60*time.Second, n)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Exiting!")
}

func translateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if r.Method != "POST" {
		err := fmt.Errorf("Only accepting JSON POST ")
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	var req translateRequest
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &req); err != nil {
		log.Println(string(body))
		log.Println(req)

		w.WriteHeader(http.StatusUnprocessableEntity)

		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}

		return
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
	found, output := cache.Get(req.Text)

	var err error

	if !found {
		output, err = googleTranslate(req.Text, req.From, req.To)
		check(err)
		log.Println("gtranslate =>", output)

		// TODO: Fallbak translation if google returns nothing or fails

		cache.Put(req.Text, output)
	}

	return output
}
