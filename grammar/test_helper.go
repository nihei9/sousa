package grammar

import "sort"

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

	st.Intern(start, SymbolKindStart)

	for _, prod := range prods {
		st.Intern(prod.lhs, SymbolKindNonTerminal)
	}

	for _, prod := range prods {
		lhs := st.Intern(prod.lhs, SymbolKindNonTerminal)
		rhs := make([]SymbolID, len(prod.rhs))
		for i, sym := range prod.rhs {
			rhs[i] = st.Intern(sym, SymbolKindTerminal)
		}

		p, _ := NewProduction(lhs, rhs)

		ps.Append(p)
	}

	return ps
}

func sortSymbols(syms []SymbolID) []SymbolID {
	dSyms := make([]SymbolID, len(syms))
	copy(dSyms, syms)
	sort.SliceStable(dSyms, func(i, j int) bool {
		return dSyms[i] < dSyms[j]
	})
	return dSyms
}

func newSymbolGetter(st *SymbolTable) func(string) SymbolID {
	return func(str string) SymbolID {
		return st.lookupByString(str)
	}
}

func newProductionGetter(st *SymbolTable, prods Productions) func(string, int) *Production {
	return func(lhs string, num int) *Production {
		id := st.Intern(lhs, SymbolKindNonTerminal)
		return prods.Get(id)[num]
	}
}

type lr0Item struct {
	lhs       string
	num       int
	dot       int
	initial   bool
	reducible bool
}

func genKernel(items []lr0Item, st *SymbolTable, prods Productions) (*KernelItems, error) {
	k := NewKernelItems()
	for _, i := range items {
		item, err := genLR0Item(i, st, prods)
		if err != nil {
			return nil, err
		}

		k.Append(item)
	}

	return k, nil
}

func genLR0Item(i lr0Item, st *SymbolTable, prods Productions) (*LR0Item, error) {
	P := newProductionGetter(st, prods)

	switch {
	case i.initial:
		return NewInitialLR0Item(P(i.lhs, i.num))
	case i.reducible:
		return NewReducibleLR0Item(P(i.lhs, i.num))
	default:
		return NewLR0Item(P(i.lhs, i.num), i.dot)
	}
}
