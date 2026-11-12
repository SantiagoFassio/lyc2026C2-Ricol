package types

import (
	"net"
	"strconv"
)

type Connection struct {
	service string
	host    string
	port    int
}

func NewConnection(service string, host string, port int) Connection {
	return Connection{service: service, host: host, port: port}
}

func (c Connection) Service() string {
	return c.service
}

func (c Connection) Host() string {
	return c.host
}

func (c Connection) Port() int {
	return c.port
}

func (c Connection) Equals(other Value) bool {
	otherConnection, ok := other.(Connection)
	return ok && c == otherConnection
}

func (c Connection) TypeName() string {
	return "connection<" + c.service + ">"
}

func (c Connection) String() string {
	return c.TypeName() + " at " + net.JoinHostPort(c.host, strconv.Itoa(c.port))
}

func (c Connection) Display() string {
	return c.String()
}

func (Connection) isValue() {}
