package main

import (
	"testing"

	. "github.com/bborbe/assert"
)

func TestResumeFail(t *testing.T) {
	if err := AssertThat(DEFAULT_PORT, Is(DEFAULT_PORT)); err != nil {
		t.Fatal(err)
	}
}
