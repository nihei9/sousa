package grammar

import (
	"testing"
)

func TestGenerateFirstSets(t *testing.T) {
	st := NewSymbolTable()

	prods := newProds(st, "E'", []*Prod{
		newProd("E'", "E"),
		newProd("E", "E", "+", "T"),
		newProd("E", "T"),
		newProd("T", "T", "*", "F"),
		newProd("T", "F"),
		newProd("F", "(", "E", ")"),
		newProd("F", "id"),
	})

	fss, err := GenerateFirstSets(prods)
	if err != nil {
		t.Fatal(err)
	}
	if fss == nil {
		t.Fatal("GenerateFirstSet returned nil without an error")
	}

	tests := []struct {
		lhs     string
		num     int
		dot     int
		symbols []string
	}{
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
	}

	for _, tt := range tests {
		lhsID := st.Intern(tt.lhs, symbolKindNonTerminal)
		if lhsID.IsNil() {
			t.Errorf("failed to intern a symbol. test: %+v", tt)
			continue
		}

		prod := prods.Get(lhsID)
		if prod == nil {
			t.Errorf("failed to get a production. test: %+v", tt)
			continue
		}

		actualFirst := fss.Get(prod[tt.num], tt.dot)
		if actualFirst == nil {
			t.Errorf("failed to get a first set. test: %+v", tt)
			continue
		}

		expectedFirst := newFirstSet()
		for _, sym := range tt.symbols {
			symID := st.Intern(sym, symbolKindTerminal)
			if symID.IsNil() {
				t.Errorf("failed to intern a symbol. test: %+v, symbol: %v", tt, sym)
				continue
			}

			expectedFirst.put(symID)
		}

		testFirstSet(t, actualFirst, expectedFirst)
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
