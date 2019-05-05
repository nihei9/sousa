package grammar

type Prod struct {
	lhs string
	rhs []string
}

func newProd(lhs string, rhs ...string) *Prod {
	return &Prod{
		lhs: lhs,
		rhs: rhs,
	}
}

func newProds(st *SymbolTable, start string, prods []*Prod) Productions {
	ps := NewProductions()

	st.Intern(start, symbolKindStart)

	for _, prod := range prods {
		st.Intern(prod.lhs, symbolKindNonTerminal)
	}

	for _, prod := range prods {
		lhs := st.Intern(prod.lhs, symbolKindNonTerminal)
		rhs := make([]SymbolID, len(prod.rhs))
		for i, sym := range prod.rhs {
			rhs[i] = st.Intern(sym, symbolKindTerminal)
		}

		p, _ := NewProduction(lhs, rhs)

		ps.Append(p)
	}

	return ps
}
