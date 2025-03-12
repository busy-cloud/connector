package internal

type Instance interface {
	Open() error
	Close() error
	Opened() bool
	Connected() bool
}
