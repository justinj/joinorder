package schema

import "fmt"

type RelationName string
type RelationID int
type Selectivity float64
type Cardinality float64

type Relation struct {
	name RelationName
	id   RelationID
	card Cardinality
}

func pair(x, y RelationID) int {
	a, b := int(x)-1, int(y)-1
	if a > b {
		a, b = b, a
	}
	return b*(b-1)/2 + a
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

func (b *Builder) AddRelation(name RelationName, card Cardinality) RelationID {
	if _, ok := b.nameToIdx[name]; ok {
		panic(fmt.Sprintf("duplicate relation name %s", name))
	}

	b.nameToIdx[name] = len(b.relations)

	for i := 0; i < len(b.relations); i++ {
		b.selectivities = append(b.selectivities, -1)
	}

	id := RelationID(len(b.relations) + 1)

	b.relations = append(b.relations, Relation{
		name: name,
		card: card,
		id:   id,
	})

	return id
}

func (b *Builder) relation(x RelationID) Relation {
	if int(x)-1 >= len(b.relations) || x < 1 {
		panic("invalid RelationID")
	}
	return b.relations[x-1]
}

func (b *Builder) AddPredicate(x, y RelationID, sel Selectivity) {
	b.selectivities[pair(x, y)] = sel
}

func (b *Builder) Build() *Schema {
	return &Schema{
		relations:     b.relations,
		selectivities: b.selectivities,
	}
}

type Schema struct {
	relations     []Relation
	selectivities []Selectivity
}

func (s *Schema) relation(x RelationID) Relation {
	if int(x)-1 >= len(s.relations) || x < 1 {
		panic("invalid RelationID")
	}
	return s.relations[x-1]
}

func (s *Schema) Adjacent(a, b RelationID) bool {
	return s.selectivities[pair(a, b)] != -1
}

func (s *Schema) NumRels() int {
	return len(s.relations)
}

func (s *Schema) Selectivity(a, b RelationID) Selectivity {
	sel := s.selectivities[pair(a, b)]
	if sel == -1 {
		return 1
	}
	return sel
}

func (s *Schema) Cardinality(a RelationID) Cardinality {
	return s.relation(a).card
}
