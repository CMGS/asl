package cmpor

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "a")
}

func TestAnalyzerSigned(t *testing.T) {
	signed = true
	defer func() { signed = false }()
	analysistest.Run(t, analysistest.TestData(), Analyzer, "b")
}
