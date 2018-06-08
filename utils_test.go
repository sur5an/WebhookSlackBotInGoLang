package main

import (
    "testing"
)

func TestContains (t *testing.T) {
    strArray := StringArray{"slack", "bot", "webhook", "rabbit", "sur5an"}
    inputExpectedOutput := []struct {
        input          string
        expectedOutput bool
    }{
        {"bot", true},
        {"test", false},
        {"rabbit", true},
        {"web", false},
        {"webhoo", false},
    }
    for _, element := range inputExpectedOutput {
        if strArray.Contains(element.input) != element.expectedOutput {
            t.Errorf("Contains failed for %s unexpected output %t came", element.input, element.expectedOutput)
        }
    }

}
