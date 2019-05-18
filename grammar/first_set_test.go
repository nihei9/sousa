package grammar

import (
	"testing"
)

type fst struct {
	lhs     string
	num     int
	dot     int
	symbols []string
	empty   bool
}

func TestGenerateFirstSets(t *testing.T) {
	tests := map[string]struct {
		genProds   func(*SymbolTable) Productions
		firstSetes []fst
	}{
		"productions contain only nonempty productions": {
			genProds: func(st *SymbolTable) Productions {
				return newProds(st, "E'", []*Prod{
					newProd("E'", "E"),
					newProd("E", "E", "+", "T"),
					newProd("E", "T"),
					newProd("T", "T", "*", "F"),
					newProd("T", "F"),
					newProd("F", "(", "E", ")"),
					newProd("F", "id"),
				})
			},
			firstSetes: []fst{
				{lhs: "E'", num: 0, dot: 0, symbols: []string{"(", "id"}},
				{lhs: "E", num: 0, dot: 0, symbols: []string{"(", "id"}},
				{lhs: "E", num: 0, dot: 1, symbols: []string{"+"}},
				{lhs: "E", num: 0, dot: 2, symbols: []string{"(", "id"}},
				{lhs: "T", num: 0, dot: 0, symbols: []string{"(", "id"}},
				{lhs: "T", num: 0, dot: 1, symbols: []string{"*"}},
				{lhs: "T", num: 0, dot: 2, symbols: []string{"(", "id"}},
				{lhs: "F", num: 0, dot: 0, symbols: []string{"("}},
				{lhs: "F", num: 0, dot: 1, symbols: []string{"(", "id"}},
				{lhs: "F", num: 0, dot: 2, symbols: []string{")"}},
				{lhs: "F", num: 1, dot: 0, symbols: []string{"id"}},
			},
		},
		"productions contain empty start production": {
			genProds: func(st *SymbolTable) Productions {
				return newProds(st, "s", []*Prod{
					newProd("s"),
				})
			},
			firstSetes: []fst{
				{lhs: "s", num: 0, dot: 0, symbols: []string{}, empty: true},
			},
		},
		"productions contain empty production": {
			genProds: func(st *SymbolTable) Productions {
				return newProds(st, "s", []*Prod{
					newProd("s", "foo"),
					newProd("foo"),
				})
			},
			firstSetes: []fst{
				{lhs: "s", num: 0, dot: 0, symbols: []string{}, empty: true},
				{lhs: "foo", num: 0, dot: 0, symbols: []string{}, empty: true},
			},
		},
		"productions contain nonempty start production and empty one": {
			genProds: func(st *SymbolTable) Productions {
				return newProds(st, "s", []*Prod{
					newProd("s", "foo"),
					newProd("s"),
				})
			},
			firstSetes: []fst{
				{lhs: "s", num: 0, dot: 0, symbols: []string{"foo"}},
				{lhs: "s", num: 1, dot: 0, symbols: []string{}, empty: true},
			},
		},
		"production contain nonempty production and empty one": {
			genProds: func(st *SymbolTable) Productions {
				return newProds(st, "s", []*Prod{
					newProd("s", "foo"),
					newProd("foo", "bar"),
					newProd("foo"),
				})
			},
			firstSetes: []fst{
				{lhs: "s", num: 0, dot: 0, symbols: []string{"bar"}, empty: true},
				{lhs: "foo", num: 0, dot: 0, symbols: []string{"bar"}},
				{lhs: "foo", num: 1, dot: 0, symbols: []string{}, empty: true},
			},
		},
	}
	for _, tt := range tests {
		st := NewSymbolTable()
		prods := tt.genProds(st)

		fss, err := GenerateFirstSets(prods)
		if err != nil {
			t.Fatal(err)
		}
		if fss == nil {
			t.Fatal("GenerateFirstSet returned nil without an error")
		}

		for _, ttFirst := range tt.firstSetes {
			lhsID := st.Intern(ttFirst.lhs, SymbolKindNonTerminal)
			if lhsID.IsNil() {
				t.Errorf("failed to intern a symbol. test: %+v", tt)
				continue
			}

			prod := prods.Get(lhsID)
			if prod == nil {
				t.Errorf("failed to get a production. test: %+v", tt)
				continue
			}

			actualFirst := fss.Get(prod[ttFirst.num], ttFirst.dot)
			if actualFirst == nil {
				t.Errorf("failed to get a first set. test: %+v", tt)
				continue
			}

			expectedFirst := newFirstSet()
			if ttFirst.empty {
				expectedFirst.putEmpty()
			}
			for _, sym := range ttFirst.symbols {
				symID := st.Intern(sym, SymbolKindTerminal)
				if symID.IsNil() {
					t.Errorf("failed to intern a symbol. test: %+v, symbol: %v", tt, sym)
					continue
				}

				expectedFirst.put(symID)
			}

			testFirstSet(t, actualFirst, expectedFirst)
		}
	}
}

func testFirstSet(t *testing.T, actual, expected *FirstSet) {
	t.Helper()

	if actual.empty != expected.empty {
		t.Errorf("empty is mismatched\nwant: %v\ngot: %v", expected.empty, actual.empty)
	}

	if len(actual.symbols) != len(expected.symbols) {
		t.Fatalf("invalid first set\nwant: %+v\ngot: %+v", expected.symbols, actual.symbols)
	}

	aSyms := sortSymbols(actual.symbols.Slice())
	eSyms := sortSymbols(expected.symbols.Slice())
	for i := 0; i < len(eSyms); i++ {
		if aSyms[i] != eSyms[i] {
			t.Fatalf("invalid first set\nwant: %+v\ngot: %+v", expected.symbols, actual.symbols)
		}
	}
}
