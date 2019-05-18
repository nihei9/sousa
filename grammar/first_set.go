package grammar

import (
	"fmt"
)

// SymbolSet is a set of SymbolIDs.
type SymbolSet map[SymbolID]struct{}

func (ss SymbolSet) String() string {
	s := ""
	for sym, _ := range ss {
		if s != "" {
			s += " "
		}
		s += sym.String()
	}

	return "{" + s + "}"
}

func (ss SymbolSet) put(sym SymbolID) {
	ss[sym] = struct{}{}
}

func (ss SymbolSet) Slice() []SymbolID {
	s := make([]SymbolID, len(ss))
	for sym, _ := range ss {
		s = append(s, sym)
	}
	return s
}

// FirstSet represents a FIRST set.
type FirstSet struct {
	symbols SymbolSet
	empty   bool
}

func newFirstSet() *FirstSet {
	return &FirstSet{
		symbols: SymbolSet{},
		empty:   false,
	}
}

func (fs *FirstSet) String() string {
	s := ""
	if fs.empty {
		s += "{ε} + "
	}
	s += fs.symbols.String()

	return s
}

func (fs *FirstSet) put(syms ...SymbolID) {
	for _, sym := range syms {
		fs.symbols.put(sym)
	}
}

func (fs *FirstSet) putEmpty() {
	fs.empty = true
}

func (fs *FirstSet) merge(target *FirstSet) {
	for sym, _ := range target.symbols {
		fs.symbols.put(sym)
	}
}

// FirstSet represents a set of FIRST sets.
type FirstSets map[ProductionFingerprint][]*FirstSet

func newFirstSets(prods Productions) FirstSets {
	fss := FirstSets{}

	for _, ps := range prods {
		for _, p := range ps {
			len := p.rhsLen
			if p.isEmpty() {
				len = 1
			}
			fss[p.fingerprint] = make([]*FirstSet, len)
		}
	}

	return fss
}

func (fss FirstSets) Get(prod *Production, head int) *FirstSet {
	fs := fss[prod.fingerprint][head]

	//	log.Printf("[FirstSets.Get] fss[%v]: %v", prod.fingerprint, fss[prod.fingerprint])
	//	log.Printf("[FirstSets.Get] return fss[%v][%v] %v", prod.fingerprint, head, fs)

	return fs
}

func (fss FirstSets) put(fs *FirstSet, prod *Production, head int) {
	fss[prod.fingerprint][head] = fs

	//	log.Printf("[FirstSets.put] fss[%v]: %v", prod.fingerprint, fss[prod.fingerprint])
}

// FirstSetComputationContext is a context of computation of the FIRST set.
type FirstSetComputationContext struct {
	prods Productions
	first FirstSets
	stack FirstSetComputationStack
}

type FirstSetComputationStack []*FirstSetFrame

func (cs FirstSetComputationStack) String() string {
	s := ""
	for _, frame := range cs {
		if s != "" {
			s += " "
		}
		s += fmt.Sprintf("%v@%v", frame.prodFingerprint, frame.head)
	}

	return "{" + s + "}"
}

func (cs *FirstSetComputationStack) push(frame *FirstSetFrame) {
	*cs = append(*cs, frame)
}

func (cs *FirstSetComputationStack) pop() {
	*cs = []*FirstSetFrame(*cs)[:len(*cs)-1]
}

type FirstSetFrame struct {
	prodFingerprint ProductionFingerprint
	head            int
}

func newFirstSetComputationContext(prods Productions) *FirstSetComputationContext {
	return &FirstSetComputationContext{
		prods: prods,
		first: newFirstSets(prods),
		stack: []*FirstSetFrame{},
	}
}

func (cc *FirstSetComputationContext) push(prod *Production, head int) {
	cc.stack.push(&FirstSetFrame{
		prodFingerprint: prod.fingerprint,
		head:            head,
	})
}

func (cc *FirstSetComputationContext) pop() {
	cc.stack.pop()
}

func (cc *FirstSetComputationContext) isAlreadyStacked(prod *Production, head int) bool {
	for _, f := range cc.stack {
		if f.prodFingerprint == prod.fingerprint && f.head == head {
			return true
		}
	}

	return false
}

func GenerateFirstSets(prods Productions) (FirstSets, error) {
	cc := newFirstSetComputationContext(prods)

	for _, ps := range prods {
		for _, p := range ps {
			if p.isEmpty() {
				_, err := first(p, 0, cc)
				if err != nil {
					return nil, err
				}
			} else {
				for i := 0; i < p.rhsLen; i++ {
					_, err := first(p, i, cc)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return cc.first, nil
}

func first(prod *Production, head int, cc *FirstSetComputationContext) (*FirstSet, error) {
	//	log.Printf("[first] %v @%v %v", prod.fingerprint, head, cc.stack)
	//	defer log.Printf("[first] return")

	// validations
	{
		if prod.isEmpty() {
			if head != 0 {
				return nil, fmt.Errorf("a production passed is the empty rule but head is not 0. got: %v", head)
			}
		} else {
			if head < 0 || head >= prod.rhsLen {
				return nil, fmt.Errorf("head is out of bounds. head must be between 0 and %v. got: %v", prod.rhsLen-1, head)
			}
		}
	}

	// guards for avoiding the infinite recursion
	{
		// already computed
		if fs := cc.first.Get(prod, head); fs != nil {
			//			log.Printf("[first] already computed")

			return fs, nil
		}

		if cc.isAlreadyStacked(prod, head) {
			//			log.Printf("[first] already stacked")

			return newFirstSet(), nil
		}
	}

	cc.push(prod, head)
	defer cc.pop()

	symbols := []SymbolID{}
	if !prod.isEmpty() {
		symbols = prod.rhs[head:]
	}

	// When symbols is empty, its FIRST set contains the only EMPTY symbol.
	if len(symbols) <= 0 {
		fs := newFirstSet()
		fs.putEmpty()
		cc.first.put(fs, prod, head)

		return fs, nil
	}

	sym := symbols[0]

	{
		symKind := sym.Kind()
		if symKind.IsNil() {
			return nil, fmt.Errorf("invalid symbol")
		}
		if symKind.IsTerminalSymbol() {
			fs := newFirstSet()
			fs.put(sym)
			cc.first.put(fs, prod, head)

			return fs, nil
		}
	}

	fs := newFirstSet()
	for _, symProd := range cc.prods.Get(sym) {
		symFs, err := first(symProd, 0, cc)
		if err != nil {
			return nil, err
		}

		fs.merge(symFs)

		if symFs.empty {
			if symProd.isEmpty() {
				fs.putEmpty()
			} else {
				f, err := first(symProd, head+1, cc)
				if err != nil {
					return nil, err
				}
				fs.merge(f)
			}
		}
	}
	cc.first.put(fs, prod, head)

	return fs, nil
}
