package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"gitgud.io/softashell/comfy-translator/translator"
	log "github.com/Sirupsen/logrus"
)

type Comfy int

func (t *Comfy) Translate(req *translator.Request, reply *translator.Response) error {
	if len(req.Text) < 1 || len(req.From) < 1 || len(req.To) < 1 {
		return fmt.Errorf("Empty arguments")
	}

	if req.From != "ja" || req.To != "en" {
		return fmt.Errorf("Unsupported languages")
	}

	response := translator.Response{
		TranslationText: translate(*req),
		From:            req.From,
		To:              req.To,
		Text:            req.Text,
	}

	*reply = response

	return nil
}

func ServeComfyRPC(listenAddr string) {
	comfy := new(Comfy)

	rpc.Register(comfy)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	http.Serve(l, nil)
}
