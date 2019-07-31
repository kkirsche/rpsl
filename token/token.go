package token

// Type is the token type
type Type int

// Token represents an emitted token from the lexer containing both it's type
// and the literal value of the token
type Token struct {
	Type    Type
	Literal string
	Column  int
	Line    int
}

const (
	// EOF is the end of file token
	EOF Type = iota

	// ILLEGAL is used when an illegal token appears in the input stream
	ILLEGAL

	// NEWLINE is a line feed character
	NEWLINE

	// INT - a numeric integer
	// Official grammar: [[:digit:]]+
	INT

	// SIGNED_INT - a signed integer, positive or negative
	// Official grammar: [+-]?{INT}
	SIGNED_INT

	// ASNO - an autonomous system number value
	// Official grammar: AS{INT}
	ASNO

	// ASNAME - an autonomous system name
	// Official grammar: AS-[[:alnum:]_-]*[[:alnum:]]
	ASNAME
)

var names = map[Type]string{
	EOF:        "EOF",
	ILLEGAL:    "ILLEGAL",
	NEWLINE:    "NEWLINE",
	INT:        "INT",
	SIGNED_INT: "SIGNED_INT",
	ASNO:       "ASNO",
	ASNAME:     "ASNAME",
}

func LookupName(token Type) string {
	if tok, ok := names[token]; ok {
		return tok
	}

	return "UNKNOWN"
}
