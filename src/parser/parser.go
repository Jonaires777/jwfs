package parser

import (
	"fmt"
	"strconv"

	"github.com/Jonaires777/src/constants"
	"github.com/Jonaires777/src/filemanager"
	"github.com/Jonaires777/src/lexer"
	"github.com/Jonaires777/src/token"
)

type Parser struct {
	l         *lexer.Lexer
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

func (p *Parser) ParseCommand() string {
	switch p.currToken.Type {
	case token.CREATE:
		return p.parseCreate()
	case token.REMOVE:
		return p.parseRemove()
	case token.LIST:
		return p.parseList()
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

func (p *Parser) parseCreate() string {
	p.nextToken()
	if p.currToken.Type != token.IDENT {
		return "Erro: esperado um nome de arquivo após create"
	}
	filename := p.currToken.Literal

	p.nextToken()
	size, err := strconv.Atoi(p.currToken.Literal)
	if err != nil {
		return "Erro: tamanho do arquivo deve ser um número inteiro"
	}

	err = filemanager.CreateFile(filename, size)
	if err != nil {
		return fmt.Sprintf("Erro ao criar o arquivo: %v", err)
	}

	return fmt.Sprintf("Arquivo '%s' criado com sucesso, tamanho: %d", filename, size)
}

func (p *Parser) parseList() string {
	files, totalUsed, err := filemanager.ListFiles()
	if err != nil {
		return fmt.Sprintf("Erro ao listar arquivos: %v", err)
	}

	if len(files) == 0 {
		return "Nenhum arquivo encontrado"
	}

	var filesList string
	for _, file := range files {
		filesList += fmt.Sprintf("Nome: %s, Tamanho: %d\n", file.Filename, file.Size)
	}

	return fmt.Sprintf("Arquivos:\n%s\nEspaço total usado: %d, Espaço total disponível: %d", filesList, totalUsed, constants.Disksize-totalUsed)
}

func (p *Parser) parseRemove() string {
	p.nextToken()
	if p.currToken.Type != token.IDENT {
		return "Erro: esperado um nome de arquivo após remove"
	}

	filename := p.currToken.Literal

	err := filemanager.RemoveFile(filename)
	if err != nil {
		return fmt.Sprintf("Erro ao remover o arquivo: %v", err)
	}

	return fmt.Sprintf("Arquivo '%s' removido com sucesso", filename)
}