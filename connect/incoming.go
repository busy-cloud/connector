package connect

import (
	"github.com/busy-cloud/connector/types"
	"io"
	"sync"
)

var incomingConnections sync.Map //[string, io.ReadWriteCloser]

type Incoming struct {
	*types.Incoming

	conn io.ReadWriteCloser
}

func (l *Incoming) Close() error {
	return l.conn.Close()
}

func GetIncoming(id string) *Incoming {
	val, ok := incomingConnections.Load(id)
	if ok {
		return val.(*Incoming)
	}
	return nil
}
