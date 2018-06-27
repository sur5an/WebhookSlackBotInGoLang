package utils

import "log"

type StringArray []string

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


var LogFatal = log.Fatalf

func logAndCloseApp(err error, msg string) {
	if err != nil {
		LogFatal("%s: %s", msg, err)
	}
}

var FailOnError = logAndCloseApp