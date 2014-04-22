package core

import (
	"testing"
)

func TestContextInitializes(t *testing.T) {
	c := NewTestConfig(t)

	ctx, err := NewContext(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	if ctx.N == nil {
		t.Fatal("No notifier")
	}

	if ctx.DL == nil {
		t.Fatal("No distributed lock manager")
	}

	if ctx.DB == nil {
		t.Fatal("No database connection")
	}

	if c.Address() == "" {
		t.Fatal("HTTP Bind address incorrect")
	}
}

func TestContextCloseMultipleCallsOK(t *testing.T) {
	c := NewTestConfig(t)

	ctx, err := NewContext(c)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Close()

	ctx.Close()
	ctx.Close()
}
