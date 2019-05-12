package parser

import (
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	tests := map[string]struct {
		src string
		err bool
	}{
		"the source contains only non-empty productions": {
			src: `E: E "+" T | T; T: T "*" F | F; F: "(" E ")" | id;`,
			err: false,
		},
		"the source contains non-empty and empty productions": {
			src: `foo: ; bar: | ; baz: | | ; bra: | abc | ;`,
			err: false,
		},
	}
	for caption, tt := range tests {
		lex := NewLexer(strings.NewReader(tt.src))
		parser, err := NewParser(lex)
		if err != nil {
			t.Errorf("%v. test: %s", err, caption)
			continue
		}
		if parser == nil {
			t.Errorf("parser is nil. test: %s", caption)
			continue
		}

		ast, err := parser.Parse()
		if tt.err {
			if err == nil {
				t.Errorf("error is nil. test: %s", caption)
				continue
			}
			if ast != nil {
				t.Errorf("AST is not nil. test: %s", caption)
			}
		} else {
			if err != nil {
				t.Errorf("%v. test: %s", err, caption)
				continue
			}
			if ast == nil {
				t.Errorf("AST is nil. test: %s", caption)
			}
			//			printAST(t, ast, 0)
		}
	}
}

func printAST(t *testing.T, ast *AST, depth int) {
	if ast == nil {
		return
	}

	indent := strings.Repeat("    ", depth)
	t.Logf("%s%s %+v", indent, ast.State, ast.Tokens)
	for _, child := range ast.Children {
		printAST(t, child, depth+1)
	}
}
