package main

import (
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"gitgud.io/softashell/comfy-translator/translator"
)

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

			if !t.Enabled() {
				continue
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
