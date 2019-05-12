package grammar

import (
	"fmt"
	"strconv"
)

type bareSymbolID int

const (
	bareSymbolIDNil   = bareSymbolID(0)
	bareSymbolIDStart = bareSymbolID(1)
)

func newBareSymbolID() bareSymbolID {
	return bareSymbolIDStart
}

func (id bareSymbolID) String() string {
	return strconv.Itoa(int(id))
}

func (id bareSymbolID) isNil() bool {
	return id == bareSymbolIDNil
}

func (id *bareSymbolID) next() bareSymbolID {
	nextID := *id
	*id = *id + 1
	return nextID
}

type SymbolKind string

const (
	SymbolKindNil         = SymbolKind("")
	SymbolKindStart       = SymbolKind("start")
	SymbolKindTerminal    = SymbolKind("terminal")
	SymbolKindNonTerminal = SymbolKind("non-terminal")
)

func (sk SymbolKind) String() string {
	return string(sk)
}

func (sk SymbolKind) IsNil() bool {
	return sk == SymbolKindNil
}

func (sk SymbolKind) IsTerminalSymbol() bool {
	return sk == SymbolKindTerminal
}

func (sk SymbolKind) IsNonTerminalSymbol() bool {
	return sk == SymbolKindNonTerminal || sk == SymbolKindStart
}

func (sk SymbolKind) IsStartSymbol() bool {
	return sk == SymbolKindStart
}

type symbolIDGenerator struct {
	bareID bareSymbolID
}

func newSymbolIDGenerator() *symbolIDGenerator {
	return &symbolIDGenerator{
		bareID: newBareSymbolID(),
	}
}

func (g *symbolIDGenerator) next(kind SymbolKind) (SymbolID, bareSymbolID) {
	if kind.IsNil() {
		return symbolIDNil, bareSymbolIDNil
	}

	bareID := g.bareID.next()

	prefix := ""
	if kind.IsTerminalSymbol() {
		prefix = "t"
	} else if kind.IsStartSymbol() {
		prefix = "s"
	} else if kind.IsNonTerminalSymbol() {
		prefix = "n"
	}

	return SymbolID(fmt.Sprintf("%s%v", prefix, bareID)), bareID
}

type SymbolID string

const (
	symbolIDNil = SymbolID("")
)

func (id SymbolID) String() string {
	return string(id)
}

func (id SymbolID) IsNil() bool {
	return id == symbolIDNil
}

func (id SymbolID) Kind() SymbolKind {
	if id.IsNil() {
		return SymbolKindNil
	}

	switch id[:1] {
	case "t":
		return SymbolKindTerminal
	case "s":
		return SymbolKindStart
	case "n":
		return SymbolKindNonTerminal
	}

	return SymbolKindNil
}

type Symbol struct {
	id     SymbolID
	bareID bareSymbolID
	kind   SymbolKind
}

type SymbolTable struct {
	str2Sym map[string]*Symbol
	id2Sym  map[bareSymbolID]*Symbol
	idGen   *symbolIDGenerator
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		str2Sym: map[string]*Symbol{},
		id2Sym:  map[bareSymbolID]*Symbol{},
		idGen:   newSymbolIDGenerator(),
	}
}

func (st *SymbolTable) Intern(str string, kind SymbolKind) SymbolID {
	if str == "" || kind.IsNil() {
		return symbolIDNil
	}

	if sym, ok := st.str2Sym[str]; ok {
		return sym.id
	}

	id, bareID := st.idGen.next(kind)
	sym := &Symbol{
		id:     id,
		bareID: bareID,
		kind:   kind,
	}

	st.str2Sym[str] = sym
	st.id2Sym[bareID] = sym

	return id
}

func (st *SymbolTable) lookupByString(str string) SymbolID {
	if str == "" {
		return symbolIDNil
	}

	if sym, ok := st.str2Sym[str]; ok {
		return sym.id
	}

	return symbolIDNil
}
