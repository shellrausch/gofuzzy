package output

import (
	"github.com/shellrausch/gofuzzy/fuzz"
)

type null struct{}

func (null) init()                          {}
func (null) write(r *fuzz.Result)           {}
func (null) writeProgress(p *fuzz.Progress) {}
func (null) close()                         {}
