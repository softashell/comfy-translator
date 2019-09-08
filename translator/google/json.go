package google

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Jeffail/gabs/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type responsePair struct {
	input  string
	output string
}

func cleanResponseText(s string) string {
	// Remove trailing whitespace
	s = strings.TrimRightFunc(s, unicode.IsSpace)

	return s
}

func decodeResponse(s string) ([]responsePair, error) {
	var out []responsePair

	jsonParsed, err := gabs.ParseJSON([]byte(s))
	if err != nil {
		err = errors.Wrap(err, "failed to parse json")
		return nil, err
	}

	for _, child := range jsonParsed.Index(0).Children() {
		if len(child.Children()) < 2 {
			return nil, fmt.Errorf("returned invalid object: %s", string(child.EncodeJSON()))
		}

		if child.Index(0).Data() == nil || child.Index(0).Data() == nil {
			continue
		}

		translatedText := cleanResponseText(child.Index(0).Data().(string))
		inputText := cleanResponseText(child.Index(1).Data().(string))

		log.Debugf("%q => %q\n", inputText, translatedText)

		out = append(out, responsePair{
			input:  inputText,
			output: translatedText,
		})
	}

	return out, nil
}
