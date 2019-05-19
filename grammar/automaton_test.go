package grammar

import (
	"testing"
)

func TestGenerateLR0Automaton(t *testing.T) {
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

	automaton, err := GenerateLR0Automaton(st, prods, V("E'"))
	if err != nil {
		t.Fatal(err)
	}

	{
		k, err := genInitialKernel(V("E'"), st, prods)
		if err != nil {
			t.Fatal(err)
		}

		if automaton.initialState != k.Fingerprint() {
			t.Fatalf("unexpected initial state\nwant: %v\ngot: %v", k.Fingerprint(), automaton.initialState)
		}
	}

	tests := map[int]struct {
		kernels    []lr0Item
		others     []lr0Item
		nextStates map[string]int
	}{
		0: {
			kernels: []lr0Item{
				{lhs: "E'", num: 0, initial: true},
			},
			others: []lr0Item{
				{lhs: "E", num: 0, dot: 0},
				{lhs: "E", num: 1, dot: 0},
				{lhs: "T", num: 0, dot: 0},
				{lhs: "T", num: 1, dot: 0},
				{lhs: "F", num: 0, dot: 0},
				{lhs: "F", num: 1, dot: 0},
			},
			nextStates: map[string]int{
				"E":  1,
				"T":  2,
				"F":  3,
				"(":  4,
				"id": 5,
			},
		},
		1: {
			kernels: []lr0Item{
				{lhs: "E'", num: 0, dot: 1},
				{lhs: "E", num: 0, dot: 1},
			},
			others: []lr0Item{},
			nextStates: map[string]int{
				"+": 6,
			},
		},
		2: {
			kernels: []lr0Item{
				{lhs: "E", num: 1, dot: 1},
				{lhs: "T", num: 0, dot: 1},
			},
			others: []lr0Item{},
			nextStates: map[string]int{
				"*": 7,
			},
		},
		3: {
			kernels: []lr0Item{
				{lhs: "T", num: 1, dot: 1},
			},
			others:     []lr0Item{},
			nextStates: map[string]int{},
		},
		4: {
			kernels: []lr0Item{
				{lhs: "F", num: 0, dot: 1},
			},
			others: []lr0Item{
				{lhs: "E", num: 0, dot: 0},
				{lhs: "E", num: 1, dot: 0},
				{lhs: "T", num: 0, dot: 0},
				{lhs: "T", num: 1, dot: 0},
				{lhs: "F", num: 0, dot: 0},
				{lhs: "F", num: 1, dot: 0},
			},
			nextStates: map[string]int{
				"T":  2,
				"F":  3,
				"(":  4,
				"id": 5,
				"E":  8,
			},
		},
		5: {
			kernels: []lr0Item{
				{lhs: "F", num: 1, dot: 1},
			},
			others:     []lr0Item{},
			nextStates: map[string]int{},
		},
		6: {
			kernels: []lr0Item{
				{lhs: "E", num: 0, dot: 2},
			},
			others: []lr0Item{
				{lhs: "T", num: 0, dot: 0},
				{lhs: "T", num: 1, dot: 0},
				{lhs: "F", num: 0, dot: 0},
				{lhs: "F", num: 1, dot: 0},
			},
			nextStates: map[string]int{
				"F":  3,
				"(":  4,
				"id": 5,
				"T":  9,
			},
		},
		7: {
			kernels: []lr0Item{
				{lhs: "T", num: 0, dot: 2},
			},
			others: []lr0Item{
				{lhs: "F", num: 0, dot: 0},
				{lhs: "F", num: 1, dot: 0},
			},
			nextStates: map[string]int{
				"(":  4,
				"id": 5,
				"F":  10,
			},
		},
		8: {
			kernels: []lr0Item{
				{lhs: "E", num: 0, dot: 1},
				{lhs: "F", num: 0, dot: 2},
			},
			others: []lr0Item{},
			nextStates: map[string]int{
				"+": 6,
				")": 11,
			},
		},
		9: {
			kernels: []lr0Item{
				{lhs: "E", num: 0, reducible: true},
				{lhs: "T", num: 0, dot: 1},
			},
			others: []lr0Item{},
			nextStates: map[string]int{
				"*": 7,
			},
		},
		10: {
			kernels: []lr0Item{
				{lhs: "T", num: 0, reducible: true},
			},
			others:     []lr0Item{},
			nextStates: map[string]int{},
		},
		11: {
			kernels: []lr0Item{
				{lhs: "F", num: 0, reducible: true},
			},
			others:     []lr0Item{},
			nextStates: map[string]int{},
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
			t.Error("the fingerprint of the kernel is nil")
			continue
		}

		state, ok := automaton.states[kernelFp]
		if !ok {
			t.Errorf("failed to get a state %v", kernelFp)
			continue
		}

		for _, i := range tt.others {
			item, err := genLR0Item(i, st, prods)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := state.Items[item.fingerprint]; !ok {
				t.Fatal("invalid state")
			}
		}

		if len(state.Items) != len(tt.kernels)+len(tt.others) {
			t.Errorf("invalid state\nwant: %v items\ngot: %v items", len(state.Items), len(tt.kernels)+len(tt.others))
			continue
		}

		for sym, next := range tt.nextStates {
			nextTt, ok := tests[next]
			if !ok {
				t.Error("invalid tests")
				continue
			}

			k, err := genKernel(nextTt.kernels, st, prods)
			if err != nil {
				t.Error(err)
				continue
			}

			kernelFp := k.Fingerprint()
			if kernelFp.IsNil() {
				t.Error("the fingerprint of the kernel is nil")
				continue
			}

			if state.GoTo[V(sym)] != kernelFp {
				t.Errorf("invalid goto %v", kernelFp)
				continue
			}
		}

		if len(state.GoTo) != len(tt.nextStates) {
			t.Errorf("invalid goto\nwant: %v items\ngot: %v items", len(tt.nextStates), len(state.GoTo))
		}
	}

	if len(automaton.states) != len(tests) {
		t.Fatalf("number of states is missmatch\nwant: %v states\ngot: %v states", len(tests), len(automaton.states))
	}
}

func TestKernelItems(t *testing.T) {
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

	P := newProductionGetter(st, prods)

	t.Run("", func(t *testing.T) {
		tests := []struct {
			kernels        []lr0Item
			nilFingerprint bool
		}{
			// KernelItems contains only the start rule.
			{
				kernels: []lr0Item{
					{lhs: "E'", num: 0, initial: true},
				},
				nilFingerprint: false,
			},
			// KernelItems contains just one rule.
			{
				kernels: []lr0Item{
					{lhs: "T", num: 1, reducible: true},
				},
				nilFingerprint: false,
			},
			// KernelItems contains some rules.
			{
				kernels: []lr0Item{
					{lhs: "E", num: 1, reducible: true},
					{lhs: "T", num: 1, dot: 1},
				},
				nilFingerprint: false,
			},
			// KernelItems contains no item.
			{
				kernels:        []lr0Item{},
				nilFingerprint: true,
			},
		}

		for _, tt := range tests {
			k, err := genKernel(tt.kernels, st, prods)
			if err != nil {
				t.Error(err)
				continue
			}

			fp := k.Fingerprint()
			if tt.nilFingerprint {
				if !fp.IsNil() {
					t.Error("fingerprint is not nil")
					continue
				}
			} else {
				if fp.IsNil() {
					t.Fatal("fingerprint is nil")
					continue
				}
			}
		}
	})

	t.Run("non-kernel item append to KernelItems", func(t *testing.T) {
		kernelItems := NewKernelItems()

		item, err := NewInitialLR0Item(P("E'", 0))
		if err != nil {
			t.Fatal(err)
		}
		err = kernelItems.Append(item)
		if err != nil {
			t.Fatal(err)
		}

		item, err = NewLR0Item(P("E", 0), 0)
		if err != nil {
			t.Fatal(err)
		}
		err = kernelItems.Append(item)
		if err == nil {
			t.Error("no error was returned")
		}
	})
}
