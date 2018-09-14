package output

import (
	"os"

	"github.com/shellrausch/gofuzzy/fuzz/client"
)

// Output contains the output writers.
type Output struct {
	cliWriter  FuzzWriter
	fileWriter FuzzWriter
}

// FuzzWriter must be implemented by every output format which wants to write.
type FuzzWriter interface {
	init()
	write(*client.Result)
	writeProgress(*client.Progress)
	close()
}

// SupportedFormats returns all available and supported output
// formats to which gofuzzy can write to.
func SupportedFormats() map[string]bool {
	return map[string]bool{"csv": true, "txt": true, "json": true}
}

// New sets the output file and decides on which output media
// the results should be shown. We always output on the CLI, also if another
// output media is provided.
func New(filename, outputFormat string) *Output {
	f, _ := os.Create(filename)

	o := &Output{}
	switch outputFormat {
	case "csv":
		o.fileWriter = csv{file: f}
	case "txt":
		o.fileWriter = txt{file: f}
	case "json":
		o.fileWriter = json{file: f}
	default:
		o.fileWriter = null{}
	}
	o.fileWriter.init()

	// We write always to the CLI.
	o.cliWriter = cli{}
	o.cliWriter.init()

	return o
}

// Write writes the result to the defined output and additionaly to the CLI.
func (o *Output) Write(r *client.Result) {
	o.fileWriter.write(r)
	o.cliWriter.write(r)
}

// WriteProgress writes the progress to the defined output and additionaly to the CLI.
func (o *Output) WriteProgress(pr *client.Progress) {
	o.fileWriter.writeProgress(pr)
	o.cliWriter.writeProgress(pr)
}

// Close closes output writer.
func (o *Output) Close() {
	o.fileWriter.close()
}
