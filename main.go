package main

import (
	"log"

	"github.com/shellrausch/gofuzzy/fuzz/client"
	"github.com/shellrausch/gofuzzy/fuzz/opts"
	"github.com/shellrausch/gofuzzy/fuzz/output"
)

func main() {
	opt := opts.New()
	if err := opt.Parse(output.SupportedFormats()); err != nil {
		log.Fatal(err)
	}
	out := output.New(opt.OutputFile, opt.OutputFormat)

	chans := client.New(opt)
	go client.Start(opt)

	for {
		select {
		case r := <-chans.Result:
			out.Write(r)
		case p := <-chans.Progress:
			go out.WriteProgress(p)
		case <-chans.Finish:
			out.Close()
			return
		}
	}
}
