package queries

import (
	"github.com/justinj/joinorder/schema"
)

// Bushy is a simple example of a query for which a bushy tree is preferable to
// a left-deep tree.
//
// Its query graph looks like this:
//
// B - C
// |   |
// A   D

func Bushy() *schema.Schema {
	builder := schema.NewBuilder()

	a := builder.AddRelation("A", 1000)
	b := builder.AddRelation("B", 900)
	c := builder.AddRelation("C", 800)
	d := builder.AddRelation("D", 700)

	builder.AddPredicate(a, b, 0.01)
	builder.AddPredicate(b, c, 0.5)
	builder.AddPredicate(c, d, 0.01)

	return builder.Build()
}
