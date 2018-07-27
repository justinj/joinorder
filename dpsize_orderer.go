package main

import (
	"fmt"

	"github.com/justinj/joinorder/join"
	"github.com/justinj/joinorder/schema"
	"github.com/justinj/joinorder/util"
)

type DPSizeOrderer struct {
	s     *schema.Schema
	j     *join.Forest
	costs map[join.GroupID]float64
	cards map[join.GroupID]schema.Cardinality
}

func NewDPSizeOrderer(s *schema.Schema) *DPSizeOrderer {
	return &DPSizeOrderer{
		s:     s,
		j:     join.NewForest(s),
		costs: make(map[join.GroupID]float64),
		cards: make(map[join.GroupID]schema.Cardinality),
	}
}

func (o *DPSizeOrderer) Order() join.GroupID {
	subproblems := [][]join.GroupID{nil, []join.GroupID{0}}

	units := schema.NewRelSetMap()
	for i := 1; i <= o.s.NumRels(); i++ {
		units.Set(util.MakeFastIntSet(i), i)
		l := o.j.AddLeaf(schema.RelationID(i))
		subproblems[1] = append(subproblems[1], l)
		o.costs[l] = 0
		o.cards[l] = o.s.Cardinality(schema.RelationID(i))
	}

	bests := []*schema.RelSetMap{nil, units}
	var finalIdx join.GroupID

	for s := 2; s <= o.s.NumRels(); s++ {
		bests = append(bests, schema.NewRelSetMap())
		subproblems = append(subproblems, []join.GroupID{0})
		for s1 := 1; s1 < s; s1++ {
			s2 := s - s1
			for _, l := range subproblems[s1][1:] {
				for _, r := range subproblems[s2][1:] {
					lMembers := o.j.GetMembers(l)
					rMembers := o.j.GetMembers(r)

					if lMembers.Intersects(rMembers) {
						continue
					}

					if !o.s.SubgraphsAdjacent(lMembers, rMembers) {
						continue
					}

					resultingSet := lMembers.Union(rMembers)

					lCost := o.costs[l]
					rCost := o.costs[r]
					sel := o.s.ComplexSelectivity(lMembers, rMembers)

					newCard := float64(o.cards[l]) * float64(o.cards[r]) * float64(sel)
					newCost := lCost + rCost + newCard

					oldBestIdx := join.GroupID(bests[s].Get(resultingSet))
					if oldBestIdx == 0 || newCost < o.costs[oldBestIdx] {
						new := o.j.AddJoin(l, r)
						if oldBestIdx == 0 {
							subproblems[s] = append(subproblems[s], new)
						}
						finalIdx = new
						o.cards[new] = schema.Cardinality(newCard)
						o.costs[new] = newCost
						bests[s].Set(resultingSet, int(new))
					}
				}
			}
		}
	}

	fmt.Println(o.j.FormatString(finalIdx))

	return 0
}
