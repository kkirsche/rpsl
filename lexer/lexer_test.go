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

func TestLexAutonomousSystemNumber(t *testing.T) {
	input := `aut-num:        as065537
as-name:        TEST-AS
descr:          description
+               foo
remarks:        ---> Uplinks
export:         to AS3356 announce AS-SETTEST
import:         from AS3356 accept ANY
export:         to AS174 announce AS-SETTEST
import:         from AS174 accept ANY
mp-export:      afi ipv6.unicast to AS174 announce AS-SETTEST
mp-import:      afi ipv6.unicast from AS174 accept ANY
export:         to AS8359 announce AS-SETTEST
import:         from AS8359 accept ANY
export:         to AS3257 announce AS-SETTEST
import:         from AS3257 accept ANY
export:         to AS3549 announce AS-SETTEST
import:         from AS3549 accept ANY
export:         to AS9002 announce AS-SETTEST
import:         from AS9002 accept ANY
mp-export:      afi ipv6.unicast to AS9002 announce AS-SETTEST
mp-import:      afi ipv6.unicast from AS9002 accept ANY
remarks:        ---> Peers
export:         to AS31117 announce AS-SETTEST AS-UAIX
import:         from AS31117 accept AS-ENERGOTEL
export:         to AS8501 announce AS-SETTEST AS-UAIX
import:         from AS8501 accept AS-PLNET
export:         to AS35297 announce AS-SETTEST
import:         from AS35297 accept AS-DL-WORLD
export:         to AS13188 announce AS-SETTEST
import:         from AS13188 accept AS-BANKINFORM
export:         to AS12389 announce AS-SETTEST
import:         from AS12389 accept AS-ROSTELECOM
export:         to AS35395 announce AS-SETTEST
import:         from AS35395 accept AS-GECIXUAIX
export:         to AS50952 announce AS-SETTEST
import:         from AS50952 accept AS-DATAIX
admin-c:        PERSON-TEST
tech-c:         PERSON-TEST
mnt-by:         TEST-MNT
changed:        changed@example.com 20190701 # comment
source:         TEST
remarks:        remark
`

	tests := testExpectations{
		testExpectation{token.CLASS_AUT_NUM, "aut-num", 1},
		testExpectation{token.DATA_ASN, "as065537", 1},
		testExpectation{token.ATTR_AS_NAME, "as-name", 2},
		testExpectation{token.DATA_NIC_HANDLE, "TEST-AS", 2},
		testExpectation{token.ATTR_DESCRIPTION, "descr", 3},
		testExpectation{token.DATA_STRING, "description", 3},
		testExpectation{token.ATTR_CONTINUATION, "+", 4},
		testExpectation{token.DATA_STRING, "foo", 4},
		testExpectation{token.ATTR_REMARKS, "remarks", 5},
		testExpectation{token.DATA_STRING, "---> Uplinks", 5},
		testExpectation{token.ATTR_EXPORT, "export", 6},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS3356 announce AS-SETTEST", 6},
		testExpectation{token.ATTR_IMPORT, "import", 7},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS3356 accept ANY", 7},
		testExpectation{token.ATTR_EXPORT, "export", 8},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS174 announce AS-SETTEST", 8},
		testExpectation{token.ATTR_IMPORT, "import", 9},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS174 accept ANY", 9},
		testExpectation{token.ATTR_MULTI_PROTO_EXPORT_POLICY, "mp-export", 10},
		testExpectation{token.DATA_MULTI_PROTO_EXPORT_POLICY, "afi ipv6.unicast to AS174 announce AS-SETTEST", 10},
		testExpectation{token.ATTR_MULTI_PROTO_IMPORT_POLICY, "mp-import", 11},
		testExpectation{token.DATA_MULTI_PROTO_IMPORT_POLICY, "afi ipv6.unicast from AS174 accept ANY", 11},
		testExpectation{token.ATTR_EXPORT, "export", 12},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS8359 announce AS-SETTEST", 12},
		testExpectation{token.ATTR_IMPORT, "import", 13},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS8359 accept ANY", 13},
		testExpectation{token.ATTR_EXPORT, "export", 14},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS3257 announce AS-SETTEST", 14},
		testExpectation{token.ATTR_IMPORT, "import", 15},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS3257 accept ANY", 15},
		testExpectation{token.ATTR_EXPORT, "export", 16},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS3549 announce AS-SETTEST", 16},
		testExpectation{token.ATTR_IMPORT, "import", 17},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS3549 accept ANY", 17},
		testExpectation{token.ATTR_EXPORT, "export", 18},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS9002 announce AS-SETTEST", 18},
		testExpectation{token.ATTR_IMPORT, "import", 19},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS9002 accept ANY", 19},
		testExpectation{token.ATTR_MULTI_PROTO_EXPORT_POLICY, "mp-export", 20},
		testExpectation{token.DATA_MULTI_PROTO_EXPORT_POLICY, "afi ipv6.unicast to AS9002 announce AS-SETTEST", 20},
		testExpectation{token.ATTR_MULTI_PROTO_IMPORT_POLICY, "mp-import", 21},
		testExpectation{token.DATA_MULTI_PROTO_IMPORT_POLICY, "afi ipv6.unicast from AS9002 accept ANY", 21},
		testExpectation{token.ATTR_REMARKS, "remarks", 22},
		testExpectation{token.DATA_STRING, "---> Peers", 22},
		testExpectation{token.ATTR_EXPORT, "export", 23},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS31117 announce AS-SETTEST AS-UAIX", 23},
		testExpectation{token.ATTR_IMPORT, "import", 24},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS31117 accept AS-ENERGOTEL", 24},
		testExpectation{token.ATTR_EXPORT, "export", 25},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS8501 announce AS-SETTEST AS-UAIX", 25},
		testExpectation{token.ATTR_IMPORT, "import", 26},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS8501 accept AS-PLNET", 26},
		testExpectation{token.ATTR_EXPORT, "export", 27},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS35297 announce AS-SETTEST", 27},
		testExpectation{token.ATTR_IMPORT, "import", 28},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS35297 accept AS-DL-WORLD", 28},
		testExpectation{token.ATTR_EXPORT, "export", 29},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS13188 announce AS-SETTEST", 29},
		testExpectation{token.ATTR_IMPORT, "import", 30},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS13188 accept AS-BANKINFORM", 30},
		testExpectation{token.ATTR_EXPORT, "export", 31},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS12389 announce AS-SETTEST", 31},
		testExpectation{token.ATTR_IMPORT, "import", 32},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS12389 accept AS-ROSTELECOM", 32},
		testExpectation{token.ATTR_EXPORT, "export", 33},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS35395 announce AS-SETTEST", 33},
		testExpectation{token.ATTR_IMPORT, "import", 34},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS35395 accept AS-GECIXUAIX", 34},
		testExpectation{token.ATTR_EXPORT, "export", 35},
		testExpectation{token.DATA_EXPORT_POLICY, "to AS50952 announce AS-SETTEST", 35},
		testExpectation{token.ATTR_IMPORT, "import", 36},
		testExpectation{token.DATA_IMPORT_POLICY, "from AS50952 accept AS-DATAIX", 36},
		testExpectation{token.ATTR_ADMIN_CONTACT, "admin-c", 37},
		testExpectation{token.DATA_NIC_HANDLE, "PERSON-TEST", 37},
		testExpectation{token.ATTR_TECHNICAL_CONTACT, "tech-c", 38},
		testExpectation{token.DATA_NIC_HANDLE, "PERSON-TEST", 38},
		testExpectation{token.ATTR_MAINTAINED_BY, "mnt-by", 39},
		testExpectation{token.DATA_NIC_HANDLE, "TEST-MNT", 39},
		testExpectation{token.ATTR_CHANGED_AT_AND_BY, "changed", 40},
		testExpectation{token.DATA_EMAIL, "changed@example.com", 40},
		testExpectation{token.DATA_DATE, "20190701", 40},
		testExpectation{token.ATTR_REGISTRY_SOURCE, "source", 41},
		testExpectation{token.DATA_REGISTRY_NAME, "TEST", 41},
		testExpectation{token.ATTR_REMARKS, "remarks", 42},
		testExpectation{token.DATA_STRING, "remark", 42},
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
