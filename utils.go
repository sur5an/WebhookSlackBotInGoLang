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

func failOnError(err error, msg string) {
	if err != nil {
		logFatal("%s: %s", msg, err)
	}
}
