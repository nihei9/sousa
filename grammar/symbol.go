package grammar

import (
	"fmt"
	"strconv"
)

type SymbolID int

const (
	symbolIDNil   = SymbolID(0)
	symbolIDStart = SymbolID(1)
)

func NewSymbolID() SymbolID {
	return symbolIDStart
}

func (sid SymbolID) String() string {
	return strconv.Itoa(int(sid))
}

func (sid SymbolID) IsNil() bool {
	return sid == symbolIDNil
}

func (sid *SymbolID) Next() SymbolID {
	id := *sid
	*sid = *sid + 1
	return id
}

type SymbolKind string

const (
	symbolKindNil         = SymbolKind("")
	symbolKindTerminal    = SymbolKind("terminal")
	symbolKindNonTerminal = SymbolKind("non-terminal")
)

func (sk SymbolKind) String() string {
	return string(sk)
}

func (sk SymbolKind) IsNil() bool {
	return sk == symbolKindNil
}

func (sk SymbolKind) IsTerminalSymbol() bool {
	return sk == symbolKindTerminal
}

func (sk SymbolKind) IsNonTerminalSymbol() bool {
	return sk == symbolKindNonTerminal
}

type Symbol struct {
	id   SymbolID
	kind SymbolKind
}

type SymbolTable struct {
	str2Sym map[string]*Symbol
	id2Sym  map[SymbolID]*Symbol
	id      SymbolID
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		str2Sym: map[string]*Symbol{},
		id2Sym:  map[SymbolID]*Symbol{},
		id:      NewSymbolID(),
	}
}

func (st *SymbolTable) Intern(str string) SymbolID {
	if sym, ok := st.str2Sym[str]; ok {
		return sym.id
	}

	sym := &Symbol{
		id:   st.id.Next(),
		kind: symbolKindTerminal,
	}

	st.str2Sym[str] = sym
	st.id2Sym[sym.id] = sym

	return sym.id
}

func (st *SymbolTable) SymbolKind(id SymbolID) (SymbolKind, error) {
	if sym, ok := st.id2Sym[id]; ok {
		if sym.kind.IsNil() {
			return symbolKindNil, fmt.Errorf("symbol kind is nil. symbol id: %s", id)
		}

		return sym.kind, nil
	}

	return symbolKindNil, fmt.Errorf("unknown symbol id")
}

func (st *SymbolTable) MarkAsNonTerminalSymbol(id SymbolID) error {
	if id.IsNil() {
		return fmt.Errorf("symbol id passed is nil")
	}

	sym, ok := st.id2Sym[id]
	if !ok {
		return fmt.Errorf("symbol id passed doesn't exist. symbol id: %s", id)
	}

	sym.kind = symbolKindNonTerminal

	return nil
}
