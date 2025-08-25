package frontend

import (
	"strings"

	"yummy-go.com/m/v2/span"
)

type Lexer struct {
	source      string
	sourceLines []string
	path        string
	current     lexerState
	mark        lexerState
}

type lexerState struct {
	Lineno    uint
	Index     uint
	LineIndex uint
}

func (s *Lexer) span() span.Span {
	return span.Span{
		From: span.Position{
			Index:     s.mark.Index,
			LineIndex: s.mark.LineIndex,
			Lineno:    s.mark.Lineno,
		},
		To: span.Position{
			Index:     s.current.Index,
			LineIndex: s.current.LineIndex,
			Lineno:    s.current.Lineno,
		},
		Source:      &s.source,
		SourceLines: &s.sourceLines,
		Path:        &s.path,
	}
}

func (s *Lexer) token(tokenType TokenType) *Token {
	return &Token{
		Type: tokenType,
		Span: s.span(),
	}
}

func (s *Lexer) consume() byte {
	if s.current.Index >= uint(len(s.source)) {
		return '\u0000'
	}
	if s.source[s.current.Index] == '\n' {
		s.current.Lineno += 1
		s.current.LineIndex = 0
	} else {
		s.current.LineIndex += 1
	}
	result := s.source[s.current.Index]
	s.current.Index += 1
	return result
}

func (s *Lexer) peek() byte {
	if s.current.Index >= uint(len(s.source)) {
		return '\u0000'
	}
	return s.source[s.current.Index]
}

func (s *Lexer) setMark() {
	s.mark = s.current
}

func isNumberic(char byte) bool {
	return char >= '0' && char <= '9'
}

func isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func isValidIdentifierFollowing(char byte) bool {
	return isNumberic(char) || isAlpha(char) || char == '_'
}

func isValidIdentifierLeading(char byte) bool {
	return isAlpha(char) || char == '_'
}

func (s *Lexer) NextToken() *Token {
	s.setMark()
	current := s.consume()
	for current == '\n' || current == ' ' || current == '\t' {
		s.setMark()
		current = s.consume()
	}
	switch current {
	case '\u0000':
		return nil // EOF
	case '(':
		return s.token(TokenOpenParen)
	case '[':
		return s.token(TokenOpenBracket)
	case '{':
		return s.token(TokenOpenBrace)
	case ')':
		return s.token(TokenCloseParen)
	case ']':
		return s.token(TokenCloseBracket)
	case '}':
		return s.token(TokenCloseBrace)
	case '.':
		return s.token(TokenOpMember)
	case ',':
		return s.token(TokenComma)
	case ';':
		return s.token(TokenSemi)
	case '+':
		return s.token(TokenOpAdd)
	case '*':
		return s.token(TokenOpMul)
	case '/':
		return s.token(TokenOpDiv)
	case ':':
		if s.peek() == '=' {
			s.consume()
			return s.token(TokenDeclareAssign)
		}
		return s.token(TokenColon)
	case '=':
		if s.peek() == '=' {
			s.consume()
			return s.token(TokenOpEqu)
		}
		return s.token(TokenAssign)
	case '>':
		if s.peek() == '=' {
			s.consume()
			return s.token(TokenOpGte)
		}
		return s.token(TokenOpGes)
	case '<':
		if s.peek() == '=' {
			s.consume()
			return s.token(TokenOpLte)
		}
		return s.token(TokenOpLes)
	case '!':
		if s.peek() == '=' {
			s.consume()
			return s.token(TokenOpNeq)
		}
		return s.token(TokenOpNot)
	case '&':
		if s.peek() == '&' {
			s.consume()
			return s.token(TokenOpAnd)
		}
		span.Report(s.span(), span.Error, "operator bit-wise and [&] is not allowed")
		return s.token(TokenBroken)
	case '|':
		if s.peek() == '|' {
			s.consume()
			return s.token(TokenOpOr)
		}
		span.Report(s.span(), span.Error, "operator bit-wise or [|] is not allowed")
		return s.token(TokenBroken)
	case '"':
	outer1:
		for {
			current := s.consume()
			switch current {
			case '\\':
				s.consume()
			case '"':
				break outer1
			}
		}
		return s.token(TokenLiteralString)
	case '#':
		if s.consume() != '"' {
			return s.token(TokenBroken)
		}
	outer2:
		for {
			current := s.consume()
			switch current {
			case '\\':
				s.consume()
			case '"':
				break outer2
			}
		}
		return s.token(TokenRawIdentifier)
	case '-':
		if isNumberic(s.peek()) {
			current = s.consume()
		} else {
			return s.token(TokenOpSub)
		}
		fallthrough
	default:
		if isNumberic(current) {
			for isNumberic(s.peek()) {
				s.consume()
				if s.peek() == '.' {
					s.consume()
				}
			}
			return s.token(TokenLiteralNumber)
		} else if isValidIdentifierLeading(current) {
			for isValidIdentifierFollowing(s.peek()) {
				s.consume()
			}
			token := s.token(TokenIdentifier)
			switch token.Span.String() {
			case "func":
				return s.token(TokenKeywordFunc)
			case "for":
				return s.token(TokenKeywordFor)
			case "if":
				return s.token(TokenKeywordIf)
			case "else":
				return s.token(TokenKeywordElse)
			case "return":
				return s.token(TokenKeywordReturn)
			case "var":
				return s.token(TokenKeywordVar)
			case "target":
				return s.token(TokenKeywordTarget)
			case "number":
				return s.token(TokenTypeNumber)
			case "string":
				return s.token(TokenTypeString)
			case "bool":
				return s.token(TokenTypeBool)
			case "struct":
				return s.token(TokenKeywordStruct)
			case "true":
				return s.token(TokenLiteralTrue)
			case "false":
				return s.token(TokenLiteralFalse)
			}
			return token
		}
	}
	span.Report(s.span(), span.Error, "unexpected char %c", current)
	return nil // Error
}

func NewLexer(path string, source string) Lexer {
	sourceLines := strings.Split(source, "\n")
	return Lexer{
		path:        path,
		source:      source,
		sourceLines: sourceLines,
	}
}
