package grammar

import (
	"fmt"
)

type ProductionFingerprint string

func newProductionFingerprint(lhs SymbolID, rhs []SymbolID) ProductionFingerprint {
	rhsFp := ""
	for _, sym := range rhs {
		if rhsFp != "" {
			rhsFp += ","
		}
		rhsFp += sym.String()
	}

	return ProductionFingerprint(fmt.Sprintf("(%s->%s)", lhs, rhsFp))
}

func (fp ProductionFingerprint) String() string {
	return string(fp)
}

type Production struct {
	fingerprint ProductionFingerprint
	lhs         SymbolID
	rhs         []SymbolID
}

func NewProduction(lhs SymbolID, rhs []SymbolID) (*Production, error) {
	if lhs.IsNil() {
		return nil, fmt.Errorf("symbol id of lhs is nil")
	}

	for _, sym := range rhs {
		if sym.IsNil() {
			return nil, fmt.Errorf("rhs contains nil symbol id")
		}
	}

	return &Production{
		fingerprint: newProductionFingerprint(lhs, rhs),
		lhs:         lhs,
		rhs:         rhs,
	}, nil
}

func (prod *Production) String() string {
	rhs := ""
	if len(prod.rhs) > 0 {
		for _, sym := range prod.rhs {
			if rhs != "" {
				rhs += " "
			}
			rhs += sym.String()
		}
	} else {
		rhs = "ε"
	}

	return fmt.Sprintf("%v → %v", prod.lhs, rhs)
}
