package frontend

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"yummy-go.com/m/v2/span"
)

var (
	colorSpan  = color.New(color.FgHiBlue, color.Bold)
	colorKey   = color.New(color.FgHiYellow, color.Bold)
	colorTitle = color.New(color.FgHiMagenta, color.Bold)
	colorIndex = color.New(color.FgHiGreen, color.Bold)
	colorToken = color.New(color.FgHiCyan, color.Italic, color.Bold)
)

type Display interface {
	Display(indent uint)
}

func displayIndent(indent uint) {
	fmt.Print(strings.Repeat(" ", int(indent)))
}

func (s Token) Display(indent uint) {
	colorToken.Printf("[%s](%s)", s.Type, s.Span.String())
	fmt.Print(" <- ")
	displaySpan(s.Span)
	fmt.Println()
}

func displaySpan(span span.Span) {
	colorSpan.Printf("line [%d:%d], [%d:%d]", span.From.Lineno+1, span.To.Lineno+1, span.From.LineIndex, span.To.LineIndex)
}

func displayTitle(title string, titleSpan span.Span) {
	colorTitle.Print(title)
	fmt.Print(" <- ")
	displaySpan(titleSpan)
	fmt.Println()
}

func displayKV[T Display](indent uint, key string, v T) {
	displayIndent(indent)
	colorKey.Print(key)
	fmt.Print(": ")
	v.Display(indent + 1)
}

func displayKVList[T Display](indent uint, key string, vlist []T) {
	displayIndent(indent)
	colorKey.Print(key)
	fmt.Print(": ")
	colorIndex.Printf("Array[%d]\n", len(vlist))
	for idx, v := range vlist {
		displayIndent(indent + 1)
		fmt.Print("[")
		colorIndex.Print(idx)
		fmt.Print("]: ")
		v.Display(indent + 1)
	}
}

func (s Program) Display(indent uint) {
	displayTitle("Program", s.Span)
	displayKV(indent+1, "target", s.Target)
	displayKVList(indent+1, "declarations", s.Declarations)
}

func (s FunctionDeclaration) Display(indent uint) {
	displayTitle("FunctionDeclaration", s.Span)
	displayKV(indent+1, "name", s.Name)
	displayKV(indent+1, "body", s.Body)
}

func (s Block) Display(indent uint) {
	displayTitle("FunctionDeclaration", s.Span)
	displayKVList(indent+1, "statements", s.Statements)
}
