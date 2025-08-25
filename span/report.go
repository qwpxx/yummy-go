package span

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const ReportContextLineToPrint = 2

type ReportLevel string

const (
	Error ReportLevel = "error"
	Warn  ReportLevel = "warn"
	Info  ReportLevel = "info"
)

var (
	colorErr    = color.New(color.FgHiRed, color.Bold)
	colorWarn   = color.New(color.FgHiYellow, color.Bold)
	colorInfo   = color.New(color.FgHiCyan, color.Bold)
	colorIgnore = color.New(color.FgHiBlack)
	colorTip    = color.New(color.FgHiBlue, color.Bold)
	colorMark   = color.New(color.FgHiYellow, color.Bold)
	colorCode   = color.New()
)

var reportLevelColorMap = map[ReportLevel]*color.Color{
	Error: colorErr,
	Warn:  colorWarn,
	Info:  colorInfo,
}

var Stats map[ReportLevel]uint = map[ReportLevel]uint{
	Error: 0,
	Warn:  0,
	Info:  0,
}

func ResetStats() {
	Stats[Error] = 0
	Stats[Warn] = 0
	Stats[Info] = 0
}

func GetStats(level ReportLevel) uint {
	return Stats[level]
}

func ReportNoSpan(level ReportLevel, message string, args ...any) error {
	Stats[level] += 1
	reportLevelColorMap[level].Printf("%s", level)
	fmt.Printf(": ")
	fmt.Printf(message, args...)
	fmt.Printf("\n")
	return fmt.Errorf("%s: %s", level, message)
}

func Report(span Span, level ReportLevel, message string, args ...any) error {
	Stats[level] += 1
	reportLevelColorMap[level].Printf("%s", level)
	fmt.Printf(": ")
	fmt.Printf(message, args...)
	fmt.Printf("\n  -> ")
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

	emittedEsp := false
	if lineStart > 0 {
		colorIgnore.Printf("...\n")
		emittedEsp = true
	}
	for line := lineStart; line <= lineEnd; line += 1 {
		lineContent := (*span.SourceLines)[line]
		if strings.TrimSpace(lineContent) == "" {
			if !emittedEsp {
				colorIgnore.Printf("...\n")
			}
			emittedEsp = true
			continue
		}
		emittedEsp = false
		if line < int(span.From.Lineno) || line > int(span.To.Lineno) {
			colorIgnore.Printf(" %-4d ", line+1)
		} else {
			colorIgnore.Printf(" %-4d ", line+1)
		}
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
			colorMark.Printf("%s", lineContent)
		}
		fmt.Println()
	}
	if !emittedEsp && lineEnd < len(*span.SourceLines)-1 {
		colorIgnore.Printf("...\n")
	}
	return fmt.Errorf("%s: %s", level, message)
}

func Pluralize(n uint, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}
