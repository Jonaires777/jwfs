package lexer

import (
	"github.com/Jonaires777/src/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (poins to current char in examination)
	readPosition int  // current reading position in input (after current char in examination)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhiteSpace()

	if isLetter(l.ch) {
		literal := l.readIdentifier()
		tok.Type = token.LookupIdent(literal)
		tok.Literal = literal
		return tok
	} else if isDigit(l.ch) {
		tok.Type = token.INT
		tok.Literal = l.readNumber()
		return tok
	} else {
		switch l.ch {
		case 0:
			tok.Literal = ""
			tok.Type = token.EOF
		default:
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhiteSpace() {
	if l.ch == ' ' || l.ch == '\n' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
