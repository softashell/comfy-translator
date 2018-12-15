package main

import (
	"fmt"
	"os"
	"sort"

	log "github.com/Sirupsen/logrus"

	"gitgud.io/softashell/comfy-translator/cache"
	"gitgud.io/softashell/comfy-translator/config"
	"gitgud.io/softashell/comfy-translator/translator"
	"gitgud.io/softashell/comfy-translator/translator/bing"
	"gitgud.io/softashell/comfy-translator/translator/google"
	"gitgud.io/softashell/comfy-translator/translator/yandex"
)

var (
	c           *cache.Cache
	q           *Queue
	conf        *config.Config
	translators []translator.Translator
)

func main() {
	log.Info("Starting comfy translator")

	debug := os.Getenv("DEBUG")
	if len(debug) > 0 {
		log.SetLevel(log.DebugLevel)
	}

	var err error

	conf = config.NewConfig()
	err = conf.Load("comfy-translator.toml")
	if err != nil {
		log.Fatal(err)
	}

	c, err = cache.NewCache(conf.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize translation cache: %v", err)
	}
	defer c.Close()

	q = NewQueue()

	port := os.Getenv("PORT")
	if port == "" {
		port = conf.Port
	}

	startTranslators()

	listenAddr := fmt.Sprintf("%s:%s", conf.Host, port)

	log.WithFields(log.Fields{
		"addr": listenAddr,
	}).Info("Ready to accept connections")

	ServeComfyRPC(listenAddr)
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

		conf, found := conf.Translator[name]
		if !found {
			log.Errorf("Couldn't find config for %q", name)
		}

		if conf.Enabled {
			log.Infof("%s: Starting", name)
			err := t[i].Start(conf)
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
	}

	sort.Slice(t, func(i, j int) bool {
		return conf.Translator[t[i].Name()].Priority < conf.Translator[t[j].Name()].Priority
	})

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
