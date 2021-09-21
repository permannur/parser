package logger

import (
	"fmt"
	"io"
	"parser/config"
)

var debug bool
var wr io.Writer

func Init(injectedWriter io.Writer) {
	debug = config.Values().GetDebug()
	wr = injectedWriter
}

func Write(s string) {
	if debug {
		fmt.Fprintf(wr, "%s", s)
	}
}
