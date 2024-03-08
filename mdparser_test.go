package kpless

import (
	"testing"
)

func TestOptRegexp(t *testing.T) {
	ls := []string{
		"* [title](#title_next)",
		"* 1d100 [title](#title_next)",
		"* 1d100 [title](#title_next) 1d100",
	}
	for _, l := range ls {
		if optRegexp.MatchString(l) == false {
			panic(l)
		}
	}
}
