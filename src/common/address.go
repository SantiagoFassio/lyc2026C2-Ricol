package common

import (
	"net"
	"net/netip"
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
	if ip, err := netip.ParseAddr(host); err == nil {
		host = ip.String()
	}
	return Address{Host: host, Port: int(port)}, true
}

func (a Address) String() string {
	return net.JoinHostPort(a.Host, strconv.Itoa(a.Port))
}
