package lexer

import (
	"testing"

	"github.com/kkirsche/rpsl/token"
	"github.com/stretchr/testify/assert"
)

type testExpectations []testExpectation

type testExpectation struct {
	typ     token.Type
	literal string
	column  int
	line    int
}

func TestLexMaintainer(t *testing.T) {
	input := `mntner:         TEST-MNT
descr:          unįcöde tæst2 🌈🦄
admin-c:        PERSON-TEST
notify:         notify@example.net
upd-to:         upd-to@example.net
mnt-nfy:        mnt-nfy@example.net
mnt-nfy:        mnt-nfy2@example.net
auth:           PGPKey-80F238C6
auth:           CRYPT-PW LEuuhsBJNFV0Q  # crypt-password
auth:           MD5-pw $1$fgW84Y9r$kKEn9MUq8PChNKpQhO6BM.  # md5-password
mnt-by:         TEST-MNT
mnt-by:         OTHER1-MNT,OTHER2-MNT
changed:        changed@example.com 20190701 # comment
remarks:        unįcöde tæst 🌈🦄
source:         TEST
remarks:        remark
`

	tests := testExpectations{
		testExpectation{token.MAINTAINER, "mntner", 6, 1},
		testExpectation{token.STRING, "TEST-MNT", 24, 1},
		testExpectation{token.DESCRIPTION, "descr", 5, 2},
		testExpectation{token.STRING, "unįcöde tæst2 🌈🦄", 34, 2}, // note that we're ending on column 34 because the emoji are width 2
		testExpectation{token.ADMIN_CONTACT, "admin-c", 7, 3},
		testExpectation{token.STRING, "PERSON-TEST", 27, 3},
		testExpectation{token.NOTIFY_EMAIL, "notify", 6, 4},
		testExpectation{token.EMAIL, "notify@example.net", 34, 4},
		testExpectation{token.UPDATED_TO_EMAIL, "upd-to", 6, 5},
		testExpectation{token.EMAIL, "upd-to@example.net", 34, 5},
		testExpectation{token.MAINTAINER_NOTIFY_EMAIL, "mnt-nfy", 7, 6},
		testExpectation{token.EMAIL, "mnt-nfy@example.net", 35, 6},
		testExpectation{token.MAINTAINER_NOTIFY_EMAIL, "mnt-nfy", 7, 7},
		testExpectation{token.EMAIL, "mnt-nfy2@example.net", 36, 7},
		testExpectation{token.EOF, "", 0, 0},
	}

	l := Lex("maintainer-object", input)

	for _, tt := range tests {
		tok := l.NextToken()
		failure := false

		if !assert.Equal(t, tt.typ, tok.Type, "Invalid token type '%s', expected '%s'", tok.Type, tt.typ) {
			failure = true
		}

		if !assert.Equal(t, tt.literal, tok.Literal, "Invalid token literal '%s', expected '%s'", tok.Literal, tt.literal) {
			failure = true
		}

		if !assert.Equal(t, tt.column, tok.Column, "Invalid column number %d for token literal '%s'", tok.Column, tok.Literal) {
			failure = true
		}

		if !assert.Equal(t, tt.line, tok.Line, "Invalid line number %d for token literal '%s'", tok.Line, tok.Literal) {
			failure = true
		}

		if failure {
			t.FailNow()
		}
	}
}
