package output

import (
	"fmt"
	"io"

	"github.com/shellrausch/gofuzzy/fuzz/client"
)

type txt struct {
	file io.Writer
}

func (t txt) init() {
	o := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", "Content-Length", "Words", "Lines", "Header", "Status-Code", "Payload")
	fmt.Fprintln(t.file, o)
}

func (t txt) write(r *client.Result) {
	o := fmt.Sprintf("%d\t\t\t\t%d\t\t%d\t\t%d\t\t%d\t\t\t%s", r.ContentLength, r.NumWords, r.NumLines, r.HeaderSize, r.StatusCode, r.Payload)
	fmt.Fprintln(t.file, o)
}

func (txt) writeProgress(p *client.Progress) {}
func (txt) close()                           {}
