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
	var found bool
	var out, source string

	// Checks if there are pending translation jobs for current request and wait for them to be completed
	if ch, wait := q.Join(req); wait {
		out := <-ch

		log.WithFields(log.Fields{
			"time":   time.Since(start),
			"source": "queue",
		}).Infof("%q -> %q", req.Text, out)

		return out
	}

	for _, t := range translators {
		source = t.Name()

		log.Debugf("Translating with %s", source)

		out, found, err = c.Get(source, req.Text)
		if found {
			source = source + "(cache)"

			// cached error
			if err != nil {
				log.Warnf("%s: %s", source, err)
				continue
			}

			// found translation with no errors
			break
		}

		if !t.Enabled() {
			continue
		}

		out, err = t.Translate(&req)
		if err != nil {
			log.Warnf("%s: %s", source, err)

			if err := c.Put(source, req.Text, out, err); err != nil {
				log.Warnf("%s: %s", source, err)
			}

			continue
		}

		if len(out) > 0 {
			err = c.Put(source, req.Text, out, nil)
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Errorf("Failed to save result to %s cache", source)
			}
		}

		break
	}

	// Notify waiting requests that we did the job
	q.Push(req, out)

	if len(out) < 1 {
		// TODO: Return original text or try to handle error in handler
		log.Errorf("All services failed to translate %q", req.Text)
	}

	log.WithFields(log.Fields{
		"time":   time.Since(start),
		"source": source,
	}).Infof("%q -> %q", req.Text, out)

	return out

}
