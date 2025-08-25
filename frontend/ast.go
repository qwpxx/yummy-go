package frontend

import "yummy-go.com/m/v2/span"

type Program struct {
	Target       Token
	Declarations []Declaration
	Span         span.Span
}

type DeclarationType uint

const (
	GlobalVariableDeclarationType DeclarationType = iota
	FunctionDeclarationType
)

type Declaration interface {
	Display
	Type() DeclarationType
}

type FunctionDeclaration struct {
	Name Token
	Body Block
	Span span.Span
}

func (s *FunctionDeclaration) Type() DeclarationType {
	return FunctionDeclarationType
}

type Block struct {
	Statements []Statement
	Span       span.Span
}

type StatementType uint

const ()

type Statement interface {
	Display
	Type() StatementType
}
