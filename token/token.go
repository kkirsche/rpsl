package token

import (
	"fmt"
	"strings"
)

// Type is the token type
type Type int

// Token represents an emitted token from the lexer containing both it's type
// and the literal value of the token
type Token struct {
	Type    Type
	Literal string
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
	DATA_ASN
	DATA_CRYPT_PASS
	DATA_DATE
	DATA_EMAIL
	DATA_MAIL_FROM_PASS
	DATA_MD5_PASS
	DATA_NIC_HANDLE
	DATA_NO_AUTH
	DATA_NUMBER
	DATA_PGP_KEY
	DATA_REGISTRY_NAME
	DATA_STRING
	DATA_TELEPHONE_OR_FAX_NUMBER
	DATA_EXPORT_POLICY
	DATA_IMPORT_POLICY

	// Object Classes
	CLASS_AS_SET
	CLASS_AUT_NUM
	CLASS_DICTIONARY
	CLASS_FILTER_SET
	CLASS_MAINTAINER
	CLASS_PEERING_SET
	CLASS_PERSON
	CLASS_ROLE
	CLASS_ROUTE
	CLASS_ROUTE6
	CLASS_ROUTER
	CLASS_ROUTER_SET
	CLASS_ROUTE_SET

	// Object Attributes
	ATTR_ADDRESS
	ATTR_ADMIN_CONTACT
	ATTR_AS_NAME
	ATTR_AUTHENTICATION
	ATTR_CHANGED_AT_AND_BY
	ATTR_CONTINUATION // +: indicates repeat of the last attribute type / name
	ATTR_DESCRIPTION
	ATTR_EMAIL
	ATTR_EXPORT
	ATTR_FAX_NUMBER
	ATTR_IMPORT
	ATTR_MAINTAINED_BY
	ATTR_MAINTAINER_NOTIFY_EMAIL
	ATTR_NIC_HANDLE
	ATTR_NOTIFY_EMAIL
	ATTR_PHONE_NUMBER
	ATTR_REGISTRY_SOURCE
	ATTR_REMARKS
	ATTR_TECHNICAL_CONTACT
	ATTR_UPDATED_TO_EMAIL
)

var names = map[Type]string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",
	// Data Types
	DATA_ASN:                     "DATA_ASN",
	DATA_CRYPT_PASS:              "DATA_CRYPT_PASS",
	DATA_DATE:                    "DATA_DATE",
	DATA_EMAIL:                   "DATA_EMAIL",
	DATA_MAIL_FROM_PASS:          "DATA_MAIL_FROM_PASS",
	DATA_MD5_PASS:                "DATA_MD5_PASS",
	DATA_NIC_HANDLE:              "DATA_NIC_HANDLE",
	DATA_NO_AUTH:                 "DATA_NO_AUTH",
	DATA_NUMBER:                  "DATA_NUMBER",
	DATA_PGP_KEY:                 "DATA_PGP_KEY",
	DATA_REGISTRY_NAME:           "DATA_REGISTRY_NAME",
	DATA_STRING:                  "DATA_STRING",
	DATA_TELEPHONE_OR_FAX_NUMBER: "DATA_TELEPHONE_OR_FAX_NUMBER",
	DATA_EXPORT_POLICY:           "DATA_EXPORT_POLICY",
	DATA_IMPORT_POLICY:           "DATA_IMPORT_POLICY",
	// Object Classes
	CLASS_AS_SET:      "CLASS_AS_SET",
	CLASS_AUT_NUM:     "CLASS_AUT_NUM",
	CLASS_DICTIONARY:  "CLASS_DICTIONARY",
	CLASS_FILTER_SET:  "CLASS_FILTER_SET",
	CLASS_MAINTAINER:  "CLASS_MAINTAINER",
	CLASS_PEERING_SET: "CLASS_PEERING_SET",
	CLASS_PERSON:      "CLASS_PERSON",
	CLASS_ROLE:        "CLASS_ROLE",
	CLASS_ROUTE6:      "CLASS_ROUTE6",
	CLASS_ROUTE:       "CLASS_ROUTE",
	CLASS_ROUTER:      "CLASS_ROUTER",
	CLASS_ROUTER_SET:  "CLASS_ROUTER_SET",
	CLASS_ROUTE_SET:   "CLASS_ROUTE_SET",
	// Object Attributes
	ATTR_ADDRESS:                 "ATTR_ADDRESS",
	ATTR_ADMIN_CONTACT:           "ATTR_ADMIN_CONTACT",
	ATTR_AS_NAME:                 "ATTR_AS_NAME",
	ATTR_AUTHENTICATION:          "ATTR_AUTHENTICATION",
	ATTR_CHANGED_AT_AND_BY:       "ATTR_CHANGED_AT_AND_BY",
	ATTR_CONTINUATION:            "ATTR_CONTINUATION",
	ATTR_DESCRIPTION:             "ATTR_DESCRIPTION",
	ATTR_EMAIL:                   "ATTR_EMAIL",
	ATTR_EXPORT:                  "ATTR_EXPORT",
	ATTR_FAX_NUMBER:              "ATTR_FAX_NUMBER",
	ATTR_IMPORT:                  "ATTR_IMPORT",
	ATTR_MAINTAINED_BY:           "ATTR_MAINTAINED_BY",
	ATTR_MAINTAINER_NOTIFY_EMAIL: "ATTR_MAINTAINER_NOTIFY_EMAIL",
	ATTR_NIC_HANDLE:              "ATTR_NIC_HANDLE",
	ATTR_NOTIFY_EMAIL:            "ATTR_NOTIFY_EMAIL",
	ATTR_PHONE_NUMBER:            "ATTR_PHONE_NUMBER",
	ATTR_REGISTRY_SOURCE:         "ATTR_REGISTRY_SOURCE",
	ATTR_REMARKS:                 "ATTR_REMARKS",
	ATTR_TECHNICAL_CONTACT:       "ATTR_TECHNICAL_CONTACT",
	ATTR_UPDATED_TO_EMAIL:        "ATTR_UPDATED_TO_EMAIL",
}

var objectStrings = map[Type]string{
	// Object Classes
	CLASS_AS_SET:      "as-set",
	CLASS_AUT_NUM:     "aut-num",
	CLASS_DICTIONARY:  "dictionary",
	CLASS_FILTER_SET:  "filter-set",
	CLASS_MAINTAINER:  "mntner",
	CLASS_PEERING_SET: "peering-set",
	CLASS_PERSON:      "person",
	CLASS_ROLE:        "role",
	CLASS_ROUTE6:      "route6",
	CLASS_ROUTE:       "route",
	CLASS_ROUTER:      "inet-rtr",
	CLASS_ROUTER_SET:  "rtr-set",
	CLASS_ROUTE_SET:   "route-set",
	// Object Attributes
	ATTR_ADDRESS:                 "address",
	ATTR_ADMIN_CONTACT:           "admin-c",
	ATTR_AS_NAME:                 "as-name",
	ATTR_AUTHENTICATION:          "auth",
	ATTR_CHANGED_AT_AND_BY:       "changed",
	ATTR_CONTINUATION:            "+",
	ATTR_DESCRIPTION:             "descr",
	ATTR_EMAIL:                   "e-mail",
	ATTR_EXPORT:                  "export",
	ATTR_FAX_NUMBER:              "fax-no",
	ATTR_IMPORT:                  "import",
	ATTR_MAINTAINED_BY:           "mnt-by",
	ATTR_MAINTAINER_NOTIFY_EMAIL: "mnt-nfy",
	ATTR_NIC_HANDLE:              "nic-hdl",
	ATTR_NOTIFY_EMAIL:            "notify",
	ATTR_PHONE_NUMBER:            "phone",
	ATTR_REGISTRY_SOURCE:         "source",
	ATTR_REMARKS:                 "remarks",
	ATTR_TECHNICAL_CONTACT:       "tech-c",
	ATTR_UPDATED_TO_EMAIL:        "upd-to",
	// Object Value Prefixes
	DATA_CRYPT_PASS:     "CRYPT-PW",
	DATA_MAIL_FROM_PASS: "MAIL-FROM",
	DATA_MD5_PASS:       "MD5-pw",
	DATA_NO_AUTH:        "NONE",
	DATA_PGP_KEY:        "PGPKey-",
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
		return strings.ToLower(tok)
	}

	panic(fmt.Sprintf("Unknown token type received: %d", t))
}
