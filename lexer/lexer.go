package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kkirsche/rpsl/token"
)

const (
	alpha        = "abcdefghijklmnopqrstuvwxyz"
	digits       = "0123456789"
	hexDigits    = digits + "ABCDEFabcedf"
	alphaNumeric = alpha + digits
	period       = "."
	hyphen       = "-"
	underscore   = "_"
	colon        = ":"
	semicolon    = ";"
	equal        = "="
	plus         = "+"
	at           = "@"
	pound        = "#"
	forwardSlash = "/"
	backSlash    = "\\"
	dollarSign   = "$"
	whitespace   = " \t\x85\xA0"
	newline      = "\n\v\f\r"
	comma        = ","
)

// Lexer is the structure responsible for managing the lexical scanning
// of the input text.
type Lexer struct {
	name             string           // name is the input text name, used for error reporting purposes
	input            string           // the input data / reader being scanned by the Lexer
	start            int              // the start position of the item currently being read
	pos              int              // the current read position in the input, we tokenize input[start:pos]
	lastRune         rune             // the last read rune
	lastRuneWidth    int              // unicode has dynamic width characters, this tracks the width of the last read rune
	tokens           chan token.Token // channel of scanned tokens
	lineNum          int              // the current line number in the input text (based on line feeds)
	continuationPath func(*Lexer, stateFn) stateFn
}

// stateFn is a function representing the current lex state. A lex state may be
// in autonomous system set, in route object, find object type, etc. This is a
// recursive function which allows us to move between states without having to
// forget the state and rediscover it each time.
type stateFn func(*Lexer) stateFn
type stagedStateFn func(*Lexer, stateFn) stateFn

// eof constant is a magic constant allowing us to signal that we've hit the
// end of the file
const eof = -1

// Lex creates a new Lexer and starts running it
func Lex(inputName, inputText string) *Lexer {
	l := &Lexer{
		name:    inputName,
		input:   inputText,
		tokens:  make(chan token.Token, 2),
		lineNum: 1,
	}

	go l.run()

	return l
}

// run is responsible for executing the lexical scanner state machine
func (l *Lexer) run() {
	// lexObjectClass is the default state machine. The first thing we do with an
	// object in RPSL is try to determine what type of object it is.
	for state := lexObjectClass; state != nil; {
		state = state(l)
	}

	close(l.tokens)
}

//*============================================================================
// Helper Functions
//*============================================================================

// NextToken returns the next token stored in the channel
func (l *Lexer) NextToken() token.Token {
	return <-l.tokens
}

// readRune returns the next rune in the input
func (l *Lexer) readRune() rune {
	// if the current read position is farther than the length of the input,
	// we've hit the end of the file.
	if l.pos >= len(l.input) {
		l.lastRuneWidth = 0
		fmt.Println("EOF")
		return eof
	}

	l.lastRune, l.lastRuneWidth = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.lastRuneWidth
	if l.lastRune == '\n' {
		l.lineNum++
	}
	return l.lastRune
}

// backup is responsible for moving our read position in the input back one
// unicode character (based on the lastRuneWidth)
func (l *Lexer) backup() {
	l.pos -= l.lastRuneWidth
	// taken from https://golang.org/src/text/template/parse/lex.go
	if l.lastRuneWidth == 1 && l.input[l.pos] == '\n' {
		l.lineNum--
	}
}

// peek looks up the next rune in the input but does not advance our position
func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return eof
	}

	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

// emit is used to send the current token to the output channel. As it's a
// buffered channel we can store the next token and the peek token
func (l *Lexer) emit(t token.Type) {
	l.tokens <- newToken(t, string(l.input[l.start:l.pos]), l.lineNum)
	l.start = l.pos
}

// newToken is a helper to generate a new token
func newToken(tType token.Type, literal string, line int) token.Token {
	if tType == token.EOF {
		line = 0
	}
	return token.Token{Type: tType, Literal: literal, Line: line}
}

// ignore skips over the pending input before this point
func (l *Lexer) ignore() {
	l.start = l.pos
}

func (l *Lexer) accept(valid string) bool {
	valid = strings.ToLower(valid) + strings.ToUpper(valid)
	if strings.IndexRune(valid, l.readRune()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptExcept(invalid string) bool {
	invalid = strings.ToLower(invalid) + strings.ToUpper(invalid)
	if strings.IndexRune(invalid, l.readRune()) == -1 {
		return true
	}
	l.backup()
	return false
}

// acceptRun is used to read in any as many of the valid characters as possible
func (l *Lexer) acceptRun(valid string) bool {
	valid = strings.ToLower(valid) + strings.ToUpper(valid)
	accepted := false
	for l.accept(valid) == true {
		accepted = true
	}

	return accepted
}

// acceptExceptRun is used to read in as many characters which do not match the
// invalid string as possible
func (l *Lexer) acceptExceptRun(invalid string) bool {
	invalid = strings.ToLower(invalid) + strings.ToUpper(invalid)
	accepted := false
	for l.acceptExcept(invalid) == true {
		accepted = true
	}

	return accepted
}

//*============================================================================
// State Functions
//*============================================================================

// lexObjectClass is used to determine what class of RPSL object we are on
func lexObjectClass(l *Lexer) stateFn {
	for {
		switch {
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_MAINTAINER.Name()):
			return lexAttrName(l, token.CLASS_MAINTAINER, lexNICHandleAttrValue, lexClassAttributes)
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_PERSON.Name()):
			return lexAttrName(l, token.CLASS_PERSON, lexFreeformAttrValue, lexClassAttributes)
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_ROLE.Name()):
			return lexAttrName(l, token.CLASS_ROLE, lexFreeformAttrValue, lexClassAttributes)
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_AUT_NUM.Name()):
			return lexAttrName(l, token.CLASS_AUT_NUM, lexAutNumAttrValue, lexClassAttributes)
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_AS_SET.Name()):
			// TODO
			fallthrough
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_ROUTE.Name()):
			// TODO
			fallthrough
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_ROUTE6.Name()):
			// TODO
			fallthrough
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_ROUTE_SET.Name()):
			// TODO
			fallthrough
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_FILTER_SET.Name()):
			// TODO
			fallthrough
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_ROUTER.Name()):
			// TODO
			fallthrough
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_ROUTER_SET.Name()):
			// TODO
			fallthrough
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_PEERING_SET.Name()):
			// TODO
			fallthrough
		case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.CLASS_DICTIONARY.Name()):
			// TODO
			fallthrough
		default:
			l.emit(token.EOF)
			return nil
		}

		// if l.readRune() == eof {
		// 	break
		// }
	}

	// if l.pos > l.start {
	// 	l.emit(token.STRING)
	// }
	// fmt.Println("emitting EOF")
	// l.emit(token.EOF)

	// // return nil to stop the state machine
	// return nil
}

func lexClassAttributes(l *Lexer) stateFn {
	l.acceptRun(whitespace)
	if l.accept(pound) {
		l.acceptExceptRun(newline)
		l.acceptRun(newline)
	}
	l.accept(newline)
	l.ignore()

	switch {
	case strings.HasPrefix(l.input[l.pos:], " "):
		// per the RFC:
		// An attribute's value can be split over multiple lines, by having a
		// space, a tab or a plus ('+') character as the first character of the
		// continuation lines.  The character "+" for line continuation allows
		// attribute values to contain blank lines.  More spaces may optionally
		// be used after the continuation character to increase readability.
		fallthrough
	case strings.HasPrefix(l.input[l.pos:], "\t"):
		fallthrough
	case strings.HasPrefix(l.input[l.pos:], token.ATTR_CONTINUATION.Name()):
		// don't use lexAttrName as we expect a : in that, we don't use one in a
		// continuation circumstance.
		// the use of the + is fine here, as a space, tab or plus is a single character
		// any extra get consumed by the acceptRun whitespace piece
		l.pos += len(token.ATTR_CONTINUATION.Name())
		l.emit(token.ATTR_CONTINUATION)
		l.acceptRun(whitespace)
		l.ignore()
		return l.continuationPath(l, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_AS_NAME.Name()):
		return lexAttrName(l, token.ATTR_AS_NAME, lexNICHandleAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_DESCRIPTION.Name()):
		return lexAttrName(l, token.ATTR_DESCRIPTION, lexFreeformAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_AUTHENTICATION.Name()):
		return lexAttrName(l, token.ATTR_AUTHENTICATION, lexAuthenticationAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_UPDATED_TO_EMAIL.Name()):
		return lexAttrName(l, token.ATTR_UPDATED_TO_EMAIL, lexEmailAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_MAINTAINER_NOTIFY_EMAIL.Name()):
		return lexAttrName(l, token.ATTR_MAINTAINER_NOTIFY_EMAIL, lexEmailAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_TECHNICAL_CONTACT.Name()):
		return lexAttrName(l, token.ATTR_TECHNICAL_CONTACT, lexNICHandleAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_ADMIN_CONTACT.Name()):
		return lexAttrName(l, token.ATTR_ADMIN_CONTACT, lexNICHandleAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_REMARKS.Name()):
		return lexAttrName(l, token.ATTR_REMARKS, lexFreeformAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_NOTIFY_EMAIL.Name()):
		return lexAttrName(l, token.ATTR_NOTIFY_EMAIL, lexEmailAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_MAINTAINED_BY.Name()):
		return lexAttrName(l, token.ATTR_MAINTAINED_BY, lexNICHandleAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_CHANGED_AT_AND_BY.Name()):
		return lexAttrName(l, token.ATTR_CHANGED_AT_AND_BY, lexEmailAndDateAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_REGISTRY_SOURCE.Name()):
		return lexAttrName(l, token.ATTR_REGISTRY_SOURCE, lexRegistrySourceAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_NIC_HANDLE.Name()):
		return lexAttrName(l, token.ATTR_NIC_HANDLE, lexNICHandleAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_ADDRESS.Name()):
		return lexAttrName(l, token.ATTR_ADDRESS, lexFreeformAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_PHONE_NUMBER.Name()):
		return lexAttrName(l, token.ATTR_PHONE_NUMBER, lexPhoneOrFaxAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_FAX_NUMBER.Name()):
		return lexAttrName(l, token.ATTR_FAX_NUMBER, lexPhoneOrFaxAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_EMAIL.Name()):
		return lexAttrName(l, token.ATTR_EMAIL, lexEmailAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_EXPORT.Name()):
		return lexAttrName(l, token.ATTR_EXPORT, lexExportAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_IMPORT.Name()):
		return lexAttrName(l, token.ATTR_IMPORT, lexImportAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_MULTI_PROTO_EXPORT_POLICY.Name()):
		return lexAttrName(l, token.ATTR_MULTI_PROTO_EXPORT_POLICY, lexMultiProtoExportAttrValue, lexClassAttributes)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.ATTR_MULTI_PROTO_IMPORT_POLICY.Name()):
		return lexAttrName(l, token.ATTR_MULTI_PROTO_IMPORT_POLICY, lexMultiProtoImportAttrValue, lexClassAttributes)
	default:
		return lexObjectClass(l)
	}
}

func lexAttrName(l *Lexer, tokenType token.Type, valueStateFn stagedStateFn, returnToStateFn stateFn) stateFn {
	l.pos += len(tokenType.Name())
	l.emit(tokenType)

	if !l.accept(":") {
		l.emit(token.ILLEGAL)
		return nil
	}

	l.acceptRun(whitespace)
	// ignore the colon and any whitespace following it
	l.ignore()
	l.continuationPath = valueStateFn
	return valueStateFn(l, returnToStateFn)
}

func lexNICHandleAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	// NIC handles (Network Information Centre handles) are alphanumeric
	// object names must start with a letter
	// and must begin with a letter
	if !l.accept(alpha) {
		l.emit(token.ILLEGAL)
		return nil
	}

	l.acceptRun(alphaNumeric + hyphen + underscore)
	if l.pos > l.start {
		l.emit(token.DATA_NIC_HANDLE)
	}

	l.acceptRun(whitespace)
	l.ignore()

	foundNextHandle := l.accept(comma)
	for foundNextHandle == true {
		l.acceptRun(whitespace)
		l.ignore()

		if !l.accept(alpha) {
			l.emit(token.ILLEGAL)
			return nil
		}

		l.acceptRun(alphaNumeric + hyphen + underscore)
		if l.pos > l.start {
			l.emit(token.DATA_NIC_HANDLE)
		}

		l.acceptRun(whitespace)
		l.ignore()
		foundNextHandle = l.accept(comma)
	}

	return nextStateFn
}

func lexEmailAndDateAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	// ignore the state function aspect of the email attribute reader
	_ = lexEmailAttrValue(l, nil)
	// accept any whitespace between the email and the date
	l.acceptRun(whitespace)
	l.ignore()

	// read in the 8 digit date
	// YYYY = 4 numbers for year
	// MM = 2 digits for month
	// DD = 2 digits for day
	for i := 0; i < 8; i++ {
		if !l.accept(digits) {
			l.emit(token.ILLEGAL)
			return nil
		}
	}

	if l.pos > l.start {
		l.emit(token.DATA_DATE)
	}

	return nextStateFn
}

func lexEmailAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	// Maintainer Notify Email is an email address
	// it's the parser's responsibility to use net/mail.ParseAddress
	// to validate that we lexed it correctly
	if !l.acceptExceptRun(whitespace + newline + at) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if !l.accept(at) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if !l.acceptRun(alphaNumeric + period + hyphen + underscore + colon) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if l.pos > l.start {
		l.emit(token.DATA_EMAIL)
	}

	return nextStateFn
}

func lexFreeformAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	l.acceptExceptRun(newline)
	if l.pos > l.start {
		l.emit(token.DATA_STRING)
	}

	return nextStateFn
}

func lexRegistrySourceAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	l.acceptExceptRun(newline)
	if l.pos > l.start {
		l.emit(token.DATA_REGISTRY_NAME)
	}

	return nextStateFn
}

func lexAuthenticationAttrValue(l *Lexer, returnToStateFn stateFn) stateFn {
	switch {
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.DATA_PGP_KEY.Name()):
		l.pos += len(token.DATA_PGP_KEY.Name())
		l.ignore()
		return lexPGPKeyAuthAttrValue(l, returnToStateFn)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.DATA_CRYPT_PASS.Name()):
		l.pos += len(token.DATA_CRYPT_PASS.Name())
		l.acceptRun(whitespace)
		l.ignore()
		return lexCryptPassAuthAttrValue(l, returnToStateFn)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.DATA_MD5_PASS.Name()):
		l.pos += len(token.DATA_MD5_PASS.Name())
		l.acceptRun(whitespace)
		l.ignore()
		return lexMD5PassAuthAttrValue(l, returnToStateFn)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.DATA_MAIL_FROM_PASS.Name()):
		l.pos += len(token.DATA_MAIL_FROM_PASS.Name())
		l.acceptRun(whitespace)
		l.ignore()
		return lexMailFromPassAuthAttrValue(l, returnToStateFn)
	case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), token.DATA_NO_AUTH.Name()):
		return lexNoAuthAttrValue(l, returnToStateFn)
	default:
		l.emit(token.ILLEGAL)
		return nil
	}
}

func lexPGPKeyAuthAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	// accept hex numbers representing the PGP key
	for i := 0; i < 8; i++ {
		if !l.accept(hexDigits) {
			l.emit(token.ILLEGAL)
			return nil
		}
	}

	if l.pos > l.start {
		l.emit(token.DATA_PGP_KEY)
	}

	return nextStateFn
}

func lexCryptPassAuthAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	// you can generate a crypt password via:
	// openssl passwd -crypt MyPassword
	for i := 0; i < 13; i++ {
		if !l.accept(alphaNumeric + forwardSlash + backSlash) {
			l.emit(token.ILLEGAL)
			return nil
		}
	}

	if l.pos > l.start {
		l.emit(token.DATA_CRYPT_PASS)
	}

	return nextStateFn
}

func lexMD5PassAuthAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	if !l.accept(dollarSign) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if !l.accept("1") {
		l.emit(token.ILLEGAL)
		return nil
	}

	if !l.accept(dollarSign) {
		l.emit(token.ILLEGAL)
		return nil
	}

	for i := 0; i < 8; i++ {
		if !l.accept(alphaNumeric + period + forwardSlash) {
			l.emit(token.ILLEGAL)
			return nil
		}
	}

	if !l.accept(dollarSign) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if !l.acceptRun(alphaNumeric + period + forwardSlash) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if l.pos > l.start {
		l.emit(token.DATA_MD5_PASS)
	}

	return nextStateFn
}

func lexMailFromPassAuthAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	if !l.acceptExceptRun(whitespace + newline + at) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if !l.accept(at) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if !l.acceptRun(alphaNumeric + period + hyphen + underscore + colon) {
		l.emit(token.ILLEGAL)
		return nil
	}

	if l.pos > l.start {
		l.emit(token.DATA_MAIL_FROM_PASS)
	}

	return nextStateFn
}

func lexNoAuthAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	l.pos += len(token.DATA_NO_AUTH.Name())

	if l.pos > l.start {
		l.emit(token.DATA_NO_AUTH)
	}

	return nextStateFn
}

func lexPhoneOrFaxAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	// The phone and the fax-no attributes have the following syntax:
	//    phone: +<country-code> <area code> <city> <subscriber> [ext. <extension>]

	if !l.accept(plus) {
		l.emit(token.ILLEGAL)
	}

	// country code
	l.acceptRun(digits)
	l.acceptRun(whitespace)

	// area code
	l.acceptRun(digits)
	l.acceptRun(whitespace)

	// city
	l.acceptRun(digits)
	l.acceptRun(whitespace)

	// subscriber
	l.acceptRun(digits)
	l.acceptRun(whitespace)

	// check if we find the e from ext.
	if l.accept("e") {
		// read in the remaining xt.
		ext := []string{"x", "t", "."}
		for _, x := range ext {
			if !l.accept(x) {
				l.emit(token.ILLEGAL)
			}
		}

		// extension
		l.acceptRun(digits)
	}

	if l.pos > l.start {
		l.emit(token.DATA_TELEPHONE_OR_FAX_NUMBER)
	}

	return nextStateFn
}

func lexExportAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	// lex [protocol <protocol-1>] [into <protocol-2>]
	succeeded := partialLexProtocol(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexToPeer(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAction(l, "announce")
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAnnouncement(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	if l.pos > l.start {
		l.emit(token.DATA_EXPORT_POLICY)
	}

	return nextStateFn
}

func lexAutNumAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	for _, t := range "AS" {
		if !l.accept(string(t)) {
			l.emit(token.ILLEGAL)
			return nil
		}
	}

	l.acceptRun(digits)
	if l.pos > l.start {
		l.emit(token.DATA_ASN)
	}

	return nextStateFn
}

func lexImportAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	// lex [protocol <protocol-1>] [into <protocol-2>]
	succeeded := partialLexProtocol(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexFromPeer(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAction(l, "accept")
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAcceptAS(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	if l.pos > l.start {
		l.emit(token.DATA_IMPORT_POLICY)
	}

	return nextStateFn
}

func lexMultiProtoExportAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	succeeded := partialLexProtocol(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAddressFamilyIdentifier(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexToPeer(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAction(l, "announce")
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAnnouncement(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	if l.pos > l.start {
		l.emit(token.DATA_MULTI_PROTO_EXPORT_POLICY)
	}

	return nextStateFn
}

func lexMultiProtoImportAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	succeeded := partialLexProtocol(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAddressFamilyIdentifier(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexFromPeer(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAction(l, "accept")
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	succeeded = partialLexAcceptAS(l)
	if !succeeded {
		l.emit(token.ILLEGAL)
		return nil
	}

	if l.pos > l.start {
		l.emit(token.DATA_MULTI_PROTO_IMPORT_POLICY)
	}

	return nextStateFn
}

func partialLexProtocol(l *Lexer) bool {
	// if the export policy begins with protocol
	if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "protocol") {
		for _, t := range "protocol" {
			if !l.accept(string(t)) {
				return false
			}
		}
		l.acceptRun(whitespace)

		// protocol name 1
		// e.g. BGP, BGP4, OSPF, STATIC, RIP, IS-IS, etc.
		l.acceptRun(alphaNumeric + hyphen)
		l.acceptRun(whitespace)
	}

	if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "into") {
		for _, t := range "into" {
			if !l.accept(string(t)) {
				return false
			}
		}
		l.acceptRun(whitespace)

		// protocol name 2
		// e.g. BGP, BGP4, OSPF, STATIC, RIP, IS-IS, etc.
		l.acceptRun(alphaNumeric + hyphen)
		l.acceptRun(whitespace)
	}

	l.acceptRun(whitespace)
	return true
}

func partialLexAddressFamilyIdentifier(l *Lexer) bool {
	// if the export policy begins with protocol
	if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "afi") {
		for _, t := range "afi" {
			if !l.accept(string(t)) {
				return false
			}
		}
		l.acceptRun(whitespace)

		// afi value - any.unicast, ipv6.unicast, etc.
		for tokenizingAFI := true; tokenizingAFI == true; {
			afiType := ""
			// AFI type list
			// https://tools.ietf.org/html/rfc4012#section-2.2
			switch {
			case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "ipv4.unicast"):
				afiType = "ipv4.unicast"
			case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "ipv4.multicast"):
				afiType = "ipv4.multicast"
			case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "ipv4"):
				afiType = "ipv4"
			case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "ipv6.unicast"):
				afiType = "ipv6.unicast"
			case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "ipv6.multicast"):
				afiType = "ipv6.multicast"
			case strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "ipv6"):
				afiType = "ipv6"
			default:
				// illegal AFI
				return false
			}

			for _, t := range afiType {
				if !l.accept(string(t)) {
					return false
				}
			}

			l.acceptRun(whitespace)
			if !l.accept(comma) {
				tokenizingAFI = false
			}
		}
	}
	l.acceptRun(whitespace)
	return true
}

func partialLexToPeer(l *Lexer) bool {
	// to peer / mp-peer is a required attribute
	for _, t := range "to" {
		if !l.accept(string(t)) {
			return false
		}
	}
	l.acceptRun(whitespace)

	for _, t := range "AS" {
		if !l.accept(string(t)) {
			return false
		}
	}
	l.acceptRun(digits)
	l.acceptRun(whitespace)
	return true
}

func partialLexFromPeer(l *Lexer) bool {
	// from peer / mp-peer is a required attribute
	for _, t := range "from" {
		if !l.accept(string(t)) {
			return false
		}
	}
	l.acceptRun(whitespace)

	for _, t := range "AS" {
		if !l.accept(string(t)) {
			return false
		}
	}
	if l.accept(hyphen) {
		if !l.acceptRun(alphaNumeric) {
			return false
		}
	} else {
		if !l.acceptRun(digits) {
			return false
		}
	}

	l.acceptRun(whitespace)
	return true
}

func partialLexAction(l *Lexer, nextClause string) bool {
	if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "action") {
		for _, t := range "action" {
			if !l.accept(string(t)) {
				return false
			}
		}

		for parsingAction := true; parsingAction == true; {
			// action name
			if !l.acceptRun(alphaNumeric + hyphen + period) {
				return false
			}
			l.acceptRun(whitespace)
			// assignment
			if !l.accept(equal) {
				return false
			}
			l.acceptRun(whitespace)
			// action value
			if !l.acceptRun(alphaNumeric + hyphen + period) {
				return false
			}
			// action end
			if !l.accept(semicolon) {
				return false
			}
			l.acceptRun(whitespace)
			// check for announce clause instead of another action
			if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), nextClause) {
				parsingAction = false
			}
		}
	}

	l.acceptRun(whitespace)
	return true
}

func partialLexAnnouncement(l *Lexer) bool {
	for _, t := range "announce" {
		if !l.accept(string(t)) {
			return false
		}
	}
	l.acceptRun(whitespace)

	for parsingAS := true; parsingAS == true; {
		for _, t := range "AS" {
			if !l.accept(string(t)) {
				return false
			}
		}

		if l.accept(hyphen) {
			l.acceptRun(alphaNumeric + underscore)
		} else {
			l.acceptRun(digits)
		}

		l.acceptRun(whitespace)
		// check for another AS instead of the end of the line
		if !strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), "AS") {
			parsingAS = false
		}
	}

	l.acceptRun(whitespace)
	return true
}

func partialLexAcceptAS(l *Lexer) bool {
	for _, t := range "accept" {
		if !l.accept(string(t)) {
			return false
		}
	}
	l.acceptRun(whitespace)

	if strings.HasPrefix(strings.ToLower(l.input[l.pos:]), "any") {
		for _, t := range "ANY" {
			if !l.accept(string(t)) {
				return false
			}
		}
	} else {
		for parsingAS := true; parsingAS == true; {
			for _, t := range "AS" {
				if !l.accept(string(t)) {
					return false
				}
			}

			if l.accept(hyphen) {
				l.acceptRun(alphaNumeric + underscore + hyphen)
			} else {
				l.acceptRun(digits)
			}

			l.acceptRun(whitespace)
			// check for another AS instead of the end of the line
			if !strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), "AS") {
				parsingAS = false
			}
		}
	}

	l.acceptRun(whitespace)
	return true
}
