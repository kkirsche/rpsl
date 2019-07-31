package lexer

import (
	"testing"

	"github.com/kkirsche/rpsl/token"
	"github.com/stretchr/testify/assert"
)

type expected struct {
	tType    token.Type
	tLiteral string
	tColumn  int
	tLine    int
}

func TestNextChar(t *testing.T) {
	input := `1234
+1234
-1234
AS701
AS-A123
AS-12A3
AS-12-
`

	tests := []expected{
		expected{token.INT, "1234", 1, 1},
		expected{token.NEWLINE, "\n", 5, 1},
		expected{token.SIGNED_INT, "+1234", 1, 2},
		expected{token.NEWLINE, "\n", 6, 2},
		expected{token.SIGNED_INT, "-1234", 1, 3},
		expected{token.NEWLINE, "\n", 6, 3},
		expected{token.ASNO, "AS701", 1, 4},
		expected{token.NEWLINE, "\n", 6, 4},
		expected{token.ASNAME, "AS-A123", 1, 5},
		expected{token.NEWLINE, "\n", 8, 5},
		expected{token.ASNAME, "AS-12A3", 1, 6},
		expected{token.NEWLINE, "\n", 8, 6},
		expected{token.ILLEGAL, "AS-12-", 1, 7},
		expected{token.NEWLINE, "\n", 7, 7},
		expected{token.EOF, "", 0, 8},
	}

	lex := New(input)

	for _, tt := range tests {
		invalid := false
		tok := lex.NextToken()

		if !assert.Equal(t, tt.tType, tok.Type, "Invalid token type '%s', expected '%s'", token.LookupName(tok.Type), token.LookupName(tt.tType)) {
			invalid = true
		}
		if !assert.Equal(t, tt.tLiteral, tok.Literal, "Invalid token literal '%s', expected '%s'", tok.Literal, tt.tLiteral) {
			invalid = true
		}
		if !assert.Equal(t, tt.tColumn, tok.Column, "Invalid column number %d for token literal '%s'", tok.Column, tok.Literal) {
			invalid = true
		}
		if !assert.Equal(t, tt.tLine, tok.Line, "Invalid line number %d for token literal '%s'", tok.Line, tok.Literal) {
			invalid = true
		}

		if invalid {
			t.FailNow()
		}
	}
}
