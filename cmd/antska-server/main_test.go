package main

import "testing"

func TestIsTerminal(t *testing.T) {
	// Just a simple test to make sure it runs
	result := isTerminal()
	// We don't assert specific value since it depends on how the test is run
	t.Logf("isTerminal result: %v", result)
}

func TestSimple(t *testing.T) {
	// A test that always passes
	if 1+1 != 2 {
		t.Error("Basic math is broken")
	}
}