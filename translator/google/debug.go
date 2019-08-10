package google

import (
	"bufio"
	"os"

	log "github.com/sirupsen/logrus"
)

func dumpRequest(requestURL string, responseText string) {
	f, err := os.OpenFile("translation-errors.txt", os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(f)
	w.WriteString(requestURL)
	w.WriteRune('\n')
	w.WriteString(responseText)

	w.Flush()
}
