package queries

import (
	"fmt"

	"github.com/justinj/joinorder/schema"
)

func QueryByName(n string) *schema.Schema {
	switch n {
	case "bushy":
		return Bushy()
	}
	panic(fmt.Sprintf("unknown query %q", n))
}
