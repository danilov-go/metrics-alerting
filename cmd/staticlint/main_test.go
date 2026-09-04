package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOsExitAnalyzer_InMemory(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), osExitCheckAnalyzer, "./...")
}
