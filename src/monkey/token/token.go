package token


type TokenType string

type Token struct {
	Type	TokenType
	Literal string
}

// possible tokens:
const (
	ILLEGAL = "ILLEGAL"
	EOF	= "EOF"

	IDENT	= "IDENT"
	INT	= "INT"

	ASSIGN	= "="
	PLUS	= "+"

	COMMA	= ","
	SEMICOLON = ";"

	LPAREN	= "("
	RPAREN	= ")"
	LBRACE	= "{"
	RBRACE	= "}"

	FUNCTION= "FUNCTION"
	LET	= "LET"
)

var keywords = map[string]TokenType{
	"fn": FUNCTION,
	"let": LET,
}

func LookupIdent(ident string) TokenType {
	// fn or let
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	// ah must be IDENT
	return IDENT
}
