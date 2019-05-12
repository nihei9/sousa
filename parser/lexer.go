package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Lexer interface {
	Next() (Token, error)
	Error() error
	LastToken() Token
}

type lexer struct {
	src         *bufio.Reader
	pos         Position
	tokenCh     chan Token
	lastToken   Token
	lastChar    rune
	prevCharPos Position
	err         error
	unreadable  bool
}

func NewLexer(src io.Reader) Lexer {
	return &lexer{
		src:         bufio.NewReader(src),
		pos:         newPosition(),
		tokenCh:     nil,
		lastToken:   nil,
		prevCharPos: newPosition(),
		err:         nil,
		unreadable:  false,
	}
}

func (l *lexer) Next() (Token, error) {
	tok, err := l.next()
	if tok != nil {
		l.lastToken = tok
	}
	if err != nil {
		l.err = err
	}
	return tok, err
}

func (l *lexer) next() (Token, error) {
	err := l.skipWhitespace()
	if err != nil {
		return nil, err
	}

	pos := l.pos
	c, eof, err := l.read()
	if err != nil {
		return nil, err
	}
	if eof {
		return newEOFToken(pos), nil
	}

	switch {
	case c == '|':
		return newSymbolToken(TokenTypeVBar, pos), nil
	case c == ':':
		return newSymbolToken(TokenTypeColon, pos), nil
	case c == ';':
		return newSymbolToken(TokenTypeSemicolon, pos), nil
	case c == '"':
		text, err := l.readString()
		if err != nil {
			return nil, err
		}
		return newStringToken(text, pos), nil
	case isIDChar(c):
		text, err := l.readID()
		if err != nil {
			return nil, err
		}
		return newIDToken(text, pos), nil
	}

	unknownText, err := l.readUnknown()
	if err != nil {
		return nil, err
	}

	return newUnknownToken(unknownText, pos), nil
}

func (l *lexer) readString() (string, error) {
	var b strings.Builder
	for {
		c, eof, err := l.read()
		if err != nil {
			return "", err
		}
		if eof {
			return "", fmt.Errorf("string unclosed")
		}
		if c == '"' {
			break
		}
		fmt.Fprint(&b, string(c))
	}

	return b.String(), nil
}

func (l *lexer) readID() (string, error) {
	var b strings.Builder
	fmt.Fprint(&b, string(l.lastChar))
	for {
		c, eof, err := l.read()
		if err != nil {
			return "", err
		}
		if eof {
			break
		}
		if !isIDChar(c) {
			err := l.unread()
			if err != nil {
				return "", err
			}
			break
		}
		fmt.Fprint(&b, string(c))
	}

	return b.String(), nil
}

func (l *lexer) readUnknown() (string, error) {
	var b strings.Builder
	fmt.Fprint(&b, string(l.lastChar))
	for {
		c, eof, err := l.read()
		if err != nil {
			return "", err
		}
		if eof {
			break
		}
		if !isUnknownChar(c) {
			err := l.unread()
			if err != nil {
				return "", err
			}
			break
		}
		fmt.Fprint(&b, string(c))
	}

	return b.String(), nil
}

func (l *lexer) skipWhitespace() error {
	for {
		c, eof, err := l.read()
		if err != nil {
			return err
		}
		if eof {
			return nil
		}
		if !isWhitespace(c) {
			err := l.unread()
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (l *lexer) read() (rune, bool, error) {
	c, _, err := l.src.ReadRune()
	if err != nil {
		if err == io.EOF {
			return '_', true, nil
		}
		return '_', false, err
	}

	l.prevCharPos = l.pos
	l.pos.incrementBy(c)
	l.lastChar = c
	l.unreadable = true
	return c, false, nil
}

func (l *lexer) unread() error {
	if !l.unreadable {
		return fmt.Errorf("unreadable")
	}
	l.unreadable = false
	l.pos = l.prevCharPos
	return l.src.UnreadRune()
}

func isIDChar(c rune) bool {
	return unicode.IsLetter(c)
}

func isWhitespace(c rune) bool {
	return unicode.IsSpace(c)
}

func isUnknownChar(c rune) bool {
	return !isFirstChar(c)
}

func isFirstChar(c rune) bool {
	return c == ':' || c == '|' || c == ';' || c == '"' || isIDChar(c) || isWhitespace(c)
}

func (l *lexer) Error() error {
	return l.err
}

func (l *lexer) LastToken() Token {
	return l.lastToken
}
