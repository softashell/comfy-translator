package google

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type responseStruct [][]string

var (
	nullRegex         = regexp.MustCompile(`(,null)+(,\d+)?`)
	otherGarbageRegex = regexp.MustCompile(`(?:(?:,\[null,".*])?,"[a-z]+"(?:,\[\[.*)?)(])$`)
)

func cleanJson(s string) string {
	s = nullRegex.ReplaceAllString(s, "")
	s = otherGarbageRegex.ReplaceAllString(s, "$1")

	// Strip first and last bracket
	if strings.HasSuffix(s, "]]]") {
		s = s[1 : len(s)-len("]")]
	} else {
		s = s[1:]
	}

	return s
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func cleanResponseText(s string) string {
	s = trimSuffix(s, "\n")

	return s
}

func decodeResponse(s string) (map[string]string, error) {
	resultMap := make(map[string]string)

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

		log.Debug("%q => %q\n", inputText, translatedText)

		resultMap[inputText] = translatedText
	}

	return resultMap, nil
}
