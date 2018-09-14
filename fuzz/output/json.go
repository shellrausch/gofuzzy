package output

import (
	jsn "encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/shellrausch/gofuzzy/fuzz/client"
)

type json struct {
	file io.Writer
}

var jsonResults []*client.Result

func (json) write(r *client.Result) {
	jsonResults = append(jsonResults, r)
}

func (j json) close() {
	json, err := jsn.Marshal(jsonResults)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(j.file, string(json))
}

func (json) init()                            {}
func (json) writeProgress(p *client.Progress) {}
