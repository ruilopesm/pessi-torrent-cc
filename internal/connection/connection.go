package connection

import (
	"PessiTorrent/internal/packets"
	"PessiTorrent/internal/serialization"
	"fmt"
	"net"
)

type Connection struct {
	conn net.Conn
	buf  []byte
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
		buf:  make([]byte, 1024),
	}
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) ReadPacket() (interface{}, error) {
	n, err := c.conn.Read(c.buf)
	if err != nil {
		return nil, err
	}

	// TODO: enhance this part
	packetType := c.buf[0]
	packetStruct := packets.PacketStructFromType(packetType)
	if packetStruct == nil {
		fmt.Println("invalid packet type")
		return nil, nil
	}

	err = serialization.Deserialize(c.buf[:n], packetStruct)
	if err != nil {
		return nil, err
	}

	return packetStruct, nil
}

func (c *Connection) WritePacket(packet interface{}) error {
	buf, err := serialization.Serialize(packet)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(buf)
	if err != nil {
		return err
	}

	return nil
}
