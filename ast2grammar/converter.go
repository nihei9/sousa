package ast2grammar

import (
	"fmt"

	"github.com/nihei9/sousa/grammar"
	"github.com/nihei9/sousa/parser"
)

type Grammar struct {
	SymbolTable          *grammar.SymbolTable
	Productions          grammar.Productions
	AugmentedStartSymbol grammar.SymbolID
}

func Convert(root *parser.AST) (*Grammar, error) {
	st := grammar.NewSymbolTable()
	prods := grammar.NewProductions()
	g := &Grammar{
		SymbolTable: st,
		Productions: prods,
	}

	isFirst := true
	for _, prodAST := range root.Children {
		if prodAST.State != parser.StateProduction {
			continue
		}

		if isFirst {
			isFirst = false

			lhsAST := prodAST.Children[0]
			augmentedStartSymbol := fmt.Sprintf("%s'", lhsAST.Tokens[0].Text())
			lhsID := st.Intern(augmentedStartSymbol, grammar.SymbolKindStart)

			startSymbol := lhsAST.Tokens[0].Text()
			startSymbolID := st.Intern(startSymbol, grammar.SymbolKindNonTerminal)
			rhsIDs := []grammar.SymbolID{startSymbolID}

			prod, err := grammar.NewProduction(lhsID, rhsIDs)
			if err != nil {
				return nil, err
			}

			prods.Append(prod)

			g.AugmentedStartSymbol = lhsID
		}

		lhsAST := prodAST.Children[0]
		st.Intern(lhsAST.Tokens[0].Text(), grammar.SymbolKindNonTerminal)
	}

	for _, prodAST := range root.Children {
		if prodAST.State != parser.StateProduction {
			continue
		}

		lhsAST := prodAST.Children[0]
		lhsID := st.Intern(lhsAST.Tokens[0].Text(), grammar.SymbolKindNonTerminal)

		for _, altAST := range prodAST.Children[1].Children {
			rhsIDs := make([]grammar.SymbolID, len(altAST.Tokens))
			for i, termTok := range altAST.Tokens {
				termID := st.Intern(termTok.Text(), grammar.SymbolKindTerminal)
				rhsIDs[i] = termID
			}

			prod, err := grammar.NewProduction(lhsID, rhsIDs)
			if err != nil {
				return nil, err
			}

			prods.Append(prod)
		}
	}

	return g, nil
}
