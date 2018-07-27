package main

import "testing"

func TestSchema(t *testing.T) {
	b := NewBuilder()

	b.AddRelation("A", 100)
	b.AddRelation("B", 1000)
	b.AddRelation("C", 3)

	s := b.Build()

	_ = s
}
