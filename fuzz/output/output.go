package output

import (
	"io"
	"os"

	"github.com/shellrausch/gofuzzy/fuzz"
)

var cli Writer
var outputFile io.Writer
var supportedFormats map[string]bool

func init() {
	supportedFormats = map[string]bool{"csv": true, "txt": true, "json": true}
}

// Writer must be implemented by every output format which wants to write.
type Writer interface {
	init()
	write(*fuzz.Result)
	writeProgress(*fuzz.Progress)
	close()
}

// Formats returns all available and supported output formats to which gofuzzy can write to.
func Formats() map[string]bool {
	return supportedFormats
}

// SetOutput sets the output file and decides on which output media
// the results should be shown. We always output on the CLI, also if another
// output media is provided.
func SetOutput(filename, outputFormat string) Writer {
	outputFile, _ = os.Create(filename)
	// We want always a CLI output.
	cli = tabCli{}
	cli.init()

	var ow Writer

	switch outputFormat {
	case "csv":
		ow = csv{}
	case "txt":
		ow = txt{}
	case "json":
		ow = json{}
	default:
		ow = null{}
	}

	ow.init()

	return ow
}

// Write just writes the result to the defined output and additionaly to the CLI.
func Write(ow Writer, r *fuzz.Result) {
	ow.write(r)
	cli.write(r)
}

// WriteProgress just writes the progress to the defined output and additionaly to the CLI.
func WriteProgress(ow Writer, pr *fuzz.Progress) {
	ow.writeProgress(pr)
	cli.writeProgress(pr)
}

// Close closes output writer.
func Close(ow Writer) {
	ow.close()
}
