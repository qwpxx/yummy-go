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
	Arguments      []Argument
	ReturnTypeView TypeView
	Body           Block
	ProcCode       string
	ArgumentIds    string
	Warp           bool
	StackSize      uint
	Span           span.Span
}

type Argument struct {
	Name     string
	TypeView TypeView
	Span     span.Span
}

func (s *Argument) GetTypeView() TypeView {
	return s.TypeView
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
	Acessor Acessor
	Value   Expression
	Span    span.Span
}

func (s *AssignStatement) Type() StatementType {
	return AssignStatementType
}

type AcessorType uint

const (
	AcessorVariable AcessorType = iota
)

type Acessor interface {
	Type() AcessorType
}

type VariableAcessor struct {
	Declaration VariableDeclaration
	Span        span.Span
}

func (s *VariableAcessor) Type() AcessorType {
	return AcessorVariable
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
	AcessorExpressionType
	BinaryExpressionType
	UnaryExpressionType
	CallExpressionType
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

type AcessorExpression struct {
	Acessor Acessor
}

func (s *AcessorExpression) Type() ExpressionType {
	return AcessorExpressionType
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

type CallExpression struct {
	Function  *FunctionDeclaration
	Arguments []Expression
}

func (s *CallExpression) Type() ExpressionType {
	return CallExpressionType
}
