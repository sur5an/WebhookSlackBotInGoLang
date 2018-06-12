package main

import "log"

type StringArray []string
var logFatal = log.Fatalf

func (s StringArray) Contains(str string) (found bool) {
	found = false
	for _, n := range s {
		if str == n {
			found = true
			return
		}
	}
	return
}

var failOnError = logAndCloseApp
func logAndCloseApp(err error, msg string) {
	if err != nil {
		logFatal("%s: %s", msg, err)
	}
}
