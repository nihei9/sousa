package grammar

import (
	"fmt"
	"strconv"
)

type ProductionID int

func (id ProductionID) String() string {
	return strconv.Itoa(int(id))
}

type productionIDGenerator struct {
	id ProductionID
}

func newProductionIDGenerator() *productionIDGenerator {
	return &productionIDGenerator{
		id: ProductionID(0),
	}
}

func (gen *productionIDGenerator) next() ProductionID {
	id := gen.id
	gen.id = ProductionID(int(gen.id) + 1)
	return id
}

type Productions interface {
	Append(*Production)
	Get(SymbolID) []*Production
	LookupByFingerprint(ProductionFingerprint) *Production
	All() map[SymbolID][]*Production
}

// production is an inplementation of the Productions interface.
type productions struct {
	prods   map[SymbolID][]*Production
	fp2prod map[ProductionFingerprint]*Production
	idGen   *productionIDGenerator
}

func NewProductions() Productions {
	return &productions{
		prods:   map[SymbolID][]*Production{},
		fp2prod: map[ProductionFingerprint]*Production{},
		idGen:   newProductionIDGenerator(),
	}
}

func (prods *productions) Append(prod *Production) {
	prod.id = prods.idGen.next()
	if arr, ok := prods.prods[prod.lhs]; ok {
		prods.prods[prod.lhs] = append(arr, prod)
	} else {
		prods.prods[prod.lhs] = []*Production{prod}
	}
	prods.fp2prod[prod.fingerprint] = prod
}

func (prods *productions) Get(lhs SymbolID) []*Production {
	if lhs.IsNil() {
		return nil
	}

	return prods.prods[lhs]
}

func (prods *productions) LookupByFingerprint(fp ProductionFingerprint) *Production {
	return prods.fp2prod[fp]
}

func (prods *productions) All() map[SymbolID][]*Production {
	return prods.prods
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

func (fp ProductionFingerprint) IsNil() bool {
	return string(fp) == ""
}

type Production struct {
	id          ProductionID
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

func (prod *Production) ID() ProductionID {
	return prod.id
}

func (prod *Production) RHS() ([]SymbolID, int) {
	return prod.rhs, prod.rhsLen
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
