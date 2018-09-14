package output

import (
	"fmt"
	"io"

	"github.com/shellrausch/gofuzzy/fuzz/client"
)

type csv struct {
	file io.Writer
}

func (c csv) init() {
	o := fmt.Sprintf("%s;%s;%s;%s;%s;%s", "Content-Length", "Words", "Lines", "Header", "Status-Code", "Payload")
	fmt.Fprintln(c.file, o)
}

func (c csv) write(r *client.Result) {
	o := fmt.Sprintf("%d;%d;%d;%d;%d;%s", r.ContentLength, r.NumWords, r.NumLines, r.HeaderSize, r.StatusCode, r.Payload)
	fmt.Fprintln(c.file, o)
}

func (csv) writeProgress(p *client.Progress) {}
func (csv) close()                           {}
