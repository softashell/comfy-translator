package google

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
)

func mergeOutput(input []inputObject, output []responsePair) []responsePair {
	inputCount := len(input)
	outputCount := len(output)

	if inputCount == 0 || outputCount == 0 {
		return output
	}

	if inputCount > outputCount {
		log.Fatal("Truncated output", spew.Sdump(input), spew.Sdump(output))
	}

	for i := 0; i < outputCount-1; i++ {
		in := input[i].req.Text
		out := output[i].input

		if strings.TrimSpace(in) == strings.TrimSpace(out) {
			continue
		}

		if len(in) < len(out) {
			log.Fatalf("original text is smaller than output! %q %q", in, out)
		}

		// TODO: Loop and handle more than one item join
		next := i + 1
		if next > outputCount-1 {
			log.Fatal("output exhausted, unable to get proper results")
		}

		nextOut := output[next].input

		// Only check next input if it exists
		if next < inputCount-1 {
			nextIn := input[next].req.Text

			if nextIn == nextOut {
				log.Fatalf("output has truncated input string\n%q == %q\n%s\n%s", nextIn, nextOut, spew.Sdump(input), spew.Sdump(output))
			}
		}

		out += nextOut

		// Update current record
		output[i].input = out
		output[i].output += output[next].output

		// Delete next item in output
		if next == outputCount-1 {
			// Truncate
			output = output[:i+1]
		} else {
			// Cut
			output = append(output[:next], output[next+1:]...)
		}

		outputCount--

		// Exit if we have balanced items
		if inputCount == outputCount {
			break
		}
	}

	return output
}
