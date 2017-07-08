package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/negroni"
	"gopkg.in/tylerb/graceful.v1"

	"gitgud.io/softashell/comfy-translator/config"
	"gitgud.io/softashell/comfy-translator/translator"
	"gitgud.io/softashell/comfy-translator/translator/bing"
	"gitgud.io/softashell/comfy-translator/translator/google"
	"gitgud.io/softashell/comfy-translator/translator/yandex"
)

var (
	cache       *Cache
	conf        *config.Config
	translators []translator.Translator
)

func main() {
	log.Info("Starting comfy translator")

	debug := os.Getenv("DEBUG")
	if len(debug) > 0 {
		log.SetLevel(log.DebugLevel)
	}

	m := http.NewServeMux()
	m.HandleFunc("/api/translate", translateHandler)

	n := negroni.New()
	n.UseHandler(m)

	var err error

	conf = config.NewConfig()
	err = conf.Load("comfy-translator.toml")
	if err != nil {
		log.Fatal(err)
	}

	cache, err = NewCache(conf.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize translation cache: %v", err)
	}
	defer cache.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = conf.Port
	}

	startTranslators()

	listenAddr := fmt.Sprintf("%s:%s", conf.Host, port)

	log.WithFields(log.Fields{
		"addr": listenAddr,
	}).Info("Ready to accept connections")

	err = graceful.RunWithErr(listenAddr, 60*time.Second, n)
	if err != nil {
		log.Fatal(err)
	}
}

func startTranslators() {
	translators = append(translators,
		google.New(),
		bing.New(),   // FIXME: Bing starts refusing connection pretty randomly and I can't tell what it doesn't like
		yandex.New(), // Pretty bad quality
	)

	log.Info("Starting translation engines")

	for _, t := range translators {
		c, found := conf.Translator[t.Name()]
		if !found {
			log.Errorf("Couldn't find config for %q", t.Name())
		}

		if c.Enabled {
			log.Infof("%s: Starting", t.Name())
			err := t.Start(c)
			if err != nil {
				log.Warn(err)
			}
		} else {
			log.Infof("%s: Disabled in config", t.Name())
		}

		err := cache.createBucket(t.Name())
		if err != nil {
			log.Errorf("Failed to create missing bucket %q", t.Name())
		}
	}
}
