package services

type LeafObject struct {
	ID int64
	Name string
}

func (l LeafObject) remote() {}