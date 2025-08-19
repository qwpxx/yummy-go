package span

import (
	"fmt"

	"github.com/fatih/color"
)

const ReportContextLineToPrint = 1

type ReportLevel string

const (
	Error ReportLevel = "error"
	Warn  ReportLevel = "warn"
	Info  ReportLevel = "info"
)

var (
	colorErr    = color.New(color.FgHiRed, color.Bold)
	colorWarn   = color.New(color.FgHiYellow, color.Bold)
	colorInfo   = color.New(color.FgHiBlue, color.Bold)
	colorIgnore = color.New(color.FgHiBlack)
	colorTip    = color.New(color.FgHiBlue, color.Bold)
	colorMark   = color.New(color.FgHiMagenta, color.Bold, color.Italic)
	colorCode   = color.New(color.Bold)
)

var reportLevelColorMap = map[ReportLevel]*color.Color{
	Error: colorErr,
	Warn:  colorWarn,
	Info:  colorInfo,
}

func ReportNoSpan(level ReportLevel, message string) {
	reportLevelColorMap[level].Printf("%s", level)
	fmt.Printf(": %s\n", message)
}

func Report(span Span, level ReportLevel, message string) {
	reportLevelColorMap[level].Printf("%s", level)
	fmt.Printf(": %s\n  -> ", message)
	if span.Path != nil {
		colorTip.Printf("%s", *span.Path)
	} else {
		colorIgnore.Printf("(no path)")
	}
	fmt.Printf(" [")
	colorTip.Printf("%d", span.From.Lineno+1)
	fmt.Printf(":")
	colorTip.Printf("%d", span.From.LineIndex+1)
	fmt.Printf("]\n")

	lineStart := max(int(span.From.Lineno)-ReportContextLineToPrint, 0)
	lineEnd := min(int(span.To.Lineno)+ReportContextLineToPrint, len(*span.SourceLines)-1)

	if lineStart > 0 {
		colorIgnore.Printf("...\n")
	}
	for line := lineStart; line <= lineEnd; line += 1 {
		if line < int(span.From.Lineno) || line > int(span.To.Lineno) {
			colorIgnore.Printf(" %-4d ", line+1)
		} else {
			colorMark.Printf(" %-4d ", line+1)
		}
		lineContent := (*span.SourceLines)[line]
		if line < int(span.From.Lineno) || line > int(span.To.Lineno) {
			colorCode.Printf("%s", lineContent)
		} else if line == int(span.From.Lineno) && line == int(span.To.Lineno) {
			colorCode.Printf("%s", lineContent[0:span.From.LineIndex])
			colorMark.Printf("%s", lineContent[span.From.LineIndex:span.To.LineIndex])
			colorCode.Printf("%s", lineContent[span.To.LineIndex:])
		} else if line == int(span.From.Lineno) {
			colorCode.Printf("%s", lineContent[0:span.From.LineIndex])
			colorMark.Printf("%s", lineContent[span.From.LineIndex:])
		} else if line == int(span.To.Lineno) {
			colorMark.Printf("%s", lineContent[0:span.To.LineIndex])
			colorCode.Printf("%s", lineContent[span.To.LineIndex:])
		} else {
			colorMark.Printf("%s", (*span.SourceLines)[line])
		}
		fmt.Println()
	}
	if lineEnd < len(*span.SourceLines)-1 {
		colorIgnore.Printf("...\n")
	}
}
