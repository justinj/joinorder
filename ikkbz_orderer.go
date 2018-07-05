package main

import "fmt"

type IKKBZOrderer struct {
	Orderer
	root    RelationID
	parents []RelationID
}

func NewIKKBZOrderer(numRels int) *IKKBZOrderer {
	return &IKKBZOrderer{
		Orderer: *NewOrderer(numRels),
		parents: make([]RelationID, numRels+1),
	}
}

func (o *IKKBZOrderer) SetRoot(r RelationID) {
	o.root = r

	for i := range o.parents {
		o.parents[i] = 0
	}
	curNode := r
	for {
		found := false
		for i := 1; i <= o.numRelations; i++ {
			if i == int(curNode) || o.parents[i] == curNode || o.parents[curNode] == RelationID(i) {
				continue
			}

			if o.Adjacent(RelationID(i), curNode) {
				found = true
				if o.parents[i] != 0 {
					panic(fmt.Sprintf("query graph was not a tree: %v, i = %d, curNode = %d", o.parents, i, curNode))
				}

				o.parents[i] = curNode
				curNode = RelationID(i)
				break
			}
		}

		if !found {
			curNode = o.parents[curNode]
		}

		if curNode == 0 {
			break
		}
	}
}

func (o *IKKBZOrderer) RootedSelectivity(r RelationID) Selectivity {
	if o.root == 0 {
		panic("root not set")
	}
	if o.root == r {
		return 1
	}
	return o.GetSelectivity(r, o.parents[r])
}

// T is the T function from Ibaraki and Kameda, representing a factor which the
// given set of relations contribute to the final row count.
func (o *IKKBZOrderer) T(s Sequence) float64 {
	p := float64(1)
	for _, r := range s {
		p *= float64(o.RootedSelectivity(r)) * float64(o.Cardinality(r))
	}
	return p
}

// C is the cost function from Ibaraki and Kameda.
//
//   C(S_1S_2) = C(S_1) + T(S_1)C(S_2)
func (o *IKKBZOrderer) C(s Sequence) float64 {
	cost := float64(0)
	factor := float64(1)
	for _, r := range s {
		contribution := float64(o.RootedSelectivity(r)) * float64(o.Cardinality(r))
		cost += factor * contribution
		factor *= contribution
	}
	return cost
}

// R is the rank function from Ibaraki and Kameda.
func (o *IKKBZOrderer) R(s Sequence) float64 {
	if len(s) == 0 {
		panic("rank of empty sequence not defined")
	}
	return (o.T(s) - 1) / o.C(s)
}

func (o *IKKBZOrderer) ChildrenOf(r RelationID) []RelationID {
	result := make([]RelationID, 0)
	for i := 1; i < len(o.parents); i++ {
		if o.parents[i] == r {
			result = append(result, RelationID(i))
		}
	}
	return result
}

// Order implementes the Ibaraki/Kameda algorithm for finding the optimal
// left-deep join order.
// TODO: this should be extended to full IKKBZ.
func (o *IKKBZOrderer) Order() Sequence {
	bestCost := float64(0)
	var bestResult Sequence
	for i := 1; i < o.numRelations; i++ {
		root := RelationID(i)
		o.SetRoot(root)
		result := o.solveWedge(root)
		flattened := make(Sequence, 0)
		for i := range result {
			flattened = append(flattened, result[i]...)
		}
		cost := o.C(flattened)
		if bestCost == 0 || cost < bestCost {
			bestCost = cost
			bestResult = flattened
		}
	}
	return bestResult
}

func (o *IKKBZOrderer) solveWedge(r RelationID) []Sequence {
	children := o.ChildrenOf(r)
	chains := make([][]Sequence, len(children))
	for i := range children {
		chains[i] = o.solveWedge(children[i])
	}

	// Now merge those chains.
	result := []Sequence{Sequence{r}}

	// TODO: more efficient merge here
	for {
		lowestIdx := -1
		lowestRank := float64(0)
		for i := range chains {
			if len(chains[i]) != 0 {
				rank := o.R(chains[i][0])
				if lowestIdx == -1 || rank < lowestRank {
					lowestIdx = i
					lowestRank = rank
				}
			}
		}

		if lowestIdx == -1 {
			break
		}
		result = append(result, chains[lowestIdx][0])
		chains[lowestIdx] = chains[lowestIdx][1:]
	}

	// Now compress decreasing sequences.
	compressed := []Sequence{result[0]}
	for i := 1; i < len(result); i++ {
		prevRank := o.R(compressed[len(compressed)-1])
		newRank := o.R(result[i])
		if newRank < prevRank {
			compressed[len(compressed)-1] = append(compressed[len(compressed)-1], result[i]...)
		} else {
			compressed = append(compressed, result[i])
		}
	}

	return result
}
