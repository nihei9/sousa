package writer

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/nihei9/sousa/grammar"
)

type Writer interface {
	Write(io.Writer) error
}

type productionsWriter struct {
	productions grammar.Productions
}

func NewProductionsWriter(productions grammar.Productions) Writer {
	return &productionsWriter{
		productions: productions,
	}
}

func (pw *productionsWriter) Write(w io.Writer) error {
	lhss := make([]grammar.SymbolID, len(pw.productions.All()))
	{
		i := 0
		for lhs, _ := range pw.productions.All() {
			lhss[i] = lhs
			i++
		}
		sort.SliceStable(lhss, func(i, j int) bool {
			return lhss[i] < lhss[j]
		})
	}

	for _, lhs := range lhss {
		buf := new(bytes.Buffer)
		for _, prod := range pw.productions.Get(lhs) {
			id := prod.ID()
			_, rhsLen := prod.RHS()
			fmt.Fprintf(buf, "%v,%v,%v\n", id, lhs, rhsLen)
		}
		w.Write(buf.Bytes())
	}

	return nil
}

type actionWriter struct {
	parsingTable *grammar.ParsingTable
	productions  grammar.Productions
}

func NewActionWriter(parsingTable *grammar.ParsingTable, productions grammar.Productions) Writer {
	return &actionWriter{
		parsingTable: parsingTable,
		productions:  productions,
	}
}

func (aw *actionWriter) Write(w io.Writer) error {
	states := aw.parsingTable.States()
	for kernelFp, actions := range aw.parsingTable.Action() {
		buf := new(bytes.Buffer)

		state, ok := states[kernelFp]
		if !ok {
			return fmt.Errorf("failed to get a state. kernel fingerprint: %v", kernelFp)
		}
		fmt.Fprintf(buf, "%v", state)

		if actions.Acceptable() {
			fmt.Fprint(buf, ",t")
		} else {
			fmt.Fprint(buf, ",f")
		}

		if prodFp, reducible := actions.ReduceByEOF(); reducible {
			prod := aw.productions.LookupByFingerprint(prodFp)
			if prod == nil {
				return fmt.Errorf("failed to get a production. fingerprint: %v", prodFp)
			}
			fmt.Fprintf(buf, ",$-r%v", prod.ID())
		}

		for sym, a := range actions.Actions() {
			switch a.Type() {
			case grammar.ActionTypeShift:
				fmt.Fprintf(buf, ",%v-s%v", sym, states[a.NextState()])
			case grammar.ActionTypeReduce:
				prod := aw.productions.LookupByFingerprint(a.Production())
				if prod == nil {
					return fmt.Errorf("failed to get a production. fingerprint: %v", a.Production())
				}
				fmt.Fprintf(buf, ",%v-r%v", sym, prod.ID())
			default:
				return fmt.Errorf("unknown action type. got: %v", a.Type())
			}
		}

		fmt.Fprint(buf, "\n")

		w.Write(buf.Bytes())
	}

	return nil
}

type goToWriter struct {
	parsingTable *grammar.ParsingTable
}

func NewGoToWriter(parsingTable *grammar.ParsingTable) Writer {
	return &goToWriter{
		parsingTable: parsingTable,
	}
}

func (gw *goToWriter) Write(w io.Writer) error {
	states := gw.parsingTable.States()
	for kernelFp, goTos := range gw.parsingTable.GoTo() {
		buf := new(bytes.Buffer)

		state, ok := states[kernelFp]
		if !ok {
			return fmt.Errorf("failed to get a state. kernel fingerprint: %v", kernelFp)
		}
		fmt.Fprintf(buf, "%v", state)

		for sym, kernelFp := range goTos {
			state, ok := states[kernelFp]
			if !ok {
				return fmt.Errorf("failed to get a state. kernel fingerprint: %v", kernelFp)
			}
			fmt.Fprintf(buf, ",%v-%v", sym, state)
		}

		fmt.Fprint(buf, "\n")

		w.Write(buf.Bytes())
	}

	return nil
}
