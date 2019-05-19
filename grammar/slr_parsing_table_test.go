package grammar

import (
	"testing"
)

type Act struct {
	t         ActionType
	nextState int
	prod      ProductionFingerprint
}

func shift(nextState int) Act {
	return Act{
		t:         ActionTypeShift,
		nextState: nextState,
	}
}

func reduce(prod *Production) Act {
	return Act{
		t:    ActionTypeReduce,
		prod: prod.fingerprint,
	}
}

func TestParsingTable(t *testing.T) {
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

	V := newSymbolGetter(st)
	P := newProductionGetter(st, prods)

	first, err := GenerateFirstSets(prods)
	if err != nil {
		t.Fatal(err)
	}

	follow, err := GenerateFollowSets(prods, first)
	if err != nil {
		t.Fatal(err)
	}

	automaton, err := GenerateLR0Automaton(st, prods, st.lookupByString("E'"))
	if err != nil {
		t.Fatal(err)
	}

	slrPT, err := GenerateSLRParsingTable(automaton, follow)
	if err != nil {
		t.Fatal(err)
	}
	if slrPT == nil {
		t.Fatal("GenerateSLRParsingTable() returned nil without an error")
	}

	tests := map[int]struct {
		kernels            []lr0Item
		action             map[string]Act
		reducibleByEOF     bool
		reduceByEOFActtion Act
		acceptable         bool
		goTo               map[string]int
	}{
		0: {
			kernels: []lr0Item{
				{lhs: "E'", num: 0, initial: true},
			},
			action: map[string]Act{
				"id": shift(5),
				"(":  shift(4),
			},
			goTo: map[string]int{
				"E": 1,
				"T": 2,
				"F": 3,
			},
		},
		1: {
			kernels: []lr0Item{
				{lhs: "E'", num: 0, dot: 1},
				{lhs: "E", num: 0, dot: 1},
			},
			action: map[string]Act{
				"+": shift(6),
			},
			acceptable: true,
			goTo:       map[string]int{},
		},
		2: {
			kernels: []lr0Item{
				{lhs: "E", num: 1, dot: 1},
				{lhs: "T", num: 0, dot: 1},
			},
			action: map[string]Act{
				"+": reduce(P("E", 1)),
				"*": shift(7),
				")": reduce(P("E", 1)),
			},
			reducibleByEOF:     true,
			reduceByEOFActtion: reduce(P("E", 1)),
			goTo:               map[string]int{},
		},
		3: {
			kernels: []lr0Item{
				{lhs: "T", num: 1, dot: 1},
			},
			action: map[string]Act{
				"+": reduce(P("T", 1)),
				"*": reduce(P("T", 1)),
				")": reduce(P("T", 1)),
			},
			reducibleByEOF:     true,
			reduceByEOFActtion: reduce(P("T", 1)),
			goTo:               map[string]int{},
		},
		4: {
			kernels: []lr0Item{
				{lhs: "F", num: 0, dot: 1},
			},
			action: map[string]Act{
				"id": shift(5),
				"(":  shift(4),
			},
			goTo: map[string]int{
				"E": 8,
				"T": 2,
				"F": 3,
			},
		},
		5: {
			kernels: []lr0Item{
				{lhs: "F", num: 1, dot: 1},
			},
			action: map[string]Act{
				"+": reduce(P("F", 1)),
				"*": reduce(P("F", 1)),
				")": reduce(P("F", 1)),
			},
			reducibleByEOF:     true,
			reduceByEOFActtion: reduce(P("F", 1)),
			goTo:               map[string]int{},
		},
		6: {
			kernels: []lr0Item{
				{lhs: "E", num: 0, dot: 2},
			},
			action: map[string]Act{
				"id": shift(5),
				"(":  shift(4),
			},
			goTo: map[string]int{
				"T": 9,
				"F": 3,
			},
		},
		7: {
			kernels: []lr0Item{
				{lhs: "T", num: 0, dot: 2},
			},
			action: map[string]Act{
				"id": shift(5),
				"(":  shift(4),
			},
			goTo: map[string]int{
				"F": 10,
			},
		},
		8: {
			kernels: []lr0Item{
				{lhs: "E", num: 0, dot: 1},
				{lhs: "F", num: 0, dot: 2},
			},
			action: map[string]Act{
				"+": shift(6),
				")": shift(11),
			},
			goTo: map[string]int{},
		},
		9: {
			kernels: []lr0Item{
				{lhs: "E", num: 0, reducible: true},
				{lhs: "T", num: 0, dot: 1},
			},
			action: map[string]Act{
				"+": reduce(P("E", 0)),
				"*": shift(7),
				")": reduce(P("E", 0)),
			},
			reducibleByEOF:     true,
			reduceByEOFActtion: reduce(P("E", 0)),
			goTo:               map[string]int{},
		},
		10: {
			kernels: []lr0Item{
				{lhs: "T", num: 0, reducible: true},
			},
			action: map[string]Act{
				"+": reduce(P("T", 0)),
				"*": reduce(P("T", 0)),
				")": reduce(P("T", 0)),
			},
			reducibleByEOF:     true,
			reduceByEOFActtion: reduce(P("T", 0)),
			goTo:               map[string]int{},
		},
		11: {
			kernels: []lr0Item{
				{lhs: "F", num: 0, reducible: true},
			},
			action: map[string]Act{
				"+": reduce(P("F", 0)),
				"*": reduce(P("F", 0)),
				")": reduce(P("F", 0)),
			},
			reducibleByEOF:     true,
			reduceByEOFActtion: reduce(P("F", 0)),
			goTo:               map[string]int{},
		},
	}

	for _, tt := range tests {
		k, err := genKernel(tt.kernels, st, prods)
		if err != nil {
			t.Error(err)
			continue
		}

		kernelFp := k.Fingerprint()
		if kernelFp.IsNil() {
			t.Errorf("the fingerprint of the kernel is nil\ntest: %+v", tt)
			continue
		}

		actualActions, ok := slrPT.action[kernelFp]
		if !ok {
			t.Errorf("failed to get actions. state: %v\ntest: %+v", kernelFp, tt)
			continue
		}

		if actualActions.acceptable != tt.acceptable {
			t.Errorf("acceptable is mismatched\nwant: %v\ngot: %v\ntest: %+v", tt.acceptable, actualActions.acceptable, tt)
		}

		for sym, act := range tt.action {
			a, ok := actualActions.actions[V(sym)]
			if !ok {
				t.Errorf("failed to get an action. state: %v, symbol: %v\ntest: %+v", kernelFp, sym, tt)
				continue
			}

			if a.t != act.t {
				t.Errorf("action type is mismatched\nwant: %v\ngot: %v\ntest: %+v", act.t, a.t, tt)
				continue
			}

			if act.t == ActionTypeShift {
				k, err := genKernel(tests[act.nextState].kernels, st, prods)
				if err != nil {
					t.Error(err)
					continue
				}

				if a.nextState != k.Fingerprint() {
					t.Errorf("invalid next state\nwant: %v\ngot: %v\ntest: %+v", act.nextState, a.nextState, tt)
					continue
				}
			} else if act.t == ActionTypeReduce {

				if a.prod != act.prod {
					t.Errorf("invalid production\nwant: %v\ngot: %v\ntest: %+v", act.prod, a.prod, tt)
					continue
				}
			}
		}

		if len(actualActions.actions) != len(tt.action) {
			t.Errorf("invalid action\nwant: %v item(s)\ngot: %v item(s)\ntest: %+v", len(tt.action), len(actualActions.actions), tt)
			continue
		}

		if tt.reducibleByEOF {
			eProd := tt.reduceByEOFActtion.prod
			aProd := actualActions.reduceByEOF
			if aProd != eProd {
				t.Errorf("production is mismatched\nwant: %v\ngot: %v\ntest: %+v", eProd, aProd, tt)
			}
			continue
		}

		if len(tt.goTo) > 0 {
			actualGoTos, ok := slrPT.goTo[kernelFp]
			if !ok {
				t.Errorf("failed to get gotos. state: %v\ntest: %+v", kernelFp, tt)
				continue
			}

			for sym, goTo := range tt.goTo {
				aKernelFp, ok := actualGoTos[V(sym)]
				if !ok {
					t.Errorf("failed to get a goto. state: %v, symbol: %v\ntest: %+v", kernelFp, sym, tt)
					continue
				}

				k, err := genKernel(tests[goTo].kernels, st, prods)
				if err != nil {
					t.Error(err)
					continue
				}

				eKernelFp := k.Fingerprint()
				if eKernelFp.IsNil() {
					t.Errorf("the fingerprint of the kernel is nil\ntest: %+v", tt)
					continue
				}

				if aKernelFp != eKernelFp {
					t.Errorf("invalid next state\nwant: %v\ngot: %v\ntest: %+v", eKernelFp, aKernelFp, tt)
					continue
				}
			}

			if len(actualGoTos) != len(tt.goTo) {
				t.Errorf("invalid goto\nwant: %v item(s)\ngot: %v item(s)\ntest: %+v", len(tt.goTo), len(actualGoTos), tt)
				continue
			}
		}
	}
}
