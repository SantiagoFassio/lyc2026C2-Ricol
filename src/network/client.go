package network

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

type Client struct {
	conn             net.Conn
	serviceInterface *wire.Interface
	address          common.Address
	timeout          time.Duration
}

func Dial(address common.Address, serviceInterface *wire.Interface, timeout time.Duration) (*Client, *Failure) {
	client := &Client{serviceInterface: serviceInterface, address: address, timeout: timeout}
	deadline := time.Now().Add(timeout)
	conn, err := (&net.Dialer{Deadline: deadline}).Dial("tcp", address.String())
	if err != nil {
		return nil, client.openFailure(ConnectionFailed, client.describe(err))
	}
	client.conn = conn
	if failure := client.handshake(deadline); failure != nil {
		conn.Close()
		return nil, failure
	}
	return client, nil
}

func (c *Client) handshake(deadline time.Time) *Failure {
	c.conn.SetDeadline(deadline)
	if _, err := c.conn.Write(encodeHello(c.serviceInterface.Hash)); err != nil {
		return c.openFailure(ConnectionFailed, c.describe(err))
	}
	if err := readHeader(c.conn); err != nil {
		switch {
		case errors.Is(err, errNotRicol):
			return c.openFailure(ProtocolError, "the server is not a Ricol program")
		case errors.Is(err, errUnsupportedVersion):
			return c.openFailure(ProtocolError, err.Error())
		}
		return c.openFailure(ConnectionFailed, c.describe(err))
	}
	payload, err := readFrame(c.conn)
	if err != nil {
		return c.openFailure(ConnectionFailed, c.describe(err))
	}
	if failure := decodeWelcome(payload); failure != nil {
		return c.openFailure(failure.Kind, failure.Message)
	}
	return nil
}

func (c *Client) Call(method *wire.Method, arguments []types.Value) (types.Value, *Failure, error) {
	request, err := encodeRequest(method, arguments)
	if err != nil {
		return nil, nil, err
	}
	c.conn.SetDeadline(time.Now().Add(c.timeout))
	if _, err := c.conn.Write(request); err != nil {
		return nil, c.callFailure(method, err), nil
	}
	payload, err := readFrame(c.conn)
	if err != nil {
		return nil, c.callFailure(method, err), nil
	}
	result, failure := decodeResult(payload, method)
	return result, failure, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) openFailure(kind Kind, cause string) *Failure {
	return &Failure{Kind: kind, Message: fmt.Sprintf("cannot connect to %s at %v: %s", c.serviceInterface.Name, c.address, cause)}
}

func (c *Client) callFailure(method *wire.Method, err error) *Failure {
	if isTimeout(err) {
		return &Failure{Kind: Timeout, Message: fmt.Sprintf("'%s' did not answer within %v", method.Name, c.timeout)}
	}
	return &Failure{Kind: ConnectionLost, Message: fmt.Sprintf("connection to %s at %v lost during '%s': %s",
		c.serviceInterface.Name, c.address, method.Name, c.describe(err))}
}

func (c *Client) describe(err error) string {
	switch {
	case isTimeout(err):
		return fmt.Sprintf("timed out after %v", c.timeout)
	case errors.Is(err, io.EOF):
		return "the connection was closed"
	case errors.Is(err, io.ErrUnexpectedEOF):
		return errTruncatedFrame.Error()
	}
	return rootCause(err).Error()
}

func isTimeout(err error) bool {
	var netError net.Error
	return errors.As(err, &netError) && netError.Timeout()
}
