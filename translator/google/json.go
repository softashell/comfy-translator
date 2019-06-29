package google

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
)

type responseStruct [][]string

type responsePair struct {
	input  string
	output string
}

var (
	nullRegex         = regexp.MustCompile(`(,null)+(,\d+)?`)
	otherGarbageRegex = regexp.MustCompile(`(?:(?:,\[null,".*])?,"[a-z]+"(?:,\[\[.*)?)(])$`)
	translationSource = regexp.MustCompile(`,\[{3}"[a-z\d]+","[a-z]+_[a-z]+_\d{4}[a-z\d]+\.md"\]{3}`)
	leftowers         = regexp.MustCompile(`,"[a-z]{2,4}"]?"?$`)
)

func cleanJson(s string) string {
	s = nullRegex.ReplaceAllString(s, "")
	s = translationSource.ReplaceAllString(s, "")
	s = otherGarbageRegex.ReplaceAllString(s, "$1")
	s = leftowers.ReplaceAllString(s, "")

	// Strip first and last bracket
	if strings.HasSuffix(s, "]]]") {
		s = s[1 : len(s)-len("]")]
	} else {
		s = s[1:]
	}

	return s
}

func cleanResponseText(s string) string {
	// Remove trailing whitespace
	s = strings.TrimRightFunc(s, unicode.IsSpace)

	return s
}

func decodeResponse(s string) ([]responsePair, error) {
	var out []responsePair

	s = cleanJson(s)
	log.Debug(s)

	var resp responseStruct
	err := json.Unmarshal([]byte(s), &resp)
	if err != nil {
		return nil, err
	}

	for _, result := range resp {
		if len(result) != 2 {
			return nil, fmt.Errorf("returned incorrect result pair: %v", result)
		}

		translatedText := cleanResponseText(result[0])
		inputText := cleanResponseText(result[1])

		log.Debugf("%q => %q\n", inputText, translatedText)

		out = append(out, responsePair{
			input:  inputText,
			output: translatedText,
		})
	}

	return out, nil
}
