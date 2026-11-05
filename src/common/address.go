package common

import (
	"net"
	"strconv"
)

type Address struct {
	Host string
	Port int
}

func ParseAddress(rawAddress string) (Address, bool) {
	host, rawPort, err := net.SplitHostPort(rawAddress)
	if err != nil || host == "" {
		return Address{}, false
	}
	port, err := strconv.ParseUint(rawPort, 10, 16)
	if err != nil || port == 0 {
		return Address{}, false
	}
	return Address{Host: host, Port: int(port)}, true
}
