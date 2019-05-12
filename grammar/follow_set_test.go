package grammar

import (
	"testing"
)

func TestFollowSet(t *testing.T) {
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

	first, err := GenerateFirstSets(prods)
	if err != nil {
		t.Fatal(err)
	}
	if first == nil {
		t.Fatal("GenerateFirstSet returned nil without an error")
	}

	follow, err := GenerateFollowSets(prods, first)
	if err != nil {
		t.Fatal(err)
	}
	if follow == nil {
		t.Fatal("GenerateFollowSet returned nil without an error")
	}

	tests := []struct {
		nSym    string
		eof     bool
		symbols []string
	}{
		{nSym: "E'", eof: true, symbols: []string{}},
		{nSym: "E", eof: true, symbols: []string{"+", ")"}},
		{nSym: "T", eof: true, symbols: []string{"*", "+", ")"}},
		{nSym: "F", eof: true, symbols: []string{"*", "+", ")"}},
	}

	for _, tt := range tests {
		nSymID := st.Intern(tt.nSym, SymbolKindNonTerminal)
		if nSymID.IsNil() {
			t.Errorf("failed to intern a symbol. test: %+v", tt)
			continue
		}

		f := follow.Get(nSymID)
		if f == nil {
			t.Errorf("failed to get a follow set. %+v", tt)
			continue
		}

		expectedFollow := newFollowSet()
		if tt.eof {
			expectedFollow.putEOF()
		}
		for _, sym := range tt.symbols {
			symID := st.Intern(sym, SymbolKindTerminal)
			if symID.IsNil() {
				t.Errorf("failed to intern a symbol. test: %+v, symbol: %v", tt, sym)
				continue
			}

			expectedFollow.put(symID)
		}

		testFollowSet(t, f, expectedFollow)
	}
}

func testFollowSet(t *testing.T, actual, expected *FollowSet) {
	t.Helper()

	if actual.eof != expected.eof {
		t.Errorf("eof is mismatched\nwant: %v\ngot: %v", expected.eof, actual.eof)
	}

	if len(actual.symbols) != len(expected.symbols) {
		t.Fatalf("invalid follow set\nwant: %+v\ngot: %+v", expected.symbols, actual.symbols)
	}

	aSyms := sortSymbols(actual.symbols.Slice())
	eSyms := sortSymbols(expected.symbols.Slice())
	for i := 0; i < len(eSyms); i++ {
		if aSyms[i] != eSyms[i] {
			t.Fatalf("invalid follow set\nwant: %+v\ngot: %+v", expected.symbols, actual.symbols)
		}
	}
}
