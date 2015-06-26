package log

import (
	_ "io/ioutil"
	"log"
	"os"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func init() {
	//traceHandle := ioutil.Discard
	traceHandle := os.Stdout
	infoHandle := os.Stdout
	warningHandle := os.Stdout
	errorHandle := os.Stderr

	Trace = log.New(traceHandle, "TRACE:   ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO:    ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR:   ", log.Ldate|log.Ltime|log.Lshortfile)
}
