package ast2grammar

import (
	"strings"
	"testing"

	"github.com/nihei9/sousa/grammar"
	"github.com/nihei9/sousa/parser"
)

type production struct {
	lhs string
	rhs alternative
}

type alternative []string

func (p *production) genProduction(st *grammar.SymbolTable) (*grammar.Production, error) {
	lhsID := st.Intern(p.lhs, grammar.SymbolKindNonTerminal)
	rhsIDs := make([]grammar.SymbolID, len(p.rhs))
	for i, sym := range p.rhs {
		rhsIDs[i] = st.Intern(sym, grammar.SymbolKindTerminal)
	}

	return grammar.NewProduction(lhsID, rhsIDs)
}

func TestConvert(t *testing.T) {
	tests := []struct {
		src         string
		productions []production
	}{
		{
			src: `E: E "+" T | T; T: T "*" F | F; F: "(" E ")" | id;`,
			productions: []production{
				{lhs: "E'", rhs: alternative{"E"}},
				{lhs: "E", rhs: alternative{"E", "+", "T"}},
				{lhs: "E", rhs: alternative{"T"}},
				{lhs: "T", rhs: alternative{"T", "*", "F"}},
				{lhs: "T", rhs: alternative{"F"}},
				{lhs: "F", rhs: alternative{"(", "E", ")"}},
				{lhs: "F", rhs: alternative{"id"}},
			},
		},
	}
	for _, tt := range tests {
		lex := parser.NewLexer(strings.NewReader(tt.src))
		parser, err := parser.NewParser(lex)
		if err != nil {
			t.Error(err)
			continue
		}
		if parser == nil {
			t.Errorf("parser is nil")
			continue
		}

		root, err := parser.Parse()

		g, err := Convert(root)
		if err != nil {
			t.Error(err)
			continue
		}
		if g == nil {
			t.Error("grammar is nil")
			continue
		}
		if g.SymbolTable == nil {
			t.Error("SymbolTable is nil")
			continue
		}
		if g.Productions == nil {
			t.Error("production is nil")
			continue
		}
		if g.AugmentedStartSymbol.IsNil() {
			t.Error("augmented symbol is nil")
			continue
		}

		for _, ttProd := range tt.productions {
			expectedProd, err := ttProd.genProduction(g.SymbolTable)
			if err != nil {
				t.Error(err)
				continue
			}

			lhsID := g.SymbolTable.Intern(ttProd.lhs, grammar.SymbolKindNonTerminal)
			actualProds := g.Productions.Get(lhsID)
			if actualProds == nil {
				t.Errorf("failed to get %s-production", ttProd.lhs)
				continue
			}

			found := false
			for _, actualProd := range actualProds {
				if actualProd.Equal(expectedProd) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("failed to get an production\nwant: %+v", ttProd)
				continue
			}
		}
	}
}
