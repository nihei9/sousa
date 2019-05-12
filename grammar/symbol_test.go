package grammar

import (
	"testing"
)

func TestSymbolTable(t *testing.T) {
	validSymbols := []string{
		"0123456789",
		"abcdefghifklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?",
		" ",
		"\t\r\n",
	}

	invalidSymbols := []string{
		"",
	}

	t.Run("Intern valid symbols as a terminal symbol", func(t *testing.T) {
		testValidSymbols(t, validSymbols, SymbolKindTerminal)
	})

	t.Run("Intern valid symbols as a non-terminal symbol", func(t *testing.T) {
		testValidSymbols(t, validSymbols, SymbolKindNonTerminal)
	})

	t.Run("Intern valid symbols as a start symbol", func(t *testing.T) {
		testValidSymbols(t, validSymbols, SymbolKindStart)
	})

	t.Run("Intern invalid symbols as a terminal symbol", func(t *testing.T) {
		testInvalidSymbols(t, invalidSymbols, SymbolKindTerminal)
	})

	t.Run("Intern invalid symbols as a non-terminal symbol", func(t *testing.T) {
		testInvalidSymbols(t, invalidSymbols, SymbolKindNonTerminal)
	})

	t.Run("Intern invalid symbols as a start symbol", func(t *testing.T) {
		testInvalidSymbols(t, invalidSymbols, SymbolKindStart)
	})
}

func testValidSymbols(t *testing.T, symbols []string, kind SymbolKind) {
	t.Helper()

	st := NewSymbolTable()
	if st == nil {
		t.Fatal("NewSymbolTable() returns nil")
	}

	for _, sym := range symbols {
		sid := st.Intern(sym, kind)
		if sid.IsNil() {
			t.Fatalf("unexpected symbol ID\nwant: %v\ngot: %v", "non-nil symbol ID", sid)
		}
		if sid.Kind() != kind {
			t.Fatalf("unexpected symbol kind\nwant: %v\ngot: %v", kind, sid.Kind())
		}
	}
}

func testInvalidSymbols(t *testing.T, symbols []string, kind SymbolKind) {
	t.Helper()

	st := NewSymbolTable()
	if st == nil {
		t.Fatal("NewSymbolTable() returns nil")
	}

	for _, sym := range symbols {
		sid := st.Intern(sym, kind)
		if !sid.IsNil() {
			t.Fatalf("unexpected symbol ID\nwant: %v\ngot: %v", "nil symbol ID", sid)
		}
		if !sid.Kind().IsNil() {
			t.Fatalf("unexpected symbol kind\nwant: %v\ngot: %v", "nil symbol kind", sid.Kind())
		}
	}
}
