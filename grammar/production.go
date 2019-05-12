package grammar

import (
	"fmt"
)

type Productions map[SymbolID][]*Production

func NewProductions() Productions {
	return map[SymbolID][]*Production{}
}

func (prods Productions) Append(prod *Production) {
	if arr, ok := prods[prod.lhs]; ok {
		prods[prod.lhs] = append(arr, prod)
	} else {
		prods[prod.lhs] = []*Production{prod}
	}
}

func (prods Productions) Get(lhs SymbolID) []*Production {
	if lhs.IsNil() {
		return nil
	}

	return prods[lhs]
}

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
	rhsLen      int
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
		rhsLen:      len(rhs),
	}, nil
}

func (prod *Production) Equal(target *Production) bool {
	return prod.fingerprint == target.fingerprint
}

func (prod *Production) isEmpty() bool {
	return prod.rhsLen <= 0
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
