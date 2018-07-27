package join

import (
	"bytes"
	"fmt"

	"github.com/justinj/joinorder/schema"
	"github.com/justinj/joinorder/util"
)

type GroupID int

// Forest is a memo-like structure describing a forest of possible join
// trees. It only has one expr per group as of now.
type Forest struct {
	s      *schema.Schema
	groups []group
}

type group struct {
	j  *Forest
	id GroupID

	// relID is 0 if this is not a leaf group.
	relID schema.RelationID

	// l and r are 0 if this is a leaf group.
	l GroupID
	r GroupID

	relations schema.RelSet
}

func NewForest(s *schema.Schema) *Forest {
	return &Forest{
		s:      s,
		groups: make([]group, 1),
	}
}

func (j *Forest) AddLeaf(r schema.RelationID) GroupID {
	id := GroupID(len(j.groups))
	j.groups = append(j.groups, group{
		j:         j,
		id:        id,
		relID:     r,
		relations: util.MakeFastIntSet(int(r)),
	})

	return id
}

func (j *Forest) AddJoin(l, r GroupID) GroupID {
	id := GroupID(len(j.groups))
	j.groups = append(j.groups, group{
		j:         j,
		id:        id,
		l:         l,
		r:         r,
		relations: j.groups[l].relations.Union(j.groups[r].relations),
	})
	return id
}

func (j *Forest) GetMembers(g GroupID) schema.RelSet {
	return j.groups[g].relations
}

func (j *Forest) FormatString(g GroupID) string {
	var buf bytes.Buffer
	j.format(g, &buf)
	return buf.String()
}

func (j *Forest) String() string {
	var buf bytes.Buffer

	for i, g := range j.groups {
		if i == 0 {
			continue
		}
		fmt.Fprintf(&buf, "G%d - ", i)
		if g.relID != 0 {
			fmt.Fprintf(&buf, "[%s]", j.s.Relation(g.relID).Name)
		} else {
			fmt.Fprintf(&buf, "G%d ⋈ G%d", g.l, g.r)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

func (j *Forest) format(g GroupID, buf *bytes.Buffer) {
	if g == 0 {
		panic("zero group")
	}
	group := j.groups[g]
	if group.relID != 0 {
		buf.WriteString(string(j.s.Relation(group.relID).Name))
	} else {
		buf.WriteByte('(')
		j.format(group.l, buf)
		buf.WriteString(" ⋈ ")
		j.format(group.r, buf)
		buf.WriteByte(')')
	}
}
