package main

import (
	"bytes"
	"fmt"

	"github.com/justinj/joinorder/join"
	"github.com/justinj/joinorder/schema"
)

type IKKBZOrderer struct {
	s       *schema.Schema
	root    schema.RelationID
	parents []schema.RelationID
}

func NewIKKBZOrderer(s *schema.Schema) *IKKBZOrderer {
	return &IKKBZOrderer{
		s:       s,
		parents: make([]schema.RelationID, s.NumRels()+1),
	}
}

func (o *IKKBZOrderer) SetRoot(r schema.RelationID) {
	o.root = r

	for i := range o.parents {
		o.parents[i] = 0
	}
	curNode := r
	for {
		found := false
		for i := 1; i <= o.s.NumRels(); i++ {
			if i == int(curNode) || o.parents[i] == curNode || o.parents[curNode] == schema.RelationID(i) {
				continue
			}

			if o.s.Adjacent(schema.RelationID(i), curNode) {
				found = true
				if o.parents[i] != 0 {
					panic(fmt.Sprintf("query graph was not a tree: %v, i = %d, curNode = %d", o.parents, i, curNode))
				}

				o.parents[i] = curNode
				curNode = schema.RelationID(i)
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

func (o *IKKBZOrderer) RootedSelectivity(r schema.RelationID) schema.Selectivity {
	if o.root == 0 {
		panic("root not set")
	}
	if o.root == r {
		return 1
	}
	return o.s.Selectivity(r, o.parents[r])
}

func (o *IKKBZOrderer) String() string {
	if o.root == 0 {
		panic("must root before printing")
	}

	var buf bytes.Buffer
	o.format(o.root, &buf, 0)
	return buf.String()
}

func (o *IKKBZOrderer) format(r schema.RelationID, buf *bytes.Buffer, depth int) {
	for i := 0; i < depth; i++ {
		buf.WriteByte(' ')
	}
	fmt.Fprintf(buf, "(%d\n", int(r))
	for i := range o.parents {
		if o.parents[i] == r {
			buf.WriteByte(' ')
			o.format(schema.RelationID(i), buf, depth+1)
		}
	}
	for i := 0; i < depth; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString(")\n")
}

// T is the T function from Ibaraki and Kameda, representing a factor which the
// given set of relations contribute to the final row count.
func (o *IKKBZOrderer) T(s Sequence) float64 {
	p := float64(1)
	for _, r := range s {
		p *= float64(o.RootedSelectivity(r)) * float64(o.s.Cardinality(r))
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
		contribution := float64(o.RootedSelectivity(r)) * float64(o.s.Cardinality(r))
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

func (o *IKKBZOrderer) ChildrenOf(r schema.RelationID) []schema.RelationID {
	result := make([]schema.RelationID, 0)
	for i := 1; i < len(o.parents); i++ {
		if o.parents[i] == r {
			result = append(result, schema.RelationID(i))
		}
	}
	return result
}

// Order implementes the Ibaraki/Kameda algorithm for finding the optimal
// left-deep join order.
// TODO: this should be extended to full IKKBZ.
func (o *IKKBZOrderer) Order() join.Join {
	j := join.NewForest(o.s)
	bestCost := float64(0)
	var bestResult Sequence
	for i := 1; i <= o.s.NumRels(); i++ {
		flattened := o.SolveAtRoot(schema.RelationID(i))
		cost := o.C(flattened)
		if bestCost == 0 || cost < bestCost {
			bestCost = cost
			bestResult = flattened
		}
	}

	l := j.AddLeaf(bestResult[0])
	for i := 1; i < len(bestResult); i++ {
		r := j.AddLeaf(bestResult[i])
		l = j.AddJoin(l, r)
	}

	return j.AsJoin(l)
}

func (o *IKKBZOrderer) SolveAtRoot(r schema.RelationID) Sequence {
	o.SetRoot(r)
	result := o.solveWedge(r)
	flattened := make(Sequence, 0)
	for i := range result {
		flattened = append(flattened, result[i]...)
	}
	return flattened
}

func (o *IKKBZOrderer) solveWedge(r schema.RelationID) []Sequence {
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
