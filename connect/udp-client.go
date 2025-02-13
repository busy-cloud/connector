package connect

// UdpClient udp与tcp客户端使用方式一致
type UdpClient TcpClient

func NewUdpClient(l *Connect) *UdpClient {
	c := &UdpClient{Connect: l}
	return c
}
