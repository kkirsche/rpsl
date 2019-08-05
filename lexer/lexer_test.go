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
		testExpectation{token.CLASS_MAINTAINER, "mntner", 1},
		testExpectation{token.DATA_NIC_HANDLE, "TEST-MNT", 1},
		testExpectation{token.ATTR_DESCRIPTION, "descr", 2},
		testExpectation{token.DATA_STRING, "unįcöde tæst2 🌈🦄", 2}, // note that we're ending on column 34 because the emoji are width 2
		testExpectation{token.ATTR_ADMIN_CONTACT, "admin-c", 3},
		testExpectation{token.DATA_NIC_HANDLE, "PERSON-TEST", 3},
		testExpectation{token.ATTR_NOTIFY_EMAIL, "notify", 4},
		testExpectation{token.DATA_EMAIL, "notify@example.net", 4},
		testExpectation{token.ATTR_UPDATED_TO_EMAIL, "upd-to", 5},
		testExpectation{token.DATA_EMAIL, "upd-to@example.net", 5},
		testExpectation{token.ATTR_MAINTAINER_NOTIFY_EMAIL, "mnt-nfy", 6},
		testExpectation{token.DATA_EMAIL, "mnt-nfy@example.net", 6},
		testExpectation{token.ATTR_MAINTAINER_NOTIFY_EMAIL, "mnt-nfy", 7},
		testExpectation{token.DATA_EMAIL, "mnt-nfy2@example.net", 7},
		testExpectation{token.ATTR_AUTHENTICATION, "auth", 8},
		testExpectation{token.DATA_PGP_KEY, "80F238C6", 8},
		testExpectation{token.ATTR_AUTHENTICATION, "auth", 9},
		testExpectation{token.DATA_CRYPT_PASS, "LEuuhsBJNFV0Q", 9},
		testExpectation{token.ATTR_AUTHENTICATION, "auth", 10},
		testExpectation{token.DATA_MD5_PASS, "$1$fgW84Y9r$kKEn9MUq8PChNKpQhO6BM.", 10},
		testExpectation{token.ATTR_AUTHENTICATION, "auth", 11},
		testExpectation{token.DATA_MAIL_FROM_PASS, "auth@example.net", 11},
		testExpectation{token.ATTR_AUTHENTICATION, "auth", 12},
		testExpectation{token.DATA_NO_AUTH, "NONE", 12},
		testExpectation{token.ATTR_MAINTAINED_BY, "mnt-by", 13},
		testExpectation{token.DATA_NIC_HANDLE, "TEST-MNT", 13},
		testExpectation{token.ATTR_MAINTAINED_BY, "mnt-by", 14},
		testExpectation{token.DATA_NIC_HANDLE, "OTHER1-MNT", 14},
		testExpectation{token.DATA_NIC_HANDLE, "OTHER2-MNT", 14},
		testExpectation{token.ATTR_CHANGED_AT_AND_BY, "changed", 15},
		testExpectation{token.DATA_EMAIL, "changed@example.com", 15},
		testExpectation{token.DATA_DATE, "20190701", 15},
		testExpectation{token.ATTR_REMARKS, "remarks", 16},
		testExpectation{token.DATA_STRING, "unįcöde tæst 🌈🦄", 16},
		testExpectation{token.ATTR_TECHNICAL_CONTACT, "tech-c", 17},
		testExpectation{token.DATA_NIC_HANDLE, "PERSON-TEST", 17},
		testExpectation{token.ATTR_REGISTRY_SOURCE, "source", 18},
		testExpectation{token.DATA_REGISTRY_NAME, "TEST", 18},
		testExpectation{token.ATTR_REMARKS, "remarks", 19},
		testExpectation{token.DATA_STRING, "remark", 19},
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

func TestLexPerson(t *testing.T) {
	input := `person:         Test person
address:        DashCare BV
address:        Amsterdam
address:        The Netherlands
phone:          +31 20 000 0000
nic-hdl:        PERSON-TEST
mnt-by:         TEST-MNT
e-mail:         email@example.com
notify:         notify@example.com
changed:        changed@example.com 20190701 # comment
source:         TEST
`

	tests := testExpectations{
		testExpectation{token.CLASS_PERSON, "person", 1},
		testExpectation{token.DATA_STRING, "Test person", 1},
		testExpectation{token.ATTR_ADDRESS, "address", 2},
		testExpectation{token.DATA_STRING, "DashCare BV", 2},
		testExpectation{token.ATTR_ADDRESS, "address", 3},
		testExpectation{token.DATA_STRING, "Amsterdam", 3},
		testExpectation{token.ATTR_ADDRESS, "address", 4},
		testExpectation{token.DATA_STRING, "The Netherlands", 4},
		testExpectation{token.ATTR_PHONE_NUMBER, "phone", 5},
		testExpectation{token.DATA_TELEPHONE_OR_FAX_NUMBER, "+31 20 000 0000", 5},
		testExpectation{token.ATTR_NIC_HANDLE, "nic-hdl", 6},
		testExpectation{token.DATA_NIC_HANDLE, "PERSON-TEST", 6},
		testExpectation{token.ATTR_MAINTAINED_BY, "mnt-by", 7},
		testExpectation{token.DATA_NIC_HANDLE, "TEST-MNT", 7},
		testExpectation{token.ATTR_EMAIL, "e-mail", 8},
		testExpectation{token.DATA_EMAIL, "email@example.com", 8},
		testExpectation{token.ATTR_NOTIFY_EMAIL, "notify", 9},
		testExpectation{token.DATA_EMAIL, "notify@example.com", 9},
		testExpectation{token.ATTR_CHANGED_AT_AND_BY, "changed", 10},
		testExpectation{token.DATA_EMAIL, "changed@example.com", 10},
		testExpectation{token.DATA_DATE, "20190701", 10},
		testExpectation{token.ATTR_REGISTRY_SOURCE, "source", 11},
		testExpectation{token.DATA_REGISTRY_NAME, "TEST", 11},
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

func TestLexRole(t *testing.T) {
	input := `role:           DashCare BV
address:        address
phone:          +31200000000
fax-no:         +31200000000
e-mail:         unread@example.com
admin-c:        PERSON-TEST
tech-c:         PERSON-TEST
nic-hdl:        ROLE-TEST
notify:         notify@example.com
mnt-by:         TEST-MNT
changed:        changed@example.com 20190701 # comment
source:         TEST
remarks:        remark
`

	tests := testExpectations{
		testExpectation{token.CLASS_ROLE, "role", 1},
		testExpectation{token.DATA_STRING, "DashCare BV", 1},
		testExpectation{token.ATTR_ADDRESS, "address", 2},
		testExpectation{token.DATA_STRING, "address", 2},
		testExpectation{token.ATTR_PHONE_NUMBER, "phone", 3},
		testExpectation{token.DATA_TELEPHONE_OR_FAX_NUMBER, "+31200000000", 3},
		testExpectation{token.ATTR_FAX_NUMBER, "fax-no", 4},
		testExpectation{token.DATA_TELEPHONE_OR_FAX_NUMBER, "+31200000000", 4},
		testExpectation{token.ATTR_EMAIL, "e-mail", 5},
		testExpectation{token.DATA_EMAIL, "unread@example.com", 5},
		testExpectation{token.ATTR_ADMIN_CONTACT, "admin-c", 6},
		testExpectation{token.DATA_NIC_HANDLE, "PERSON-TEST", 6},
		testExpectation{token.ATTR_TECHNICAL_CONTACT, "tech-c", 7},
		testExpectation{token.DATA_NIC_HANDLE, "PERSON-TEST", 7},
		testExpectation{token.ATTR_NIC_HANDLE, "nic-hdl", 8},
		testExpectation{token.DATA_NIC_HANDLE, "ROLE-TEST", 8},
		testExpectation{token.ATTR_NOTIFY_EMAIL, "notify", 9},
		testExpectation{token.DATA_EMAIL, "notify@example.com", 9},
		testExpectation{token.ATTR_MAINTAINED_BY, "mnt-by", 10},
		testExpectation{token.DATA_NIC_HANDLE, "TEST-MNT", 10},
		testExpectation{token.ATTR_CHANGED_AT_AND_BY, "changed", 11},
		testExpectation{token.DATA_EMAIL, "changed@example.com", 11},
		testExpectation{token.DATA_DATE, "20190701", 11},
		testExpectation{token.ATTR_REGISTRY_SOURCE, "source", 12},
		testExpectation{token.DATA_REGISTRY_NAME, "TEST", 12},
		testExpectation{token.ATTR_REMARKS, "remarks", 13},
		testExpectation{token.DATA_STRING, "remark", 13},
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
