package schema

// RelSetMap maps RelSets to integers.
type RelSetMap struct {
	m map[uint64]int
}

func NewRelSetMap() *RelSetMap {
	return &RelSetMap{
		m: make(map[uint64]int),
	}
}

func index(s RelSet) uint64 {
	var idx uint64
	for i, ok := s.Next(0); ok; i, ok = s.Next(i + 1) {
		if i > 63 {
			panic("relset too big")
		}
		idx += 1 << uint64(i-1)
	}
	return idx
}

func (m *RelSetMap) Set(s RelSet, i int) {
	m.m[index(s)] = i
}

func (m *RelSetMap) Get(s RelSet) int {
	return m.m[index(s)]
}
