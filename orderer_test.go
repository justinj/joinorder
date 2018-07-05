package main

import (
	"testing"
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
	o := NewOrderer(3)

	o.SetCardinality(1, 50)
	o.SetCardinality(2, 1000)
	o.SetCardinality(3, 50000)

	// Joining between 2 and 3 is expensive.
	o.AddPredicate(2, 3, 0.1)

	// Joining between 1 and 2 is very selective.
	o.AddPredicate(1, 2, 0.01)

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
	o := NewOrderer(6)

	o.SetCardinality(1, 100)
	o.SetCardinality(2, 10)
	o.SetCardinality(3, 100)
	o.SetCardinality(4, 1000)
	o.SetCardinality(5, 10000)
	o.SetCardinality(6, 100000)

	o.AddPredicate(1, 2, 0.01)
	o.AddPredicate(2, 4, 0.0004)
	o.AddPredicate(2, 3, 0.1)
	o.AddPredicate(3, 5, 0.05)
	o.AddPredicate(5, 6, 0.0001)

	expected := Sequence{2, 4, 1, 3, 5, 6}
	actual := o.BruteForceOrder()
	if !expected.equal(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}

	if o.Cost(actual) != 220058 {
		t.Fatalf("expected %v, got %v", 220058, o.Cost(actual))
	}
}

func TestIKKBZOrderer(t *testing.T) {
	o := NewIKKBZOrderer(6)

	o.SetCardinality(1, 100)
	o.SetCardinality(2, 10)
	o.SetCardinality(3, 100)
	o.SetCardinality(4, 1000)
	o.SetCardinality(5, 10000)
	o.SetCardinality(6, 100000)

	o.AddPredicate(1, 2, 0.01)
	o.AddPredicate(2, 4, 0.0004)
	o.AddPredicate(2, 3, 0.1)
	o.AddPredicate(3, 5, 0.05)
	o.AddPredicate(5, 6, 0.0001)

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

	expected := Sequence{2, 4, 1, 3, 5, 6}
	actual := o.Order()

	if !expected.equal(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func TestKBZExample(t *testing.T) {
	o := NewIKKBZOrderer(5)

	o.SetCardinality(1, 100)
	o.SetCardinality(2, 1000000)
	o.SetCardinality(3, 1000)
	o.SetCardinality(4, 15000)
	o.SetCardinality(5, 50)

	o.AddPredicate(1, 2, 0.01)
	o.AddPredicate(1, 3, 1)
	o.AddPredicate(3, 4, 0.0333)
	o.AddPredicate(3, 5, 0.1)

	expected := Sequence{3, 5, 1, 4, 2}
	actual := o.Order()

	if !expected.equal(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
