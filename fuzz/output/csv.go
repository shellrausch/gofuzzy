package output

import (
	"fmt"

	"github.com/shellrausch/gofuzzy/fuzz"
)

type csv struct{}

func (csv) init() {
	o := fmt.Sprintf("%s;%s;%s;%s;%s;%s", "Content-Length", "Words", "Lines", "Header", "Status-Code", "Result")
	fmt.Fprintln(outputFile, o)
}

func (csv) write(r *fuzz.Result) {
	o := fmt.Sprintf("%d;%d;%d;%d;%d;%s", r.ContentLength, r.NumWords, r.NumLines, r.HeaderSize, r.StatusCode, r.Result)
	fmt.Fprintln(outputFile, o)
}

func (csv) writeProgress(p *fuzz.Progress) {}
func (csv) close()                         {}
