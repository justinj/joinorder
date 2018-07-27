package main

import (
	"testing"

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

// //  1 - 2 - 4
// //      |
// //      3 - 5
// //          |
// //          6

// func TestBiggerJoin(t *testing.T) {
// 	o := NewOrderer(6)

// 	o.SetCardinality(1, 100)
// 	o.SetCardinality(2, 10)
// 	o.SetCardinality(3, 100)
// 	o.SetCardinality(4, 1000)
// 	o.SetCardinality(5, 10000)
// 	o.SetCardinality(6, 100000)

// 	o.AddPredicate(1, 2, 0.01)
// 	o.AddPredicate(2, 4, 0.0004)
// 	o.AddPredicate(2, 3, 0.1)
// 	o.AddPredicate(3, 5, 0.05)
// 	o.AddPredicate(5, 6, 0.0001)

// 	expected := Sequence{2, 4, 1, 3, 5, 6}
// 	actual := o.BruteForceOrder()
// 	if !expected.equal(actual) {
// 		t.Fatalf("expected %v, got %v", expected, actual)
// 	}

// 	if o.Cost(actual) != 220058 {
// 		t.Fatalf("expected %v, got %v", 220058, o.Cost(actual))
// 	}
// }

// func TestIKKBZOrderer(t *testing.T) {
// 	o := NewIKKBZOrderer(6)

// 	o.SetCardinality(1, 100)
// 	o.SetCardinality(2, 10)
// 	o.SetCardinality(3, 100)
// 	o.SetCardinality(4, 1000)
// 	o.SetCardinality(5, 10000)
// 	o.SetCardinality(6, 100000)

// 	o.AddPredicate(1, 2, 0.01)
// 	o.AddPredicate(2, 4, 0.0004)
// 	o.AddPredicate(2, 3, 0.1)
// 	o.AddPredicate(3, 5, 0.05)
// 	o.AddPredicate(5, 6, 0.0001)

// 	o.SetRoot(3)

// 	//       3
// 	//     /   \
// 	//    2     5
// 	//  /   \   |
// 	// 1     4  6

// 	if o.RootedSelectivity(3) != 1 {
// 		t.Fatal("rooted selectivity of root is 1")
// 	}

// 	if o.RootedSelectivity(2) != 0.1 {
// 		t.Fatal("rooted selectivity of 2 is its selectivity with 3, which is 0.1")
// 	}

// 	if o.RootedSelectivity(6) != 0.0001 {
// 		t.Fatal("rooted selectivity of 6 is its selectivity with 5, which is 0.0001")
// 	}

// 	// Test T function.

// 	cases := []struct {
// 		s   Sequence
// 		exp float64
// 	}{
// 		{Sequence{}, 1},
// 		{Sequence{1}, 1},
// 		{Sequence{2}, 1},
// 		{Sequence{3}, 100},
// 		{Sequence{3, 5}, 50000},
// 	}

// 	for _, tc := range cases {
// 		actual := o.T(tc.s)
// 		if actual != tc.exp {
// 			t.Errorf("expected T(%v) to be %v, not %v", tc.s, tc.exp, actual)
// 		}
// 	}

// 	// Test C function.

// 	cases = []struct {
// 		s   Sequence
// 		exp float64
// 	}{
// 		{Sequence{}, 0},
// 		{Sequence{1}, 1},
// 		{Sequence{2}, 1},
// 		{Sequence{3}, 100},
// 		{Sequence{3, 5}, 50100},
// 		{Sequence{2, 4, 1, 3, 5, 6}, 220041.8},
// 		{Sequence{2, 1, 4, 3, 5, 6}, 220042.4},
// 	}

// 	for _, tc := range cases {
// 		actual := o.C(tc.s)
// 		if actual != tc.exp {
// 			t.Errorf("expected C(%v) to be %v, not %v", tc.s, tc.exp, actual)
// 		}
// 	}

// 	// Test R function.

// 	cases = []struct {
// 		s   Sequence
// 		exp float64
// 	}{
// 		{Sequence{1}, 0},
// 		{Sequence{2}, 0},
// 		{Sequence{3}, 0.99},
// 		{Sequence{4}, -1.4999999999999998},
// 		{Sequence{5}, 0.998},
// 		{Sequence{6}, 0.9},
// 		{Sequence{3, 5}, 0.9979840319361277},
// 		{Sequence{2, 4, 1, 3, 5, 6}, 0.9089136700390562},
// 		{Sequence{2, 1, 4, 3, 5, 6}, 0.9089111916612435},
// 	}

// 	for _, tc := range cases {
// 		actual := o.R(tc.s)
// 		if actual != tc.exp {
// 			t.Errorf("expected R(%v) to be %v, not %v", tc.s, tc.exp, actual)
// 		}
// 	}

// 	expected := Sequence{2, 4, 1, 3, 5, 6}
// 	actual := o.Order()

// 	if !expected.equal(actual) {
// 		t.Fatalf("expected %v, got %v", expected, actual)
// 	}
// }

// func TestKBZExample(t *testing.T) {
// 	o := NewIKKBZOrderer(5)

// 	o.SetCardinality(1, 100)
// 	o.SetCardinality(2, 1000000)
// 	o.SetCardinality(3, 1000)
// 	o.SetCardinality(4, 15000)
// 	o.SetCardinality(5, 50)

// 	o.AddPredicate(1, 2, 0.01)
// 	o.AddPredicate(1, 3, 1)
// 	o.AddPredicate(3, 4, 0.0333)
// 	o.AddPredicate(3, 5, 0.1)

// 	expected := Sequence{5, 3, 1, 4, 2}
// 	actual := o.Order()

// 	if !expected.equal(actual) {
// 		t.Fatalf("expected %v, got %v", expected, actual)
// 	}
// }

// func Test100WayJoin(t *testing.T) {
// 	rand.Seed(100)

// 	size := 100
// 	o := NewIKKBZOrderer(size)
// 	for i := 1; i <= size; i++ {
// 		o.SetCardinality(RelationID(i), Cardinality(rand.Intn(1000000)))
// 	}

// 	for i := 2; i <= size; i++ {
// 		other := rand.Intn(i-1) + 1
// 		selectivity := rand.Float64()
// 		o.AddPredicate(RelationID(i), RelationID(other), Selectivity(selectivity))
// 	}

// 	expected := `[1 2 31 57 56 68 60 72 38 13 48 77 29 88 90 35 76 69 3 7 17 51 81 24 8 9 19 27 30 36 95 46 47 5 96 33 82 64 73 53 66 97 61 75 26 45 44 52 20 21 28 63 70 39 84 42 85 67 34 93 100 10 11 32 40 65 89 18 94 14 15 78 25 98 55 86 83 43 79 91 59 23 62 4 6 71 16 12 22 49 50 80 37 41 54 87 58 92 99 74]`

// 	if fmt.Sprintf("%v", o.Order()) != expected {
// 		t.Fatalf("big join was wrong:\n%v\n%v", o.Order(), expected)
// 	}
// }

// func Benchmark100WayJoin(b *testing.B) {
// 	rand.Seed(100)

// 	size := 100
// 	o := NewIKKBZOrderer(size)
// 	for i := 1; i <= size; i++ {
// 		o.SetCardinality(RelationID(i), Cardinality(rand.Intn(1000000)))
// 	}

// 	for i := 2; i <= size; i++ {
// 		other := rand.Intn(i-1) + 1
// 		selectivity := rand.Float64()
// 		o.AddPredicate(RelationID(i), RelationID(other), Selectivity(selectivity))
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		o.Order()
// 	}
// }
