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
		return p.parseOrder()
	case token.READ:
		return p.parseRead()
	case token.CONCAT:
		return p.parseConcat()
	case token.HELP:
		return p.parseHelp()
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

	return fmt.Sprintf("Arquivos:\n%s\nEspaço total usado: %d, Espaço total disponível: %d", filesList, totalUsed, constants.DiskSize-totalUsed)
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

func (p *Parser) parseRead() string {
	p.nextToken()
	if p.currToken.Type != token.IDENT {
		return "Erro: esperado um nome de arquivo após read"
	}

	filename := p.currToken.Literal

	p.nextToken()
	if p.currToken.Type != token.INT {
		return "Erro: esperado um número após o nome do arquivo"
	}

	startIdx, err := strconv.Atoi(p.currToken.Literal)
	if err != nil {
		return "Erro: índice inicial deve ser um número inteiro"
	}

	p.nextToken()
	if p.currToken.Type != token.INT {
		return "Erro: esperado um número após o índice inicial"
	}

	endIdx, err := strconv.Atoi(p.currToken.Literal)
	if err != nil {
		return "Erro: índice final deve ser um número inteiro"
	}

	data, err := filemanager.ReadFile(filename, int64(startIdx), int64(endIdx))
	if err != nil {
		return fmt.Sprintf("Erro ao ler o arquivo: %v", err)
	}

	return fmt.Sprintf("Conteúdo do arquivo '%s':\n %v", filename, data)
}

func (p *Parser) parseOrder() string {
	p.nextToken()
	if p.currToken.Type != token.IDENT {
		return "Erro: esperado um nome de arquivo após order"
	}

	filename := p.currToken.Literal

	duration, err := filemanager.OrderFile(filename)
	if err != nil {
		return fmt.Sprintf("Erro ao ordenar o arquivo: %v", err)
	}

	return fmt.Sprintf("Arquivo '%s' ordenado com sucesso\nTempo em ordenação: %dms", filename, duration)
}

func (p *Parser) parseConcat() string {
	p.nextToken()
	if p.currToken.Type != token.IDENT {
		return "Erro: esperado um nome de arquivo após concat"
	}

	filename1 := p.currToken.Literal

	p.nextToken()
	if p.currToken.Type != token.IDENT {
		return "Erro: esperado um nome de arquivo após o primeiro nome"
	}

	filename2 := p.currToken.Literal

	p.nextToken()
	if p.currToken.Type != token.IDENT {
		return "Erro: esperado um nome de arquivo após o segundo nome"
	}

	newFilename := p.currToken.Literal

	err := filemanager.ConcatFiles(filename1, filename2, newFilename)
	if err != nil {
		return fmt.Sprintf("Erro ao concatenar os arquivos: %v", err)
	}

	return fmt.Sprintf("Arquivos '%s' e '%s' concatenados com sucesso", filename1, filename2)
}

func (p *Parser) parseHelp() string {
	return `
Use os seguintes comandos para interagir com o sistema de arquivos:
create <filename> <size> - criar um novo arquivo com o tamanho fornecido
remove <filename> - remover um arquivo
list - listar todos os arquivos
order <filename> - ordenar um arquivo
read <filename> <startIdx> <endIdx> - ler um arquivo
concat <filename1> <filename2> <newFile> - concatenar dois arquivos em um novo arquivo
exit - sair do programa
`
}
