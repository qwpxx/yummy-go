package main

import (
	"yummy-go.com/m/v2/frontend"
	"yummy-go.com/m/v2/span"
)

func main() {
	sourcePath := "main.yum"
	sourceCode := `target Stage

func main() {}`
	lexer := frontend.NewLexer(sourcePath, sourceCode)
	parser := frontend.NewParser(lexer)

	span.ResetStats()
	ast, err := parser.ParseProgram()
	if err != nil {
		errorCount := span.GetStats(span.Error)
		span.ReportNoSpan(span.Error, "%s: %s generated", sourcePath, span.Pluralize(errorCount, "error", "errors"))
		return
	}
	ast.Display(0)
}
