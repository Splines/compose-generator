/*
Copyright © 2021 Compose Generator Contributors
All rights reserved.
*/

package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ------------------------------------------------- P -------------------------------------------------

func TestP(t *testing.T) {
	// Test data

}

// ------------------------------------------------- Pl ------------------------------------------------

func TestPl(t *testing.T) {
	// Test data
	text := "This is a test."
	// Mock functions
	whiteCallCount := 0
	white = func(format string, a ...interface{}) {
		whiteCallCount++
		assert.Equal(t, text, format)
	}
	// Execute test
	Pl(text)
	// Assert
	assert.Equal(t, 1, whiteCallCount)
}

// ------------------------------------------------ Pel ------------------------------------------------

func TestPel1(t *testing.T) {
	// Mock functions
	printlnCallCount := 0
	println = func(a ...interface{}) (n int, err error) {
		printlnCallCount++
		assert.Zero(t, len(a))
		return 0, nil
	}
	logError = func(message string, exit bool) {
		assert.Fail(t, "Unexpected call of logError")
	}
	// Execute test
	Pel()
	// Assert
	assert.Equal(t, 1, printlnCallCount)
}

func TestPel2(t *testing.T) {
	// Mock functions
	printlnCallCount := 0
	println = func(a ...interface{}) (n int, err error) {
		printlnCallCount++
		assert.Zero(t, len(a))
		return 0, errors.New("Test")
	}
	logError = func(message string, exit bool) {
		assert.Equal(t, "Could not print empty line", message)
		assert.True(t, exit)
	}
	// Execute test
	Pel()
	// Assert
	assert.Equal(t, 1, printlnCallCount)
}
