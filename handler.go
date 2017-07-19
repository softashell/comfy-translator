package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitgud.io/softashell/comfy-translator/translator"
)

func translateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req translator.Request

	switch r.Method {
	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)

			if err := json.NewEncoder(w).Encode(err); err != nil {
				panic(err)
			}

			return
		}
		defer r.Body.Close()
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
