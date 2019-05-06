package grammar

import (
	"fmt"
	"sort"
)

type LR0ItemFingerprint string

func (fp LR0ItemFingerprint) String() string {
	return string(fp)
}

func generateLR0ItemFingerprint(i *LR0Item) LR0ItemFingerprint {
	return LR0ItemFingerprint(fmt.Sprintf("%v-%v", i.prod.fingerprint, i.dot))
}

// LR0Item represents a LR(0) item.
type LR0Item struct {
	fingerprint LR0ItemFingerprint
	prod        *Production

	// E -> E + T
	//
	// Dot | Item
	// ----+-------------
	// 0   | E ->・E + T
	// 1   | E -> E・+ T
	// 2   | E -> E +・T
	dot int

	// If initial is true, it means lhs of the production is the expansion start symbol and dot is 0.
	// It looks like S' ->・S.
	initial bool

	// E -> E + T
	//
	// When reducible is true, the item looks like E -> E + T・.
	reducible bool
}

func NewLR0Item(prod *Production, dot int) (*LR0Item, error) {
	if prod == nil {
		return nil, fmt.Errorf("production rule passed is nil")
	}

	if dot < 0 || dot > prod.rhsLen {
		return nil, fmt.Errorf("dot must be between 0 and %v", prod.rhsLen)
	}

	reducible := false
	if dot == prod.rhsLen {
		reducible = true
	}

	item := &LR0Item{
		prod:      prod,
		dot:       dot,
		initial:   false,
		reducible: reducible,
	}
	item.fingerprint = generateLR0ItemFingerprint(item)

	return item, nil
}

func NewReducibleLR0Item(prod *Production) (*LR0Item, error) {
	if prod == nil {
		return nil, fmt.Errorf("production rule passed is nil")
	}

	item := &LR0Item{
		prod:      prod,
		dot:       prod.rhsLen,
		initial:   true,
		reducible: true,
	}
	item.fingerprint = generateLR0ItemFingerprint(item)

	return item, nil
}

func NewInitialLR0Item(prod *Production) (*LR0Item, error) {
	if prod == nil {
		return nil, fmt.Errorf("production rule passed is nil")
	}
	if !prod.lhs.Kind().IsStartSymbol() {
		return nil, fmt.Errorf("LHS of a production is not the start symbol")
	}

	item := &LR0Item{
		prod:      prod,
		dot:       0,
		initial:   true,
		reducible: false,
	}
	item.fingerprint = generateLR0ItemFingerprint(item)

	return item, nil
}

func (i *LR0Item) String() string {
	s := i.prod.lhs.String() + " →"
	for n, sym := range i.prod.rhs {
		if n == i.dot {
			s += "・"
		} else {
			s += " "
		}
		s += sym.String()
	}
	if i.reducible {
		s += "・#"
	}

	return s
}

func (i *LR0Item) IsKernel() bool {
	return i.initial || i.dot > 0
}

type KernelFingerprint string

const (
	kernelFingerprintNil = KernelFingerprint("")
)

func (fp KernelFingerprint) String() string {
	return string(fp)
}

func (fp KernelFingerprint) IsNil() bool {
	return fp == kernelFingerprintNil
}

type KernelItems struct {
	memoizedFingerprint KernelFingerprint
	items               map[LR0ItemFingerprint]*LR0Item
}

func NewKernelItems() *KernelItems {
	return &KernelItems{
		memoizedFingerprint: "",
		items:               map[LR0ItemFingerprint]*LR0Item{},
	}
}

func (k *KernelItems) Append(i *LR0Item) error {
	if !i.IsKernel() {
		return fmt.Errorf("the non-kernel item was about to be added to the kernel item list")
	}

	k.items[i.fingerprint] = i

	return nil
}

func (k *KernelItems) Fingerprint() KernelFingerprint {
	if k.memoizedFingerprint != "" {
		return k.memoizedFingerprint
	}

	if len(k.items) <= 0 {
		return kernelFingerprintNil
	}

	s := []*LR0Item{}
	for _, i := range k.items {
		s = append(s, i)
	}
	sort.SliceStable(s, func(i, j int) bool {
		i1 := s[i]
		i2 := s[j]

		if i1.prod.lhs == i2.prod.lhs {
			return i1.dot < i2.dot
		}
		return i1.prod.lhs < i2.prod.lhs
	})

	fp := ""
	for _, i := range s {
		if fp != "" {
			fp += "/"
		}
		if i.reducible {
			fp += fmt.Sprintf("%v-#", i.prod.fingerprint)
		} else {
			fp += fmt.Sprintf("%v-%v", i.prod.fingerprint, i.dot)
		}
	}

	return KernelFingerprint(fp)
}

type LR0ItemSet struct {
	Fingerprint KernelFingerprint
	Items       map[LR0ItemFingerprint]*LR0Item
	GoTo        map[SymbolID]KernelFingerprint
}

func NewLR0ItemSet(k *KernelItems) (*LR0ItemSet, error) {
	if k == nil {
		return nil, fmt.Errorf("a set of items doesn't create without kernel items")
	}

	fp := k.Fingerprint()
	if fp.IsNil() {
		return nil, fmt.Errorf("the fingerprint of the kernel is nil")
	}

	is := &LR0ItemSet{
		Fingerprint: fp,
		Items:       map[LR0ItemFingerprint]*LR0Item{},
		GoTo:        map[SymbolID]KernelFingerprint{},
	}
	for fp, i := range k.items {
		is.Items[fp] = i
	}

	return is, nil
}

func (is *LR0ItemSet) ComputeClosure(st *SymbolTable, prods Productions) error {
	uncheckedItems := map[LR0ItemFingerprint]*LR0Item{}
	for fp, i := range is.Items {
		uncheckedItems[fp] = i
	}

	for len(uncheckedItems) > 0 {
		nextUncheckedItems := map[LR0ItemFingerprint]*LR0Item{}

		for _, item := range uncheckedItems {
			if item.reducible {
				continue
			}

			nextSym := item.prod.rhs[item.dot]
			nextSymKind := nextSym.Kind()
			if nextSymKind.IsNil() {
				return fmt.Errorf("invalid symbol")
			}

			if !nextSymKind.IsNonTerminalSymbol() {
				continue
			}

			for _, prod := range prods.Get(nextSym) {
				newItem, err := NewLR0Item(prod, 0)
				if err != nil {
					return err
				}
				if _, exist := is.Items[newItem.fingerprint]; exist {
					continue
				}
				is.Items[newItem.fingerprint] = newItem
				nextUncheckedItems[newItem.fingerprint] = newItem

			}
		}

		uncheckedItems = nextUncheckedItems
	}

	return nil
}

type LR0Automaton struct {
	states map[KernelFingerprint]*LR0ItemSet
}

func GenerateLR0Automaton(st *SymbolTable, prods Productions, expansionStartSymbol SymbolID) (*LR0Automaton, error) {
	if st == nil {
		return nil, fmt.Errorf("symbol table passed is nil")
	}
	if expansionStartSymbol.IsNil() || !expansionStartSymbol.Kind().IsStartSymbol() {
		return nil, fmt.Errorf("symbold passed is nil or not start symbol")
	}

	automaton := &LR0Automaton{
		states: map[KernelFingerprint]*LR0ItemSet{},
	}

	// append the initial item to automaton.states
	{
		i0Kernel := NewKernelItems()
		initialItem, err := NewInitialLR0Item(prods.Get(expansionStartSymbol)[0])
		if err != nil {
			return nil, err
		}
		err = i0Kernel.Append(initialItem)
		if err != nil {
			return nil, err
		}

		i0, err := NewLR0ItemSet(i0Kernel)
		if err != nil {
			return nil, err
		}

		automaton.states[i0.Fingerprint] = i0
	}

	uncheckedStates := map[KernelFingerprint]*LR0ItemSet{}
	for fp, is := range automaton.states {
		uncheckedStates[fp] = is
	}

	for len(uncheckedStates) > 0 {
		nextUncheckedStates := map[KernelFingerprint]*LR0ItemSet{}

		for _, state := range uncheckedStates {
			state.ComputeClosure(st, prods)

			kernelMap := map[SymbolID]*KernelItems{}
			for _, item := range state.Items {
				if item.reducible {
					continue
				}

				kItem, err := NewLR0Item(item.prod, item.dot+1)
				if err != nil {
					return nil, err
				}

				nextSym := item.prod.rhs[item.dot]
				if k, ok := kernelMap[nextSym]; ok {
					k.Append(kItem)
				} else {
					k := NewKernelItems()
					k.Append(kItem)
					kernelMap[nextSym] = k
				}
			}

			for nextSym, kItems := range kernelMap {
				is, err := NewLR0ItemSet(kItems)
				if err != nil {
					return nil, err
				}

				if _, exist := automaton.states[is.Fingerprint]; !exist {
					automaton.states[is.Fingerprint] = is
					nextUncheckedStates[is.Fingerprint] = is
				}

				state.GoTo[nextSym] = is.Fingerprint
			}
		}

		uncheckedStates = nextUncheckedStates
	}

	return automaton, nil
}
