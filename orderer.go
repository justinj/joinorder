package main

// Relations are ordered from 1.
type RelationID int

type Selectivity float64
type Cardinality int

type Sequence []RelationID

type Orderer struct {
	numRelations  int
	cardinalities []Cardinality
	selectivities []Selectivity
}

func NewOrderer(numRelations int) *Orderer {
	selectivities := make([]Selectivity, numRelations*(numRelations-1)/2)
	for i := range selectivities {
		selectivities[i] = -1
	}
	return &Orderer{
		numRelations:  numRelations,
		selectivities: selectivities,
		cardinalities: make([]Cardinality, numRelations),
	}
}

func (o *Orderer) GetSelectivity(a, b RelationID) Selectivity {
	sel := o.selectivities[o.indexPair(a, b)]
	if sel == -1 {
		return 1
	}
	return sel
}

func (o *Orderer) indexPair(a, b RelationID) int {
	if a == b {
		panic("can't index with self")
	}
	if a > b {
		a, b = b, a
	}
	a--
	b--
	return int(b*(b-1)/2 + a)
}

func (o *Orderer) Adjacent(a, b RelationID) bool {
	ab := o.selectivities[o.indexPair(a, b)]
	return ab != -1
}

func (o *Orderer) AddPredicate(a, b RelationID, sel Selectivity) {
	idx := o.indexPair(a, b)

	if o.selectivities[idx] == -1 {
		o.selectivities[idx] = 1
	}
	o.selectivities[idx] *= sel
}

func (o *Orderer) SetCardinality(a RelationID, c Cardinality) {
	o.cardinalities[a-1] = c
}

func (o *Orderer) Cardinality(a RelationID) Cardinality {
	return o.cardinalities[a-1]
}

func (o *Orderer) BruteForceOrder() Sequence {
	var best Sequence
	var bestCost float64

	start := make(Sequence, o.numRelations)
	for i := 0; i < o.numRelations; i++ {
		start[i] = RelationID(i + 1)
	}

	Perm(start, func(ord Sequence) {
		cost := o.Cost(ord)
		if bestCost == 0 || cost < bestCost {
			best = best[:0]
			best = append(best, ord...)
			bestCost = cost
		}
	})

	return best
}

func (o *Orderer) Cost(ord Sequence) float64 {
	cost := float64(o.Cardinality(ord[0]))
	numRows := float64(o.Cardinality(ord[0]))

	for i := 1; i < len(ord); i++ {
		// Calculate selectivity of this relation with all
		// previous relations.
		for j := 0; j < i; j++ {
			numRows *= float64(o.GetSelectivity(ord[i], ord[j]))
		}
		numRows *= float64(o.Cardinality(ord[i]))
		cost += numRows
	}
	return cost
}
