package interpreter

import "github.com/SantiagoFassio/lyc2026C2-Ricol/common"

type server struct {
	serviceName string
	address     common.Address
	position    common.Position
	environment *environment
}

func newServer(serviceName string, address common.Address, position common.Position, environment *environment) *server {
	return &server{
		serviceName: serviceName,
		address:     address,
		position:    position,
		environment: environment,
	}
}
