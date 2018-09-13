package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/shellrausch/gofuzzy/fuzz"
	"github.com/shellrausch/gofuzzy/fuzz/output"
)

func main() {
	opts, err := fuzz.Parse(output.Formats())
	if err != nil {
		log.Fatal(err)
	}

	writer := output.SetOutput(opts.OutputFile, opts.OutputFormat)

	chans := fuzz.NewFuzz(opts)
	go fuzz.Start(opts)

	for {
		select {
		case r := <-chans.Result:
			output.Write(writer, r)
		case p := <-chans.Progress:
			go output.WriteProgress(writer, p)
		case <-chans.Finished:
			output.Close(writer)
			return
		}
	}
}
