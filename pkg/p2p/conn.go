package p2p

import (
	"net"
	"time"
)

// Client is a TCP connection wrapper.
type Client struct {
	Conn    net.Conn
	Timeout time.Duration
}

func New(address string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn:    conn,
		Timeout: timeout,
	}, nil
}

func (c *Client) Read(b []byte) (int, error) {
	c.Conn.SetReadDeadline(time.Now().Add(c.Timeout))
	return c.Conn.Read(b)
}

func (c *Client) Write(b []byte) (int, error) {
	c.Conn.SetWriteDeadline(time.Now().Add(c.Timeout))
	return c.Conn.Write(b)
}

func (c *Client) Close() error {
	return c.Conn.Close()
}
