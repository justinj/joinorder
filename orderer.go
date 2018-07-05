package main

// Relations are ordered from 1.
type RelationID int

type Selectivity float64
type Cardinality int

type Sequence []RelationID

type Orderer struct {
	numRelations  int
	cardinalities []Cardinality
	selectivities [][]Selectivity
}

func NewOrderer(numRelations int) *Orderer {
	selectivities := make([][]Selectivity, numRelations)
	for i := range selectivities {
		selectivities[i] = make([]Selectivity, numRelations)
		for j := range selectivities[i] {
			selectivities[i][j] = -1
		}
	}
	return &Orderer{
		numRelations:  numRelations,
		selectivities: selectivities,
		cardinalities: make([]Cardinality, numRelations),
	}
}

func (o *Orderer) GetSelectivity(a, b RelationID) Selectivity {
	ab := o.selectivities[a-1][b-1]
	ba := o.selectivities[b-1][a-1]
	if ab != ba {
		panic("selectivity should be commutative")
	}
	if ab == -1 {
		return 1
	}
	return ab
}

func (o *Orderer) Adjacent(a, b RelationID) bool {
	ab := o.selectivities[a-1][b-1]
	ba := o.selectivities[b-1][a-1]
	if ab != ba {
		panic("selectivity should be commutative")
	}
	return ab != -1
}

func (o *Orderer) AddPredicate(a, b RelationID, sel Selectivity) {
	if o.selectivities[a-1][b-1] == -1 {
		o.selectivities[a-1][b-1] = 1
		o.selectivities[b-1][a-1] = 1
	}
	o.selectivities[a-1][b-1] *= sel
	o.selectivities[b-1][a-1] *= sel
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
