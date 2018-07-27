package main

import (
	"testing"

	"github.com/justinj/joinorder/join"
	"github.com/justinj/joinorder/schema"
)

func (o Sequence) equal(other Sequence) bool {
	if len(o) != len(other) {
		panic("incompatible sequence")
	}
	for i := range o {
		if o[i] != other[i] {
			return false
		}
	}
	return true
}

func TestCosting(t *testing.T) {
	builder := schema.NewBuilder()

	a := builder.AddRelation("A", 50)
	b := builder.AddRelation("B", 1000)
	c := builder.AddRelation("C", 50000)

	builder.AddPredicate(a, b, 0.01)
	builder.AddPredicate(b, c, 0.1)

	s := builder.Build()

	o := NewOrderer(s)

	cases := []struct {
		ordering Sequence
		expected float64
	}{
		{Sequence{1, 2, 3}, 2500550},
		{Sequence{2, 1, 3}, 2501500},
		{Sequence{1, 3, 2}, 5000050},
		{Sequence{3, 1, 2}, 5050000},
		{Sequence{2, 3, 1}, 7501000},
		{Sequence{3, 2, 1}, 7550000},
	}

	for _, tc := range cases {
		cost := o.Cost(tc.ordering)
		if cost != tc.expected {
			t.Errorf("expected cost of %v was %.0f instead of %v", tc.ordering, cost, tc.expected)
		}
	}
}

//  1 - 2 - 4
//      |
//      3 - 5
//          |
//          6

func TestBiggerJoin(t *testing.T) {
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

	o := NewOrderer(builder.Build())

	expected := Sequence{2, 4, 1, 3, 5, 6}
	actual := o.BruteForceOrder()
	if !expected.equal(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}

	if o.Cost(actual) != 220058 {
		t.Fatalf("expected %v, got %v", 220058, o.Cost(actual))
	}
}

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

func TestDPSizeOrderer(t *testing.T) {
	builder := schema.NewBuilder()

	a := builder.AddRelation("A", 1000)
	b := builder.AddRelation("B", 1000)
	c := builder.AddRelation("C", 1000)
	d := builder.AddRelation("D", 1000)

	builder.AddPredicate(a, b, 0.0000001)
	builder.AddPredicate(a, c, 0.5)
	builder.AddPredicate(c, d, 0.0000001)

	s := builder.Build()

	o := NewDPSizeOrderer(s)

	o.Order()
}

func TestIKKBZOrderer(t *testing.T) {
	s := makeTestSchema()
	o := NewIKKBZOrderer(s)

	o.SetRoot(3)

	//       3
	//     /   \
	//    2     5
	//  /   \   |
	// 1     4  6

	if o.RootedSelectivity(3) != 1 {
		t.Fatal("rooted selectivity of root is 1")
	}

	if o.RootedSelectivity(2) != 0.1 {
		t.Fatal("rooted selectivity of 2 is its selectivity with 3, which is 0.1")
	}

	if o.RootedSelectivity(6) != 0.0001 {
		t.Fatal("rooted selectivity of 6 is its selectivity with 5, which is 0.0001")
	}

	// Test T function.

	cases := []struct {
		s   Sequence
		exp float64
	}{
		{Sequence{}, 1},
		{Sequence{1}, 1},
		{Sequence{2}, 1},
		{Sequence{3}, 100},
		{Sequence{3, 5}, 50000},
	}

	for _, tc := range cases {
		actual := o.T(tc.s)
		if actual != tc.exp {
			t.Errorf("expected T(%v) to be %v, not %v", tc.s, tc.exp, actual)
		}
	}

	// Test C function.

	cases = []struct {
		s   Sequence
		exp float64
	}{
		{Sequence{}, 0},
		{Sequence{1}, 1},
		{Sequence{2}, 1},
		{Sequence{3}, 100},
		{Sequence{3, 5}, 50100},
		{Sequence{2, 4, 1, 3, 5, 6}, 220041.8},
		{Sequence{2, 1, 4, 3, 5, 6}, 220042.4},
	}

	for _, tc := range cases {
		actual := o.C(tc.s)
		if actual != tc.exp {
			t.Errorf("expected C(%v) to be %v, not %v", tc.s, tc.exp, actual)
		}
	}

	// Test R function.

	cases = []struct {
		s   Sequence
		exp float64
	}{
		{Sequence{1}, 0},
		{Sequence{2}, 0},
		{Sequence{3}, 0.99},
		{Sequence{4}, -1.4999999999999998},
		{Sequence{5}, 0.998},
		{Sequence{6}, 0.9},
		{Sequence{3, 5}, 0.9979840319361277},
		{Sequence{2, 4, 1, 3, 5, 6}, 0.9089136700390562},
		{Sequence{2, 1, 4, 3, 5, 6}, 0.9089111916612435},
	}

	for _, tc := range cases {
		actual := o.R(tc.s)
		if actual != tc.exp {
			t.Errorf("expected R(%v) to be %v, not %v", tc.s, tc.exp, actual)
		}
	}

	expected := "(((((B ⋈ D) ⋈ A) ⋈ C) ⋈ E) ⋈ F)"

	j := join.NewForest(s)
	g := o.Order(j)
	actual := j.FormatString(g)

	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
