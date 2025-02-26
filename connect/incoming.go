package connect

import (
	"github.com/panjf2000/gnet/v2"
	"sync"
)

var incomingConnections sync.Map //[string, io.ReadWriteCloser]

func GetIncoming(id string) gnet.Conn {
	val, ok := incomingConnections.Load(id)
	if ok {
		return val.(gnet.Conn)
	}
	return nil
}
