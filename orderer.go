package main

import "github.com/justinj/joinorder/schema"

type Sequence []schema.RelationID

type Orderer struct {
	s *schema.Schema
}

func NewOrderer(s *schema.Schema) *Orderer {
	return &Orderer{s: s}
}

func (o *Orderer) BruteForceOrder() Sequence {
	var best Sequence
	var bestCost float64

	start := make(Sequence, o.s.NumRels())
	for i := 0; i < o.s.NumRels(); i++ {
		start[i] = schema.RelationID(i + 1)
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
	cost := float64(o.s.Cardinality(ord[0]))
	numRows := float64(o.s.Cardinality(ord[0]))

	for i := 1; i < len(ord); i++ {
		// Calculate selectivity of this relation with all
		// previous relations.
		for j := 0; j < i; j++ {
			numRows *= float64(o.s.Selectivity(ord[i], ord[j]))
		}
		numRows *= float64(o.s.Cardinality(ord[i]))
		cost += numRows
	}
	return cost
}
