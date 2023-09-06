package components

type Registry interface {
	Register(ID string, name string, address string, port int64) error
}
