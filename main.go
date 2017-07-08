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
	t := []translator.Translator{
		google.New(),
		bing.New(),   // FIXME: Bing starts refusing connection pretty randomly and I can't tell what it doesn't like
		yandex.New(), // Pretty bad quality
	}

	log.Info("Starting translation engines")

	for i := range t {
		name := t[i].Name()

		c, found := conf.Translator[name]
		if !found {
			log.Errorf("Couldn't find config for %q", name)
		}

		if c.Enabled {
			log.Infof("%s: Starting", name)
			err := t[i].Start(c)
			if err != nil {
				log.Warn(err)
			}

			if t[i].Enabled() {
				log.Infof("%s: Should be started", name)
			} else {
				log.Infof("%s: Did not start", name)
			}
		} else {
			log.Infof("%s: Disabled in config", name)
		}

		err := cache.createBucket(name)
		if err != nil {
			log.Errorf("Failed to create missing bucket %q", name)
		}
	}
	/*

		sort.Slice(t, func(i, j int) bool {
			return conf.Translator[t[i].Name()].Priority < conf.Translator[t[j].Name()].Priority
		})
	*/

	var order string
	for i := range t {
		order += t[i].Name()

		if !t[i].Enabled() {
			order += " (only cache)"
		}

		if i+i < len(t) {
			order += ", "
		}
	}

	log.Infof("Translation order: %s", order)

	translators = t
}
