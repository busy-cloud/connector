package interfaces

type Linker interface {
	Open() error
	Close() error
	Opened() bool
	Connected() bool
}
