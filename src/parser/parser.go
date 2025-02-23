package parser

import (

	"github.com/Jonaires777/src/lexer"
	"github.com/Jonaires777/src/token"
)

type Parser struct {
	l *lexer.Lexer
	currToken token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// por enquanto retorna uma string
func (p *Parser) ParseCommand() string {
	switch p.currToken.Type {
	case token.CREATE:
		return "CREATE command recognised"
	case token.REMOVE:
		return "REMOVE command recognised"
	case token.LIST:
		return "LIST command recognised"
	case token.ORDER:
		return "ORDER command recognised"
	case token.READ:
		return "READ command recognised"
	case token.CONCAT:
		return "CONCAT command recognised"
	default:
		return "Unknown Command"
	}
}