package join

type Join struct {
	forest *Forest
	root   GroupID
}

func (j Join) String() string {
	return j.forest.FormatString(j.root)
}
