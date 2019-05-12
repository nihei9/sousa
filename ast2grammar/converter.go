package ast2grammar

import (
	"github.com/nihei9/sousa/grammar"
	"github.com/nihei9/sousa/parser"
)

func Convert(root *parser.AST) (*grammar.SymbolTable, grammar.Productions, error) {
	st := grammar.NewSymbolTable()
	prods := grammar.NewProductions()

	for _, prodAST := range root.Children {
		if prodAST.State != parser.StateProduction {
			continue
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
				return nil, nil, err
			}

			prods.Append(prod)
		}
	}

	return st, prods, nil
}
