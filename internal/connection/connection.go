package connection

import (
	"PessiTorrent/internal/packets"
	"PessiTorrent/internal/serialization"
	"encoding/binary"
	"net"
)

type Connection struct {
	conn net.Conn
}

func NewConnection(conn net.Conn) Connection {
	return Connection{
		conn: conn,
	}
}

func (c *Connection) WritePacket(packet interface{}) error {
	data, err := serialization.Serialize(packet)
	if err != nil {
		return err
	}

	// Write packet length
	err = binary.Write(c.conn, binary.LittleEndian, uint32(len(data)))
	if err != nil {
		return err
	}

	// Write packet data
	_, err = c.conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) ReadPacket() (interface{}, uint8, error) {
	// Read packet length
	var packetLength uint32
	err := binary.Read(c.conn, binary.LittleEndian, &packetLength)
	if err != nil {
		return nil, 0, err
	}

	// Read packet data
	data := make([]byte, packetLength)
	_, err = c.conn.Read(data)
	if err != nil {
		return nil, 0, err
	}

	// First byte is the packet type
	packetType := data[0]
	packetStruct := packets.PacketStructFromType(packetType)
	err = serialization.Deserialize(data, packetStruct)
	if err != nil {
		return nil, 0, err
	}

	return packetStruct, packetType, nil
}

func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
