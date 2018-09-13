package output

import (
	"fmt"

	"github.com/shellrausch/gofuzzy/fuzz"
)

type txt struct{}

func (txt) init() {
	o := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", "Content-Length", "Words", "Lines", "Header", "Status-Code", "Result")
	fmt.Fprintln(outputFile, o)
}

func (txt) write(r *fuzz.Result) {
	o := fmt.Sprintf("%d\t\t\t\t%d\t\t%d\t\t%d\t\t%d\t\t\t%s", r.ContentLength, r.NumWords, r.NumLines, r.HeaderSize, r.StatusCode, r.Result)
	fmt.Fprintln(outputFile, o)
}

func (txt) writeProgress(p *fuzz.Progress) {}
func (txt) close()                         {}
