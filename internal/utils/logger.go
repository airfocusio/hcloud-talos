package utils

import (
	"log"
	"os"
)

type Logger struct {
	Debug *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
}

func NewLogger() Logger {
	return Logger{
		Debug: log.New(os.Stderr, "DEBUG: ", log.Ltime|log.Lshortfile),
		Info:  log.New(os.Stderr, "INFO:  ", log.Ltime|log.Lshortfile),
		Warn:  log.New(os.Stderr, "WARN:  ", log.Ltime|log.Lshortfile),
		Error: log.New(os.Stderr, "ERROR: ", log.Ltime|log.Lshortfile),
	}
}
