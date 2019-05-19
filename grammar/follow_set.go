package grammar

import (
	"fmt"
)

type FollowSets map[SymbolID]*FollowSet

func newFollowSets() FollowSets {
	return FollowSets{}
}

func (fss FollowSets) Get(sym SymbolID) *FollowSet {
	return fss[sym]
}

func (fss FollowSets) put(fs *FollowSet, sym SymbolID) {
	fss[sym] = fs
}

type FollowSet struct {
	symbols SymbolSet
	eof     bool
}

func newFollowSet() *FollowSet {
	return &FollowSet{
		symbols: SymbolSet{},
		eof:     false,
	}
}

func (fs *FollowSet) String() string {
	if fs.eof {
		return fmt.Sprintf("{EOF} + %v", fs.symbols)
	}

	return fs.symbols.String()
}

func (fs *FollowSet) put(sym SymbolID) {
	fs.symbols.put(sym)
}

func (fs *FollowSet) putEOF() {
	fs.eof = true
}

func (fs *FollowSet) merge(fst *FirstSet, flw *FollowSet) {
	if fst != nil {
		for sym, _ := range fst.symbols {
			fs.symbols.put(sym)
		}
	}

	if flw != nil {
		for sym, _ := range flw.symbols {
			fs.symbols.put(sym)
		}

		if flw.eof {
			fs.putEOF()
		}
	}
}

// FollowSetComputationContext is a context of computation of the FOLLOW set.
type FollowSetComputationContext struct {
	prods  Productions
	first  FirstSets
	follow FollowSets
	stack  FollowSetComputationStack
}

type FollowSetComputationStack []SymbolID

func (cs FollowSetComputationStack) String() string {
	s := ""
	for _, frame := range cs {
		if s != "" {
			s += " "
		}
		s += fmt.Sprintf("%v", frame)
	}

	return "{" + s + "}"
}

func (cs *FollowSetComputationStack) push(frame SymbolID) {
	*cs = append(*cs, frame)
}

func (cs *FollowSetComputationStack) pop() {
	*cs = []SymbolID(*cs)[:len(*cs)-1]
}

func newFollowSetComputationContext(prods Productions, first FirstSets) *FollowSetComputationContext {
	return &FollowSetComputationContext{
		prods:  prods,
		first:  first,
		follow: newFollowSets(),
		stack:  []SymbolID{},
	}
}

func (cc *FollowSetComputationContext) push(sym SymbolID) {
	cc.stack.push(sym)
}

func (cc *FollowSetComputationContext) pop() {
	cc.stack.pop()
}

func (cc *FollowSetComputationContext) isAlreadyStacked(sym SymbolID) bool {
	for _, f := range cc.stack {
		if f == sym {
			return true
		}
	}

	return false
}

func GenerateFollowSets(prods Productions, first FirstSets) (FollowSets, error) {
	cc := newFollowSetComputationContext(prods, first)

	for _, ps := range prods.All() {
		for _, p := range ps {
			_, err := follow(p.lhs, cc)
			if err != nil {
				return nil, err
			}
		}
	}

	return cc.follow, nil
}

func follow(sym SymbolID, cc *FollowSetComputationContext) (*FollowSet, error) {
	// validations
	{
		if sym.IsNil() {
			return nil, fmt.Errorf("symbol is nil")
		}
	}

	// guards for avoiding the infinite recursion
	{
		// already computed
		if fs := cc.follow.Get(sym); fs != nil {
			return fs, nil
		}

		if cc.isAlreadyStacked(sym) {
			return newFollowSet(), nil
		}
	}

	cc.push(sym)
	defer cc.pop()

	fs := newFollowSet()

	if sym.Kind().IsStartSymbol() {
		fs.putEOF()
	}

	for _, ps := range cc.prods.All() {
		for _, p := range ps {
			for i, rhsSym := range p.rhs {
				if rhsSym == sym {
					if i+1 < p.rhsLen {
						fst := cc.first.Get(p, i+1)
						if fst == nil {
							return nil, fmt.Errorf("failed to get a FIRST set. %v-%v", p.fingerprint, i)
						}
						fs.merge(fst, nil)

						if !fst.empty {
							continue
						}
					}

					flw, err := follow(p.lhs, cc)
					if err != nil {
						return nil, err
					}
					fs.merge(nil, flw)
				}
			}
		}
	}
	cc.follow.put(fs, sym)

	return fs, nil
}
