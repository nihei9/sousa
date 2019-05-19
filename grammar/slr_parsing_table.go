package grammar

import (
	"fmt"
)

type ActionType int

const (
	ActionTypeShift  = 0
	ActionTypeReduce = 1
)

func (at ActionType) String() string {
	switch at {
	case ActionTypeShift:
		return "shift"
	case ActionTypeReduce:
		return "reduce"
	}

	return ""
}

type Action struct {
	t         ActionType
	nextState KernelFingerprint
	prod      ProductionFingerprint
}

type Actions struct {
	actions     map[SymbolID]*Action
	acceptable  bool
	reduceByEOF ProductionFingerprint
}

type ParsingTable struct {
	action map[KernelFingerprint]*Actions
	goTo   map[KernelFingerprint]map[SymbolID]KernelFingerprint
}

func newParsingTable() *ParsingTable {
	return &ParsingTable{
		action: map[KernelFingerprint]*Actions{},
		goTo:   map[KernelFingerprint]map[SymbolID]KernelFingerprint{},
	}
}

func (pt *ParsingTable) appendShiftAction(state KernelFingerprint, sym SymbolID, nextState KernelFingerprint) error {
	a := &Action{
		t:         ActionTypeShift,
		nextState: nextState,
	}
	return pt.appendAction(a, state, sym)
}

func (pt *ParsingTable) appendReduceAction(state KernelFingerprint, sym SymbolID, prod ProductionFingerprint) error {
	a := &Action{
		t:    ActionTypeReduce,
		prod: prod,
	}
	return pt.appendAction(a, state, sym)
}

func (pt *ParsingTable) appendReduceActionByEOF(state KernelFingerprint, prod ProductionFingerprint) {
	if _, ok := pt.action[state]; !ok {
		pt.action[state] = &Actions{
			actions:    map[SymbolID]*Action{},
			acceptable: false,
		}
	}

	pt.action[state].reduceByEOF = prod
}

func (pt *ParsingTable) appendAcceptAction(state KernelFingerprint) {
	if _, ok := pt.action[state]; !ok {
		pt.action[state] = &Actions{
			actions:    map[SymbolID]*Action{},
			acceptable: false,
		}
	}

	pt.action[state].acceptable = true
}

func (pt *ParsingTable) appendAction(a *Action, state KernelFingerprint, sym SymbolID) error {
	if !sym.Kind().IsTerminalSymbol() {
		return fmt.Errorf("a non-terminal symbol cannot append to ACTION. state: %v, symbol: %v", state, sym)
	}

	if _, ok := pt.action[state]; !ok {
		pt.action[state] = &Actions{
			actions:    map[SymbolID]*Action{},
			acceptable: false,
		}
	}

	pt.action[state].actions[sym] = a

	return nil
}

func (pt *ParsingTable) appendGoTo(state KernelFingerprint, sym SymbolID, nextState KernelFingerprint) error {
	if !sym.Kind().IsNonTerminalSymbol() {
		return fmt.Errorf("a terminal symbol cannot append to GOTO. state: %v, symbol: %v, next state: %v", state, sym, nextState)
	}

	if _, ok := pt.goTo[state]; !ok {
		pt.goTo[state] = map[SymbolID]KernelFingerprint{}
	}

	pt.goTo[state][sym] = nextState

	return nil
}

func GenerateSLRParsingTable(automaton *LR0Automaton, follow FollowSets) (*ParsingTable, error) {
	if automaton == nil || follow == nil {
		return nil, fmt.Errorf("parameters passed contains nil")
	}

	pt := newParsingTable()

	for _, state := range automaton.states {
		for _, item := range state.Items {
			if item.reducible {
				if item.prod.lhs.Kind().IsStartSymbol() {
					pt.appendAcceptAction(state.Fingerprint)
				} else {
					syms := follow.Get(item.prod.lhs)
					for sym, _ := range syms.symbols {
						err := pt.appendReduceAction(state.Fingerprint, sym, item.prod.fingerprint)
						if err != nil {
							return nil, err
						}
					}
					if syms.eof {
						pt.appendReduceActionByEOF(state.Fingerprint, item.prod.fingerprint)
					}
				}
			} else {
				sym := item.prod.rhs[item.dot]
				if !sym.Kind().IsTerminalSymbol() {
					continue
				}

				nextState, ok := state.GoTo[sym]
				if !ok {
					return nil, fmt.Errorf("next status not found")
				}
				err := pt.appendShiftAction(state.Fingerprint, sym, nextState)
				if err != nil {
					return nil, err
				}
			}
		}

		for sym, nextState := range state.GoTo {
			if sym.Kind().IsNonTerminalSymbol() {
				err := pt.appendGoTo(state.Fingerprint, sym, nextState)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return pt, nil
}
