package log

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info ]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
)

var (
	Errorf  = errorLog.Printf
	Errorln = errorLog.Println
	Infof   = infoLog.Printf
	Infoln  = infoLog.Println
)

const (
	errorLevel = iota
	infoLevel
	disable
)

func SetLevel(level int) {
	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if level > errorLevel {
		errorLog.SetOutput(ioutil.Discard)
	}

	if level > infoLevel {
		infoLog.SetOutput(ioutil.Discard)
	}
}
