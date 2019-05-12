package parser

import (
	"strings"
	"testing"
)

func pos(line, column int) Position {
	return Position{
		Line:   line,
		Column: column,
	}
}

func TestLexer_Run(t *testing.T) {
	dummyPos := newPosition()

	tests := map[string]struct {
		src    string
		tokens []Token
		err    error
	}{
		"src contains all types of tokens": {
			src: `|:;id"this is string" ???`,
			tokens: []Token{
				newSymbolToken(TokenTypeVBar, dummyPos),
				newSymbolToken(TokenTypeColon, dummyPos),
				newSymbolToken(TokenTypeSemicolon, dummyPos),
				newIDToken("id", dummyPos),
				newStringToken("this is string", dummyPos),
				newUnknownToken("???", dummyPos),
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		l := NewLexer(strings.NewReader(tt.src))
		for _, eTok := range tt.tokens {
			aTok, err := l.Next()
			if err != nil {
				t.Error(err)
				continue
			}
			if !matchToken(eTok, aTok) {
				t.Errorf("unexpected token\nwant: %v\ngot: %v", eTok, aTok)
				continue
			}
			lastTok := l.LastToken()
			if !matchToken(eTok, lastTok) {
				t.Errorf("unexpected token\nwant: %v\ngot: %v", eTok, lastTok)
				continue
			}
			err = l.Error()
			if err != nil {
				t.Error(err)
				continue
			}
		}
		err := l.Error()
		if err != tt.err {
			t.Errorf("unexpected error\nwant: %v\ngot: %v", tt.err, err)
			continue
		}
	}
}

func matchToken(expected, actual Token) bool {
	// Don't check Position
	if actual.IsUnknown() != expected.IsUnknown() || actual.Type() != expected.Type() || actual.Text() != expected.Text() {
		return false
	}

	return true
}
