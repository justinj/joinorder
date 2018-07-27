package schema

import "testing"

func TestSchema(t *testing.T) {
	builder := NewBuilder()

	a := builder.AddRelation("A", 100)
	b := builder.AddRelation("B", 1000)
	c := builder.AddRelation("C", 3)

	builder.AddPredicate(a, b, 0.2)
	builder.AddPredicate(b, c, 0.01)

	s := builder.Build()

	if s.Cardinality(a) != 100 {
		t.Fatal("cardinality of a should be 100")
	}

	if s.Cardinality(b) != 1000 {
		t.Fatal("cardinality of b should be 1000")
	}

	if s.Cardinality(c) != 3 {
		t.Fatal("cardinality of c should be 3")
	}

	if !s.Adjacent(a, b) {
		t.Fatal("a and b should be adjacent")
	}

	if !s.Adjacent(b, c) {
		t.Fatal("b and c should be adjacent")
	}

	if s.Adjacent(a, c) {
		t.Fatal("a and c should not be adjacent")
	}

	if s.Selectivity(a, b) != 0.2 {
		t.Fatal("selectivity between a and b is wrong")
	}

	if s.Selectivity(b, c) != 0.01 {
		t.Fatal("selectivity between b and c is wrong")
	}

	if s.Selectivity(a, c) != 1 {
		t.Fatal("selectivity between a and c is wrong")
	}
}
