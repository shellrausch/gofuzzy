package output

import "github.com/shellrausch/gofuzzy/fuzz/client"

type null struct{}

func (null) init()                            {}
func (null) write(r *client.Result)           {}
func (null) writeProgress(p *client.Progress) {}
func (null) close()                           {}
