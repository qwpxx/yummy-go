package mir

import "yummy-go.com/m/v2/span"

type Program struct {
	Declarations []Declaration
}

type VariableDeclaration interface {
	GetTypeView() TypeView
}

type DeclarationType uint

const (
	GlobalDeclarationType DeclarationType = iota
	FunctionDeclarationType
)

type Declaration interface {
	Type() DeclarationType
}

type GlobalDeclaration struct {
	Name     string
	TypeView TypeView
	Span     span.Span
}

func (s *GlobalDeclaration) GetTypeView() TypeView {
	return s.TypeView
}

func (s *GlobalDeclaration) Type() DeclarationType {
	return GlobalDeclarationType
}

type FunctionDeclaration struct {
	Name           string
	Arguments      map[string]Argument
	ReturnTypeView TypeView
	Body           Block
	ProcCode       string
	ArgumentIds    string
	Warp           bool
	Span           span.Span
}

type Argument struct {
	TypeView TypeView
	Span     span.Span
}

func (s *FunctionDeclaration) Type() DeclarationType {
	return FunctionDeclarationType
}

type Block struct {
	Statements []Statement
	Span       span.Span
}

type StatementType uint

const (
	DeclareStatementType StatementType = iota
	AssignStatementType
	ReturnStatementType
)

type Statement interface {
	Type() StatementType
}

type DeclareStatement struct {
	Name     string
	TypeView TypeView
	Span     span.Span
}

func (s *DeclareStatement) GetTypeView() TypeView {
	return s.TypeView
}

func (s *DeclareStatement) Type() StatementType {
	return DeclareStatementType
}

type AssignStatement struct {
	Declaration VariableDeclaration
	Value       Expression
	Span        span.Span
}

func (s *AssignStatement) Type() StatementType {
	return AssignStatementType
}

type ReturnStatement struct {
	Value Expression
	Span  span.Span
}

func (s *ReturnStatement) Type() StatementType {
	return ReturnStatementType
}

type ExpressionType uint

const (
	LiteralExpressionType ExpressionType = iota
	VariableExpressionType
	BinaryExpressionType
	UnaryExpressionType
)

type Expression interface {
	Type() ExpressionType
}

type LiteralExpression struct {
	Literal     any
	LiteralType Type
}

func (s *LiteralExpression) Type() ExpressionType {
	return LiteralExpressionType
}

type VariableExpression struct {
	Declaration VariableDeclaration
}

func (s *VariableExpression) Type() ExpressionType {
	return VariableExpressionType
}

type BinaryExpression struct {
	Lhs, Rhs   Expression
	Operator   OperatorType
	OutputType Type
}

func (s *BinaryExpression) Type() ExpressionType {
	return BinaryExpressionType
}

type UnaryExpression struct {
	Value      Expression
	Operator   OperatorType
	OutputType Type
}

func (s *UnaryExpression) Type() ExpressionType {
	return UnaryExpressionType
}
