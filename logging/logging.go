package logging

import (
	"io/ioutil"
	"log"
)

// DebugLogger is a logger that's enabled conditionally when needed
// for debugging.
var DebugLogger *log.Logger

// DebugPrefix is the prefix used for debug output
const DebugPrefix = "DEBUG: "

func init() {
	DebugLogger = log.New(ioutil.Discard, DebugPrefix, log.LstdFlags)
}
