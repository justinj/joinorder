package join

import (
	"testing"

	"github.com/justinj/joinorder/schema"
)

func makeTestSchema() *schema.Schema {
	builder := schema.NewBuilder()

	a := builder.AddRelation("A", 100)
	b := builder.AddRelation("B", 10)
	c := builder.AddRelation("C", 100)
	d := builder.AddRelation("D", 1000)
	e := builder.AddRelation("E", 10000)
	f := builder.AddRelation("F", 100000)

	builder.AddPredicate(a, b, 0.01)
	builder.AddPredicate(b, d, 0.0004)
	builder.AddPredicate(b, c, 0.1)
	builder.AddPredicate(c, e, 0.05)
	builder.AddPredicate(e, f, 0.0001)

	return builder.Build()
}

func TestJoin(t *testing.T) {
	s := makeTestSchema()

	//         1
	//     /       \
	//    2         3
	//  /   \     /   \
	// B     4   C     A
	//     /   \
	//    F     5
	//        /   \
	//       E     D

	j := NewForest(s)

	a := j.AddLeaf(s.GetRelationByName("A"))
	b := j.AddLeaf(s.GetRelationByName("B"))
	c := j.AddLeaf(s.GetRelationByName("C"))
	d := j.AddLeaf(s.GetRelationByName("D"))
	e := j.AddLeaf(s.GetRelationByName("E"))
	f := j.AddLeaf(s.GetRelationByName("F"))

	j5 := j.AddJoin(e, d)
	j4 := j.AddJoin(f, j5)
	j3 := j.AddJoin(c, a)
	j2 := j.AddJoin(b, j4)
	j1 := j.AddJoin(j2, j3)

	root := j1

	if j.FormatString(root) != "((B ⋈ (F ⋈ (E ⋈ D))) ⋈ (C ⋈ A))" {
		t.Fatal("wrong stringified output")
	}
}
