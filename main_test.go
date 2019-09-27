package main

import (
	"testing"
)

func TestUpdateMetric(t *testing.T) {
	tests := []string{
		"answer_to_everything;scope=universe;env=prod",
		"answer_to_everything;scope=universe_donot=panic;env=prod",
	}
	for _, test := range tests {
		updateMetric("testdata/" + test)
	}
}
