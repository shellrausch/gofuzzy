package output

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/shellrausch/gofuzzy/fuzz/client"
)

type cli struct{}

var tableWriter *tabwriter.Writer

func (cli) init() {
	fmt.Println(banner)
	tableWriter = new(tabwriter.Writer)
	tableWriter.Init(os.Stdout, 13, 0, 0, ' ', 0)

	fmt.Fprintln(tableWriter, "---------------------------------------------------------------------------------")
	fmt.Fprintln(tableWriter, "Chars(-hh) \t Words(-hw) \t Lines(-hl) \t Header(-hr) \t Code(-hc) \t Payload")
	fmt.Fprintln(tableWriter, "---------------------------------------------------------------------------------")
}

func (cli) write(r *client.Result) {
	o := fmt.Sprintf("%d \t %d \t %d \t %d \t %d \t %s", r.ContentLength, r.NumWords, r.NumLines, r.HeaderSize, r.StatusCode, r.Payload)
	fmt.Fprintln(tableWriter, o)
	tableWriter.Flush()
}

func (cli) writeProgress(p *client.Progress) {
	percent := int((float64(p.NumDoneRequests) / float64(p.NumApproxRequests)) * 100)
	fmt.Printf("\r%30s\r~%d/%d (%d%%)\r", "", p.NumDoneRequests, p.NumApproxRequests, percent) // Output: ~123/9000 (2%)
}

func (cli) close() {
	// Just clear the last progress output with some whitespaces.
	fmt.Printf("\r%30s\r", "")
}

var banner = `                                             
   _ __________ + _________     *
_ _ __ /  ________/ ____________________ ___   *
    __/  /  / _  /  __/  / /__ /___  /  /  / -+
   +  \____/\___/__/  \___/_____/____\_   /     
           *          - -+          /____/        *
`
