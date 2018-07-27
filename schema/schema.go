package main

import "fmt"

type RelationName string
type Selectivity float64

type Relation struct {
	name RelationName
	card int
}

func pair(a, b int) int {
	if a > b {
		a, b = b, a
	}
	return b*(b-1)/2 + a
}

type Schema struct {
	relations     []Relation
	selectivities []Selectivity
}

type Builder struct {
	relations     []Relation
	selectivities []Selectivity
	nameToIdx     map[RelationName]int
}

func NewBuilder() *Builder {
	return &Builder{
		nameToIdx: make(map[RelationName]int),
	}
}

func (b *Builder) AddRelation(name RelationName, card int) {
	if _, ok := b.nameToIdx[name]; ok {
		panic(fmt.Sprintf("duplicate relation name %s", name))
	}

	b.nameToIdx[name] = len(b.relations)

	for i := 0; i < len(b.relations); i++ {
		b.selectivities = append(b.selectivities, -1)
	}

	b.relations = append(b.relations, Relation{
		name: name,
		card: card,
	})
}

func (b *Builder) AddPredicate(x, y RelationName, sel Selectivity) {
	l, ok := b.nameToIdx[x]
	if !ok {
		panic(fmt.Sprintf("no relation %s", x))
	}

	r, ok := b.nameToIdx[y]
	if !ok {
		panic(fmt.Sprintf("no relation %s", y))
	}

	b.selectivities[pair(l, r)] = sel
}

func (b *Builder) Build() *Schema {
	return &Schema{
		relations:     b.relations,
		selectivities: b.selectivities,
	}
}
