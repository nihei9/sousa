package parser

import (
	"fmt"
)

type TokenType string

func (t TokenType) String() string {
	return string(t)
}

const (
	TokenTypeUnknown   = TokenType("UNKNOWN")
	TokenTypeColon     = TokenType(":")
	TokenTypeVBar      = TokenType("|")
	TokenTypeSemicolon = TokenType(";")
	TokenTypeID        = TokenType("ID")
	TokenTypeString    = TokenType("STRING")
)

type Position struct {
	Line   int
	Column int
}

func newPosition() Position {
	return Position{
		Line:   1,
		Column: 1,
	}
}

func (p Position) String() string {
	return fmt.Sprintf("(%v, %v)", p.Line, p.Column)
}

func (p *Position) incrementBy(c rune) {
	if c == '\n' || c == '\r' {
		p.Line += 1
		p.Column = 1
	} else {
		p.Column += 1
	}
}

type Token interface {
	Type() TokenType
	Pos() Position
	Text() string
	IsUnknown() bool
}

type UnknownToken struct {
	pos  Position
	text string
}

func newUnknownToken(text string, pos Position) Token {
	return &UnknownToken{
		pos:  pos,
		text: text,
	}
}

func (t *UnknownToken) String() string  { return fmt.Sprintf("UNKNOWN<%s>", t.text) }
func (t *UnknownToken) Type() TokenType { return TokenTypeUnknown }
func (t *UnknownToken) Pos() Position   { return t.pos }
func (t *UnknownToken) Text() string    { return t.text }
func (t *UnknownToken) IsUnknown() bool { return true }

type SymbolToken struct {
	t   TokenType
	pos Position
}

func newSymbolToken(t TokenType, pos Position) Token {
	return &SymbolToken{
		t:   t,
		pos: pos,
	}
}

func (t *SymbolToken) String() string  { return t.t.String() }
func (t *SymbolToken) Type() TokenType { return t.t }
func (t *SymbolToken) Pos() Position   { return t.pos }
func (t *SymbolToken) Text() string    { return t.t.String() }
func (t *SymbolToken) IsUnknown() bool { return false }

type IDToken struct {
	pos  Position
	text string
}

func newIDToken(text string, pos Position) Token {
	return &IDToken{
		pos:  pos,
		text: text,
	}
}

func (t *IDToken) String() string  { return t.text }
func (t *IDToken) Type() TokenType { return TokenTypeID }
func (t *IDToken) Pos() Position   { return t.pos }
func (t *IDToken) Text() string    { return t.text }
func (t *IDToken) IsUnknown() bool { return false }

type StringToken struct {
	pos  Position
	text string
}

func newStringToken(text string, pos Position) Token {
	return &StringToken{
		pos:  pos,
		text: text,
	}
}

func (t *StringToken) String() string  { return fmt.Sprintf("\"%s\"", t.text) }
func (t *StringToken) Type() TokenType { return TokenTypeString }
func (t *StringToken) Pos() Position   { return t.pos }
func (t *StringToken) Text() string    { return t.text }
func (t *StringToken) IsUnknown() bool { return false }
