package join

type Orderer interface {
	Order(*Forest) GroupID
}
