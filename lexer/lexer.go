package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kkirsche/rpsl/token"
)

// Lexer is the structure responsible for converting the input text into a
// series of tokens
type Lexer struct {
	input        string
	inputLen     int
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           rune // current char under examination
	chWidth      int  // current character's width
	column       int  // the current number of the column of the line
	line         int  // the current number of line
}

// New is used to create a new lexer instance from the input text
func New(input string) *Lexer {
	l := &Lexer{
		input:    input,
		inputLen: len(input),
		line:     1,
	}
	// this ensures that our position, readPosition, and column number are
	// all initialized before the caller uses the lexer
	l.readChar()
	return l
}

func newToken(tType token.Type, ch rune, column, line int) token.Token {
	return token.Token{Type: tType, Literal: string(ch), Column: column, Line: line}
}

func isLetter(ch rune, strict bool) bool {
	if strict {
		return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
	}

	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '-'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isAlphaNumeric(ch rune, strict bool) bool {
	return isDigit(ch) || isLetter(ch, strict)
}

func isRune(ch, match rune) bool {
	return ch == match
}

// readChar advances our position in the input string and provides us with the
// next character. If we have reached the end of the string, we set the
// character to 0, which is the ASCII "NUL" code.
func (l *Lexer) readChar() {
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}

	l.chWidth = 1
	if l.readPosition >= l.inputLen {
		l.ch = 0
	} else {
		l.ch, l.chWidth = utf8.DecodeRuneInString(l.input[l.readPosition:])
	}
	l.position = l.readPosition
	l.readPosition += l.chWidth
	l.column++
}

func (l *Lexer) backup() {
	l.readPosition -= l.chWidth
}

// peekChar is similar to readChar, but instead we peek ahead at the next
// character in the input stream rather than actually advancing forward.
// This allows for us to look for two character tokens more easily.
func (l *Lexer) peekChar() rune {
	if l.readPosition >= l.inputLen {
		return 0
	}

	c, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])

	return c
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch, false) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readAlphanumeric() string {
	start := l.position
	for isAlphaNumeric(l.ch, false) {
		l.readChar()
	}
	return l.input[start:l.position]
}

// skipWhitespace is used to skip over general whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// NextToken is used to read from the input stream and identify what the next
// token is
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch {
	case isRune(l.ch, 0):
		// EOF case
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Column = l.column - 1
		tok.Line = l.line
	case isRune(l.ch, '\n'):
		tok = newToken(token.NEWLINE, l.ch, l.column, l.line)
	case isRune(l.ch, '-'):
		fallthrough
	case isRune(l.ch, '+'):
		// this may be a signed integer
		if l.detectSignedInt(&tok) {
			return tok
		}
	case isDigit(l.ch):
		tok.Literal = l.readNumber()
		tok.Type = token.INT
		tok.Column = l.column - len(tok.Literal)
		tok.Line = l.line
		return tok
	case isLetter(l.ch, false):
		tok.Literal = l.readIdentifier()
		switch {
		case tok.Literal == "AS":
			if l.detectASNO(&tok) {
				return tok
			}
		case strings.HasPrefix(tok.Literal, "AS-"):
			if l.detectASNAME(&tok) {
				return tok
			}
		default:
			break
		}
		tok.Type = token.ILLEGAL
		tok.Column = l.column - len(tok.Literal)
		tok.Line = l.line
		return tok
	default:
		tok = newToken(token.ILLEGAL, l.ch, l.column-1, l.line)
	}

	l.readChar()
	return tok
}

func (l *Lexer) detectSignedInt(tok *token.Token) bool {
	if isDigit(l.peekChar()) {
		plus := l.ch
		l.readChar()
		num := l.readNumber()
		literal := string(plus) + string(num)

		tok.Literal = literal
		tok.Type = token.SIGNED_INT
		tok.Column = l.column - len(tok.Literal)
		tok.Line = l.line
		return true
	}

	return false
}

func (l *Lexer) detectASNO(tok *token.Token) bool {
	if isDigit(l.ch) {
		num := l.readNumber()
		tok.Literal = tok.Literal + string(num)
		tok.Type = token.ASNO
		tok.Column = l.column - len(tok.Literal)
		tok.Line = l.line
		return true
	}

	return false
}

func (l *Lexer) detectASNAME(tok *token.Token) bool {
	if isAlphaNumeric(l.ch, false) {
		asName := l.readAlphanumeric()
		tok.Literal = tok.Literal + string(asName)
		lastChar := []rune(tok.Literal[len(tok.Literal)-1:])[0]
		tok.Type = token.ASNAME
		if !isAlphaNumeric(lastChar, true) {
			fmt.Println(tok.Literal, string(lastChar))
			tok.Type = token.ILLEGAL
		}
		tok.Column = l.column - len(tok.Literal)
		tok.Line = l.line
		return true
	}

	return false
}
