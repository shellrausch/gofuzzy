package output

import (
	// Avoids naming collision with our struct name. Yes it's dirty.
	jsonn "encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/shellrausch/gofuzzy/fuzz"
)

type json struct{}

var jsonResults []*fuzz.Result

func (json) write(r *fuzz.Result) {
	jsonResults = append(jsonResults, r)
}

func (json) close() {
	json, err := jsonn.Marshal(jsonResults)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(outputFile, string(json))
}

func (json) init()                          {}
func (json) writeProgress(p *fuzz.Progress) {}
