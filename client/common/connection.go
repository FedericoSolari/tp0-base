package common

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type ConnectionHandler struct {
	conn     net.Conn
	leftover []byte
	buffer   []byte
}

func NewConnectionHandler(conn net.Conn) *ConnectionHandler {
	return &ConnectionHandler{
		conn:     conn,
		leftover: nil,
		buffer:   make([]byte, 0, 1024),
	}
}

func (c *ConnectionHandler) SendAll(data []byte) error {
	if c.conn == nil {
		return fmt.Errorf("no hay una conexion abierta")
	}

	total := len(data)
	sent := 0

	for sent < total {
		n, err := c.conn.Write(data[sent:])
		if err != nil {
			return err
		}
		sent += n
	}
	return nil
}

// busca un '\n' en el buffer y devuleve hasta el \n y lo que esta despues
func extractLine(buffer []byte) (line string, rest []byte, found bool) {
	if idx := bytes.IndexByte(buffer, '\n'); idx != -1 {
		return string(buffer[:idx+1]), buffer[idx+1:], true
	}
	return "", buffer, false
}

func (c *ConnectionHandler) RecvAll() (string, error) {
	tmp := make([]byte, 256)

	for {
		if line, rest, found := extractLine(c.buffer); found {
			c.buffer = rest
			return line, nil
		}

		bytesRead, err := c.conn.Read(tmp)
		if err != nil {
			if err == io.EOF {
				if len(c.buffer) > 0 {
					line := string(c.buffer)
					c.buffer = nil
					return line, nil
				}
				return "", io.EOF
			}
			return "", fmt.Errorf("error leyendo del socket: %w", err)
		}

		c.buffer = append(c.buffer, tmp[:bytesRead]...)
	}
}
