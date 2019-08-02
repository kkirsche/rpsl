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
auth:           MAIL-FROM auth@example.net
auth:           NONE
mnt-by:         TEST-MNT
mnt-by:         OTHER1-MNT,OTHER2-MNT
changed:        changed@example.com 20190701 # comment
remarks:        unįcöde tæst 🌈🦄
tech-c:					PERSON-TEST
source:         TEST
remarks:        remark
`

	tests := testExpectations{
		testExpectation{token.MAINTAINER, "mntner", 1},
		testExpectation{token.NIC_HANDLE, "TEST-MNT", 1},
		testExpectation{token.DESCRIPTION, "descr", 2},
		testExpectation{token.STRING, "unįcöde tæst2 🌈🦄", 2}, // note that we're ending on column 34 because the emoji are width 2
		testExpectation{token.ADMIN_CONTACT, "admin-c", 3},
		testExpectation{token.NIC_HANDLE, "PERSON-TEST", 3},
		testExpectation{token.NOTIFY_EMAIL, "notify", 4},
		testExpectation{token.EMAIL, "notify@example.net", 4},
		testExpectation{token.UPDATED_TO_EMAIL, "upd-to", 5},
		testExpectation{token.EMAIL, "upd-to@example.net", 5},
		testExpectation{token.MAINTAINER_NOTIFY_EMAIL, "mnt-nfy", 6},
		testExpectation{token.EMAIL, "mnt-nfy@example.net", 6},
		testExpectation{token.MAINTAINER_NOTIFY_EMAIL, "mnt-nfy", 7},
		testExpectation{token.EMAIL, "mnt-nfy2@example.net", 7},
		testExpectation{token.AUTHENTICATION, "auth", 8},
		testExpectation{token.PGP_KEY, "80F238C6", 8},
		testExpectation{token.AUTHENTICATION, "auth", 9},
		testExpectation{token.CRYPT_PASS, "LEuuhsBJNFV0Q", 9},
		testExpectation{token.AUTHENTICATION, "auth", 10},
		testExpectation{token.MD5_PASS, "$1$fgW84Y9r$kKEn9MUq8PChNKpQhO6BM.", 10},
		testExpectation{token.AUTHENTICATION, "auth", 11},
		testExpectation{token.MAIL_FROM_PASS, "auth@example.net", 11},
		testExpectation{token.AUTHENTICATION, "auth", 12},
		testExpectation{token.NO_AUTH, "NONE", 12},
		testExpectation{token.MAINTAINED_BY, "mnt-by", 13},
		testExpectation{token.NIC_HANDLE, "TEST-MNT", 13},
		testExpectation{token.MAINTAINED_BY, "mnt-by", 14},
		testExpectation{token.NIC_HANDLE, "OTHER1-MNT", 14},
		testExpectation{token.NIC_HANDLE, "OTHER2-MNT", 14},
		testExpectation{token.CHANGED_AT_AND_BY, "changed", 15},
		testExpectation{token.EMAIL, "changed@example.com", 15},
		testExpectation{token.DATE, "20190701", 15},
		testExpectation{token.REMARKS, "remarks", 16},
		testExpectation{token.STRING, "unįcöde tæst 🌈🦄", 16},
		testExpectation{token.TECHNICAL_CONTACT, "tech-c", 17},
		testExpectation{token.NIC_HANDLE, "PERSON-TEST", 17},
		testExpectation{token.REGISTRY_SOURCE, "source", 18},
		testExpectation{token.REGISTRY_NAME, "TEST", 18},
		testExpectation{token.REMARKS, "remarks", 19},
		testExpectation{token.STRING, "remark", 19},
		testExpectation{token.EOF, "", 0},
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

		if !assert.Equal(t, tt.line, tok.Line, "Invalid line number %d for token literal '%s'", tok.Line, tok.Literal) {
			failure = true
		}

		if failure {
			t.FailNow()
		}
	}
}
