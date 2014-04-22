package core

import (
	"testing"
)

// NewTestConfig provides the Config or fails the test with an error.
func NewTestConfig(t *testing.T) Config {
	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	return c
}
