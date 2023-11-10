package connection

import (
	"PessiTorrent/internal/packets"
	"PessiTorrent/internal/serialization"
	"encoding/binary"
	"net"
)

type Connection struct {
	conn net.Conn
	buf  []byte
}

func NewConnection(conn net.Conn) Connection {
	return Connection{
		conn: conn,
		buf:  make([]byte, 1024),
	}
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) WritePacket(packet interface{}) error {
	buf, err := serialization.Serialize(packet)
	if err != nil {
		return err
	}

	// FIXME: Move this to other module
	// Write message size (4 bytes)
	var size = len(buf)
	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, uint32(size))
	_, err = c.conn.Write(sizeBuf)
	if err != nil {
		return err
	}

	// Write message content
	_, err = c.conn.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) ReadPacket() (interface{}, uint8, error) {
	// FIXME: Move this to other module
	// Read message size (4 bytes)
	_, err := c.conn.Read(c.buf[:4])
	if err != nil {
		return nil, 0, err
	}
	size := binary.LittleEndian.Uint32(c.buf[:4])

	// Read message content
	_, err = c.conn.Read(c.buf[:size])
	if err != nil {
		return nil, 0, err
	}

	// Read packet type (first byte of message content)
	packetType := c.buf[0]
	packetStruct := packets.PacketStructFromType(packetType)
	err = serialization.Deserialize(c.buf[:size], packetStruct)
	if err != nil {
		return nil, 0, err
	}

	return packetStruct, packetType, nil
}
