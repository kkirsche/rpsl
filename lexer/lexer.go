package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kkirsche/rpsl/token"
	"github.com/mattn/go-runewidth"
)

const (
	alphaLowercase = "abcdefghijklmnopqrstuvwxyz"
	alphaUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alpha          = alphaLowercase + alphaUppercase
	digits         = "0123456789"
	hexDigits      = digits + "ABCDEFabcedf"
	alphaNumeric   = alpha + digits
	period         = "."
	hyphen         = "-"
	underscore     = "_"
	colon          = ":"
	plus           = "+"
	at             = "@"
	pound          = "#"
	forwardSlash   = "/"
	backSlash      = "\\"
	dollarSign     = "$"
	whitespace     = " \t\x85\xA0"
	newline        = "\n\v\f\r"
)

// Lexer is the structure responsible for managing the lexical scanning
// of the input text.
type Lexer struct {
	name          string           // name is the input text name, used for error reporting purposes
	input         string           // the input data / reader being scanned by the Lexer
	start         int              // the start position of the item currently being read
	pos           int              // the current read position in the input, we tokenize input[start:pos]
	lastRune      rune             // the last read rune
	lastRuneWidth int              // unicode has dynamic width characters, this tracks the width of the last read rune
	tokens        chan token.Token // channel of scanned tokens
	lineNum       int              // the current line number in the input text (based on line feeds)
	columnNum     int              // the current column number on the current line number
	// these allow us to backup
	previousCol  int
	previousLine int
}

// stateFn is a function representing the current lex state. A lex state may be
// in autonomous system set, in route object, find object type, etc. This is a
// recursive function which allows us to move between states without having to
// forget the state and rediscover it each time.
type stateFn func(*Lexer) stateFn

// eof constant is a magic constant allowing us to signal that we've hit the
// end of the file
const eof = -1

// Lex creates a new Lexer and starts running it
func Lex(inputName, inputText string) *Lexer {
	l := &Lexer{
		name:      inputName,
		input:     inputText,
		tokens:    make(chan token.Token, 2),
		columnNum: 1,
		lineNum:   1,
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
	// save our old state
	l.previousLine = l.lineNum
	l.previousCol = l.columnNum

	if l.lastRune == '\n' {
		l.columnNum = 1
		l.lineNum++
	}
	// if the current read position is farther than the length of the input,
	// we've hit the end of the file.
	if l.pos >= len(l.input) {
		l.lastRuneWidth = 0
		fmt.Println("EOF")
		return eof
	}

	l.lastRune, l.lastRuneWidth = utf8.DecodeRuneInString(l.input[l.pos:])
	l.columnNum = l.columnNum + runewidth.RuneWidth(l.lastRune)
	l.pos += l.lastRuneWidth
	return l.lastRune
}

// backup is responsible for moving our read position in the input back one
// unicode character (based on the lastRuneWidth)
func (l *Lexer) backup() {
	l.pos -= l.lastRuneWidth
	l.columnNum = l.previousCol
	l.lineNum = l.previousLine
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
	l.tokens <- newToken(t, string(l.input[l.start:l.pos]), l.columnNum, l.lineNum)
	l.start = l.pos
}

// newToken is a helper to generate a new token
func newToken(tType token.Type, literal string, column, line int) token.Token {
	if tType == token.EOF {
		column = 0
		line = 0
	}
	return token.Token{Type: tType, Literal: literal, Column: column, Line: line}
}

// ignore skips over the pending input before this point
func (l *Lexer) ignore() {
	l.start = l.pos
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.readRune()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptExcept(invalid string) bool {
	if strings.IndexRune(invalid, l.readRune()) == -1 {
		return true
	}
	l.backup()
	return false
}

// acceptRun is used to read in any as many of the valid characters as possible
func (l *Lexer) acceptRun(valid string) bool {
	accepted := false
	for l.accept(valid) == true {
		accepted = true
	}

	return accepted
}

// acceptExceptRun is used to read in as many characters which do not match the
// invalid string as possible
func (l *Lexer) acceptExceptRun(invalid string) bool {
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
		case strings.HasPrefix(l.input[l.pos:], token.MAINTAINER.Name()):
			return lexMaintainerClassName(l)
		case strings.HasPrefix(l.input[l.pos:], token.PERSON.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.ROLE.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.AUT_NUM.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.AS_SET.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.ROUTE.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.ROUTE6.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.ROUTE_SET.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.FILTER_SET.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.ROUTER.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.ROUTER_SET.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.PEERING_SET.Name()):
			fallthrough
		case strings.HasPrefix(l.input[l.pos:], token.DICTIONARY.Name()):
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

func lexMaintainerClassName(l *Lexer) stateFn {
	l.pos += len(token.MAINTAINER.Name())
	l.columnNum = len(token.MAINTAINER.Name())
	l.emit(token.MAINTAINER)

	if !l.accept(":") {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.acceptRun(whitespace)
	// ignore the colon and any whitespace following it
	l.ignore()

	return lexMaintainerClassNameValue
}

func lexMaintainerClassNameValue(l *Lexer) stateFn {
	// object names must start with a letter
	if !l.accept(alpha) {
		l.emit(token.ILLEGAL)
		return nil
	}

	l.acceptRun(alphaNumeric + hyphen + underscore)
	if l.pos > l.start {
		l.emit(token.STRING)
	}

	if !l.acceptRun(newline) {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.ignore()

	return lexMaintainerAttributes(l)
}

func lexMaintainerAttributes(l *Lexer) stateFn {
	switch {
	case strings.HasPrefix(l.input[l.pos:], token.DESCRIPTION.Name()):
		return lexDescriptionAttrName(l, lexMaintainerAttributes)
	case strings.HasPrefix(l.input[l.pos:], token.AUTHENTICATION.Name()):
		return lexAuthenticationAttrName(l, lexMaintainerAttributes)
	case strings.HasPrefix(l.input[l.pos:], token.UPDATED_TO_EMAIL.Name()):
		return lexUpdatedToEmailAttrName(l, lexMaintainerAttributes)
	case strings.HasPrefix(l.input[l.pos:], token.MAINTAINER_NOTIFY_EMAIL.Name()):
		return lexMaintainerNotifyEmailAttrName(l, lexMaintainerAttributes)
	case strings.HasPrefix(l.input[l.pos:], token.TECHNICAL_CONTACT.Name()):
		// TODO
		return lexObjectClass(l)
	case strings.HasPrefix(l.input[l.pos:], token.ADMIN_CONTACT.Name()):
		return lexAdminContactAttrName(l, lexMaintainerAttributes)
	case strings.HasPrefix(l.input[l.pos:], token.REMARKS.Name()):
		// TODO
		return lexObjectClass(l)
	case strings.HasPrefix(l.input[l.pos:], token.NOTIFY_EMAIL.Name()):
		return lexNotifyEmailAttrName(l, lexMaintainerAttributes)
	case strings.HasPrefix(l.input[l.pos:], token.MAINTAINED_BY.Name()):
		// TODO
		return lexObjectClass(l)
	case strings.HasPrefix(l.input[l.pos:], token.CHANGED_AT_AND_BY.Name()):
		// TODO
		return lexObjectClass(l)
	case strings.HasPrefix(l.input[l.pos:], token.RECORD_SOURCE.Name()):
		// TODO
		return lexObjectClass(l)
	default:
		return lexObjectClass(l)
	}
}

func lexDescriptionAttrName(l *Lexer, returnToStateFn stateFn) stateFn {
	l.pos += len(token.DESCRIPTION.Name())
	l.columnNum = len(token.DESCRIPTION.Name())
	l.emit(token.DESCRIPTION)

	if !l.accept(":") {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.acceptRun(whitespace)
	// ignore the colon and any whitespace following it
	l.ignore()

	return lexFreeformAttrValue(l, returnToStateFn)
}

func lexAdminContactAttrName(l *Lexer, returnToStateFn stateFn) stateFn {
	l.pos += len(token.ADMIN_CONTACT.Name())
	l.columnNum = len(token.ADMIN_CONTACT.Name())
	l.emit(token.ADMIN_CONTACT)

	if !l.accept(":") {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.acceptRun(whitespace)
	// ignore the colon and any whitespace following it
	l.ignore()

	return lexNICHandleAttrValue(l, lexMaintainerAttributes)
}

func lexNotifyEmailAttrName(l *Lexer, returnToStateFn stateFn) stateFn {
	l.pos += len(token.NOTIFY_EMAIL.Name())
	l.columnNum = len(token.NOTIFY_EMAIL.Name())
	l.emit(token.NOTIFY_EMAIL)

	if !l.accept(":") {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.acceptRun(whitespace)
	// ignore the colon and any whitespace following it
	l.ignore()

	return lexEmailAttrValue(l, returnToStateFn)
}

func lexUpdatedToEmailAttrName(l *Lexer, returnToStateFn stateFn) stateFn {
	l.pos += len(token.UPDATED_TO_EMAIL.Name())
	l.columnNum = len(token.UPDATED_TO_EMAIL.Name())
	l.emit(token.UPDATED_TO_EMAIL)

	if !l.accept(":") {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.acceptRun(whitespace)
	// ignore the colon and any whitespace following it
	l.ignore()

	return lexEmailAttrValue(l, returnToStateFn)
}

func lexMaintainerNotifyEmailAttrName(l *Lexer, returnToStateFn stateFn) stateFn {
	l.pos += len(token.MAINTAINER_NOTIFY_EMAIL.Name())
	l.columnNum = len(token.MAINTAINER_NOTIFY_EMAIL.Name())
	l.emit(token.MAINTAINER_NOTIFY_EMAIL)

	if !l.accept(":") {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.acceptRun(whitespace)
	// ignore the colon and any whitespace following it
	l.ignore()

	return lexEmailAttrValue(l, returnToStateFn)
}

func lexAuthenticationAttrName(l *Lexer, returnToStateFn stateFn) stateFn {
	l.pos += len(token.AUTHENTICATION.Name())
	l.columnNum = len(token.AUTHENTICATION.Name())
	l.emit(token.AUTHENTICATION)

	if !l.accept(":") {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.acceptRun(whitespace)
	// ignore the colon and any whitespace following it
	l.ignore()

	return lexAuthenticationAttrValue(l, returnToStateFn)
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
		l.emit(token.STRING)
	}

	if !l.acceptRun(newline) {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.ignore()

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
		l.emit(token.EMAIL)
	}

	if !l.acceptRun(newline) {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.ignore()

	return nextStateFn
}

func lexFreeformAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	l.acceptExceptRun(newline)
	if l.pos > l.start {
		l.emit(token.STRING)
	}

	if !l.acceptRun(newline) {
		l.emit(token.ILLEGAL)
		return nil
	}
	l.ignore()

	return nextStateFn
}

func lexAuthenticationAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	switch {
	case strings.HasPrefix(l.input[l.pos:], token.PGP_KEY.Name()):
		l.pos += len(token.PGP_KEY.Name())
		l.columnNum += runewidth.StringWidth(token.PGP_KEY.Name())
		l.ignore()
		return lexPGPKeyAuthAttrValue(l, nextStateFn)
	case strings.HasPrefix(l.input[l.pos:], token.CRYPT_PASS.Name()):
		l.pos += len(token.CRYPT_PASS.Name())
		l.columnNum += runewidth.StringWidth(token.CRYPT_PASS.Name())
		l.acceptRun(whitespace)
		l.ignore()
		return lexCryptPassAuthAttrValue(l, nextStateFn)
	case strings.HasPrefix(l.input[l.pos:], token.MD5_PASS.Name()):
		l.pos += len(token.MD5_PASS.Name())
		l.columnNum += runewidth.StringWidth(token.MD5_PASS.Name())
		l.acceptRun(whitespace)
		l.ignore()
		return lexMD5PassAuthAttrValue(l, nextStateFn)
	case strings.HasPrefix(l.input[l.pos:], token.MAIL_FROM_PASS.Name()):
		fmt.Println("eating MAIL-FROM", l.columnNum)
		l.pos += len(token.MAIL_FROM_PASS.Name())
		l.columnNum += runewidth.StringWidth(token.MAIL_FROM_PASS.Name())
		l.acceptRun(whitespace)
		l.ignore()
		fmt.Println("entering email address", l.columnNum)
		return lexMailFromPassAuthAttrValue(l, nextStateFn)
	case strings.HasPrefix(l.input[l.pos:], token.NO_AUTH.Name()):
		return lexNoAuthAttrValue(l, nextStateFn)
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
		l.emit(token.PGP_KEY)
	}

	l.acceptExceptRun(newline)
	l.acceptRun(newline)
	l.ignore()
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
		l.emit(token.CRYPT_PASS)
	}

	l.acceptExceptRun(newline)
	l.acceptRun(newline)
	l.ignore()
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
		l.emit(token.MD5_PASS)
	}

	l.acceptExceptRun(newline)
	l.acceptRun(newline)
	l.ignore()
	return nextStateFn
}

func lexMailFromPassAuthAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	fmt.Println("beginning email address", l.columnNum)
	if !l.acceptExceptRun(whitespace + newline + at) {
		l.emit(token.ILLEGAL)
		return nil
	}

	fmt.Println("eating @", l.columnNum)
	if !l.accept(at) {
		l.emit(token.ILLEGAL)
		return nil
	}

	fmt.Println("eating domain", l.columnNum)
	if !l.acceptRun(alphaNumeric + period + hyphen + underscore + colon) {
		l.emit(token.ILLEGAL)
		return nil
	}

	fmt.Println("emiting token", l.columnNum)
	if l.pos > l.start {
		l.emit(token.MAIL_FROM_PASS)
	}

	l.acceptRun(whitespace)
	l.accept(pound)
	l.acceptExceptRun(newline)
	l.acceptRun(newline)
	l.ignore()

	return nextStateFn
}

func lexNoAuthAttrValue(l *Lexer, nextStateFn stateFn) stateFn {
	return nil
}
