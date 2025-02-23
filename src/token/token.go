package token

const (

	// Special Tokens
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers and number
	IDENT = "IDENT"
	INT   = "INT"

	// Separators
	COMMA     = ","
	SEMICOLON = ";"
	NEWLINE   = "\n"

	// Commands
	CREATE = "CREATE"
	REMOVE = "REMOVE"
	LIST   = "LIST"
	ORDER  = "ORDER"
	READ   = "READ"
	CONCAT = "CONCAT"
)

var keywords = map[string]TokenType{
	"create": CREATE,
	"remove": REMOVE,
	"list":   LIST,
	"order":  ORDER,
	"read":   READ,
	"concat": CONCAT,
}

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}
