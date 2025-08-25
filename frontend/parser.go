package frontend

import "yummy-go.com/m/v2/span"

type Parser struct {
	lexer       Lexer
	peekedToken *Token
}

func NewParser(lexer Lexer) Parser {
	return Parser{
		lexer: lexer,
	}
}

func (s *Parser) peek() *Token {
	if s.peekedToken != nil {
		return s.peekedToken
	}
	token := s.lexer.NextToken()
	s.peekedToken = token
	return token
}

func (s *Parser) consume() *Token {
	if s.peekedToken != nil {
		token := s.peekedToken
		s.peekedToken = nil
		return token
	}
	return s.lexer.NextToken()
}

func (s *Parser) expect(types ...TokenType) (*Token, bool) {
	nextToken := s.peek()
	if nextToken == nil {
		return nil, false
	}
	for _, tokenType := range types {
		if nextToken.Type == tokenType {
			s.consume()
			return nextToken, true
		}
	}
	return nextToken, false
}

func (s *Parser) reportToken(token *Token, level span.ReportLevel, message string, args ...any) error {
	if token == nil {
		return span.ReportNoSpan(level, s.lexer.path+": "+message, args...)
	}
	return span.Report(token.Span, level, message, args...)
}

func formatExpectList(items []TokenType) string {
	if len(items) == 0 {
		return "nothing"
	}
	if len(items) == 1 {
		return string(items[0])
	}
	result := string(items[0])
	for i := 1; i < len(items)-1; i += 1 {
		result = result + ", " + string(items[i])
	}
	result = result + " or " + string(items[len(items)-1])
	return result
}

func (s *Parser) reportExpectToken(token *Token, expects ...TokenType) error {
	if token == nil {
		return span.ReportNoSpan(span.Error, "%s: expected %s, found EOF", s.lexer.path, formatExpectList(expects))
	}
	return span.Report(token.Span, span.Error, "expected %s, found %s", formatExpectList(expects), token.Type)
}

func (s *Parser) RestoreFromError() {
	for {
		token := s.peek()
		if token == nil {
			return
		}
		switch token.Type {
		case TokenKeywordFunc:
			return
		default:
			s.consume()
		}
	}
}

func (s *Parser) ParseProgram() (Program, error) {
	tokenTarget, ok := s.expect(TokenKeywordTarget)
	if !ok {
		return Program{}, s.reportToken(tokenTarget, span.Error, "missing target")
	}
	target, ok := s.expect(TokenIdentifier, TokenRawIdentifier)
	if !ok {
		var err = s.reportExpectToken(target, TokenIdentifier, TokenRawIdentifier)
		if target.Type == TokenLiteralString {
			err = span.ReportNoSpan(span.Info, "use raw identifiers istead of strings")
		}
		return Program{}, err
	}
	var theErr error
	declarations := make([]Declaration, 0)
	for s.peek() != nil {
		declaration, err := s.ParseDeclaration()
		if err != nil {
			theErr = err
			s.RestoreFromError()
			continue
		}
		declarations = append(declarations, declaration)
	}
	return Program{
		Target:       *target,
		Declarations: declarations,
		Span:         tokenTarget.Span,
	}, theErr
}

func (s *Parser) todo(token *Token) error {
	return s.reportToken(token, span.Error, "not implemented yet")
}

func (s *Parser) ParseDeclaration() (Declaration, error) {
	token := s.consume()
	if token == nil {
		return nil, s.reportToken(nil, span.Error, "unexpected EOF")
	}
	switch token.Type {
	case TokenKeywordFunc:
		name, ok := s.expect(TokenIdentifier, TokenRawIdentifier)
		if !ok {
			return nil, s.reportExpectToken(name, TokenIdentifier, TokenRawIdentifier)
		}
		openParen, ok := s.expect(TokenOpenParen)
		if !ok {
			return nil, s.reportExpectToken(openParen, TokenOpenParen)
		}
		closeParen, ok := s.expect(TokenCloseParen)
		if !ok {
			return nil, s.reportExpectToken(closeParen, TokenCloseParen)
		}
		body, err := s.ParseBlock()
		if err != nil {
			return nil, err
		}
		return &FunctionDeclaration{
			Name: *name,
			Span: token.Span.Merge(body.Span),
		}, nil
	case TokenKeywordVar:
		return nil, s.todo(token)
	}
	return nil, s.reportExpectToken(token, TokenKeywordFunc, TokenKeywordVar)
}

func (s *Parser) ParseBlock() (Block, error) {
	openBrace, ok := s.expect(TokenOpenBrace)
	if !ok {
		return Block{}, s.reportExpectToken(openBrace, TokenOpenBrace)
	}
	statements := make([]Statement, 0)
	closeBrace, ok := s.expect(TokenCloseBrace)
	if !ok {
		return Block{}, s.reportExpectToken(closeBrace, TokenCloseBrace)
	}
	return Block{
		Statements: statements,
		Span:       openBrace.Span.Merge(closeBrace.Span),
	}, nil
}
