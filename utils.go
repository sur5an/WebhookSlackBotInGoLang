package main

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

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
