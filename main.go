package main

import (
	"strings"

	"yummy-go.com/m/v2/span"
)

func main() {
	sourcePath := "main.yum"
	sourceCode := `target Stage

import "looks"

func main() {
    looks.Says(
        "Hello, World!"
    ) // Some comments
}`
	sourceLines := strings.Split(sourceCode, "\n")

	theSpan := span.Span{
		From: span.Position{
			Index:     18,
			LineIndex: 4,
			Lineno:    5,
		},
		To: span.Position{
			Index:     59,
			LineIndex: 5,
			Lineno:    7,
		},
		Source:      &sourceCode,
		SourceLines: &sourceLines,
		Path:        &sourcePath,
	}

	span.Report(theSpan, span.Error, "`looks.Says` not in scope")
	span.ReportNoSpan(span.Info, "did you mean `looks.Say`?")
}
