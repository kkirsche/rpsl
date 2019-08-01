package token

import (
	"fmt"
)

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

// The following list of constants are the tokens. These use integers to reduce
// memory usage and improve lexing performance
const (
	// EOF is the end of file token
	EOF Type = iota

	// ILLEGAL is used when an illegal token appears in the input stream
	ILLEGAL

	// Data Types
	STRING
	NUMBER
	EMAIL
	PGP_KEY
	CRYPT_PASS
	MD5_PASS
	MAIL_FROM_PASS
	NO_AUTH

	// Object Classes
	MAINTAINER
	PERSON
	ROLE
	AUT_NUM
	AS_SET
	ROUTE
	ROUTE6
	ROUTE_SET
	FILTER_SET
	ROUTER
	ROUTER_SET
	PEERING_SET
	DICTIONARY

	// Object Attributes
	DESCRIPTION
	AUTHENTICATION
	UPDATED_TO_EMAIL
	MAINTAINER_NOTIFY_EMAIL
	TECHNICAL_CONTACT
	ADMIN_CONTACT
	REMARKS
	NOTIFY_EMAIL
	MAINTAINED_BY
	CHANGED_AT_AND_BY
	RECORD_SOURCE
)

var names = map[Type]string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",
	// Data Types
	STRING:         "STRING",
	NUMBER:         "NUMBER",
	EMAIL:          "EMAIL",
	PGP_KEY:        "PGP_KEY",
	CRYPT_PASS:     "CRYPT_PASS",
	MD5_PASS:       "MD5_PASS",
	MAIL_FROM_PASS: "MAIL_FROM_PASS",
	NO_AUTH:        "NO_AUTH",
	// Object Classes
	MAINTAINER:  "MAINTAINER",
	PERSON:      "PERSON",
	ROLE:        "ROLE",
	AUT_NUM:     "AUT_NUM",
	AS_SET:      "AS_SET",
	ROUTE:       "ROUTE",
	ROUTE6:      "ROUTE6",
	ROUTE_SET:   "ROUTE_SET",
	FILTER_SET:  "FILTER_SET",
	ROUTER:      "ROUTER",
	ROUTER_SET:  "ROUTER_SET",
	PEERING_SET: "PEERING_SET",
	DICTIONARY:  "DICTIONARY",
	// Object Attributes
	DESCRIPTION:             "DESCRIPTION",
	AUTHENTICATION:          "AUTHENTICATION",
	UPDATED_TO_EMAIL:        "UPDATED_TO_EMAIL",
	MAINTAINER_NOTIFY_EMAIL: "MAINTAINER_NOTIFY_EMAIL",
	TECHNICAL_CONTACT:       "TECHNICAL_CONTACT",
	ADMIN_CONTACT:           "ADMIN_CONTACT",
	REMARKS:                 "REMARKS",
	NOTIFY_EMAIL:            "NOTIFY_EMAIL",
	MAINTAINED_BY:           "MAINTAINED_BY",
	CHANGED_AT_AND_BY:       "CHANGED_AT_AND_BY",
	RECORD_SOURCE:           "RECORD_SOURCE",
}

var objectStrings = map[Type]string{
	// Object Classes
	MAINTAINER:  "mntner",
	PERSON:      "person",
	ROLE:        "role",
	AUT_NUM:     "aut-num",
	AS_SET:      "as-set",
	ROUTE:       "route",
	ROUTE6:      "route6",
	ROUTE_SET:   "route-set",
	FILTER_SET:  "filter-set",
	ROUTER:      "inet-rtr",
	ROUTER_SET:  "rtr-set",
	PEERING_SET: "peering-set",
	DICTIONARY:  "dictionary",
	// Object Attributes
	DESCRIPTION:             "descr",
	AUTHENTICATION:          "auth",
	UPDATED_TO_EMAIL:        "upd-to",
	MAINTAINER_NOTIFY_EMAIL: "mnt-nfy",
	TECHNICAL_CONTACT:       "tech-c",
	ADMIN_CONTACT:           "admin-c",
	REMARKS:                 "remarks",
	NOTIFY_EMAIL:            "notify",
	MAINTAINED_BY:           "mnt-by",
	CHANGED_AT_AND_BY:       "changed",
	RECORD_SOURCE:           "source",
	// Object Value Prefixes
	PGP_KEY:        "PGPKey-",
	CRYPT_PASS:     "CRYPT-PW",
	MD5_PASS:       "MD5-pw",
	MAIL_FROM_PASS: "MAIL-FROM",
	NO_AUTH:        "NONE",
}

// String is used to translate a token type into it's name
func (t Type) String() string {
	if tok, ok := names[t]; ok {
		return tok
	}

	panic(fmt.Sprintf("Unknown token type received: %d", t))
}

// Name is a method which allows us to get the name of a token
// type. If we can't find one, we panic because this is to assist the lexer and
// it should not be possible to call this translation function on an invalid
// value
func (t Type) Name() string {
	if tok, ok := objectStrings[t]; ok {
		return tok
	}

	panic(fmt.Sprintf("Unknown token type received: %d", t))
}
