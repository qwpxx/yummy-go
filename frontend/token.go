package frontend

import (
	"github.com/fatih/color"
	"yummy-go.com/m/v2/span"
)

type TokenType string

var TokenBroken = TokenType(color.HiRedString("(broken)"))

const (
	// Literals
	TokenLiteralNumber TokenType = "number literal"
	TokenLiteralTrue   TokenType = "true"
	TokenLiteralFalse  TokenType = "false"
	TokenLiteralString TokenType = "string literal"
	TokenIdentifier    TokenType = "idenifier"
	TokenRawIdentifier TokenType = "raw idenifier"
	// Parens
	TokenOpenParen    TokenType = "("
	TokenCloseParen   TokenType = ")"
	TokenOpenBracket  TokenType = "["
	TokenCloseBracket TokenType = "]"
	TokenOpenBrace    TokenType = "{"
	TokenCloseBrace   TokenType = "}"
	// Other Symbols
	TokenComma         TokenType = "[,]"
	TokenSemi          TokenType = "[;]"
	TokenColon         TokenType = "[:]"
	TokenAssign        TokenType = "[=]"
	TokenDeclareAssign TokenType = "[:=]"
	// Operators
	TokenOpAdd    TokenType = "operator [+]"
	TokenOpSub    TokenType = "operator [-]"
	TokenOpMul    TokenType = "operator [*]"
	TokenOpDiv    TokenType = "operator [/]"
	TokenOpEqu    TokenType = "operator [==]"
	TokenOpNeq    TokenType = "operator [!=]"
	TokenOpLes    TokenType = "operator [<]"
	TokenOpGes    TokenType = "operator [>]"
	TokenOpLte    TokenType = "operator [<=]"
	TokenOpGte    TokenType = "operator [>=]"
	TokenOpAnd    TokenType = "operator [||]"
	TokenOpOr     TokenType = "operator [&&]"
	TokenOpNot    TokenType = "operator [!]"
	TokenOpMember TokenType = "operator [.]"
	// Keywords
	TokenKeywordFor    TokenType = "keyword for"
	TokenKeywordVar    TokenType = "keyword var"
	TokenKeywordReturn TokenType = "keyword return"
	TokenKeywordIf     TokenType = "keyword if"
	TokenKeywordElse   TokenType = "keyword else"
	TokenKeywordTarget TokenType = "keyword target"
	TokenKeywordFunc   TokenType = "keyword func"
	TokenKeywordStruct TokenType = "keyword struct"
	// Types
	TokenTypeString TokenType = "type string"
	TokenTypeNumber TokenType = "type number"
	TokenTypeBool   TokenType = "type bool"
)

type Token struct {
	Type TokenType
	Span span.Span
}
