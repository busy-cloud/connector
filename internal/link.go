package internal

import "io"

type Link interface {
	io.ReadWriteCloser
	Open() error
	Opened() bool
	Connected() bool
}
