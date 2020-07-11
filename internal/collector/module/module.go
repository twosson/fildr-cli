package module

type Module interface {
	Name() string
	Start() error
	Stop()
}
