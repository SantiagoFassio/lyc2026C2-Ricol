package ricol_test

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/network"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/ricol"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/wire"
)

const echoService = `service Echo {
    func echo(text: String) -> String;
}
`

func echoInterface() *wire.Interface {
	return wire.NewInterface(types.NewServiceType("Echo", []types.ServiceMethod{
		{Name: "echo", Signature: types.Signature{Params: []types.Type{types.Str}, Return: types.Str}},
	}))
}

func localAddress(addr net.Addr) common.Address {
	return common.Address{Host: "127.0.0.1", Port: addr.(*net.TCPAddr).Port}
}

func startEchoServer(t *testing.T) common.Address {
	t.Helper()

	server, err := network.Listen(common.Address{Host: "127.0.0.1", Port: 0}, echoInterface())
	if err != nil {
		t.Fatalf("network.Listen() unexpected error: %v", err)
	}
	served := make(chan error, 1)
	go func() {
		served <- server.Serve(func(_ *wire.Method, arguments []types.Value) (types.Value, error) {
			return types.NewString("echo: " + arguments[0].Display()), nil
		})
	}()
	t.Cleanup(func() {
		server.Close()
		<-served
	})
	return localAddress(server.Addr())
}

func freeAddress(t *testing.T) common.Address {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() unexpected error: %v", err)
	}
	address := localAddress(listener.Addr())
	listener.Close()
	return address
}

func echoClient(address common.Address) string {
	return echoService + fmt.Sprintf(`
session connect Echo at "%v" as c {
    print c.echo("hola");
} rescue e {
    print e;
}
`, address)
}

func TestClientProgramOverTheNetwork(t *testing.T) {
	address := startEchoServer(t)

	output, err := run(t, sourceFiles{"main.ric": echoClient(address)}, ricol.ModeFull)

	if err != nil {
		t.Fatalf("ricol.Run() unexpected error: %v", err)
	}
	if output != "echo: hola" {
		t.Errorf("ricol.Run() output = %q; want %q", output, "echo: hola")
	}
}

func TestClientProgramWithoutServer(t *testing.T) {
	address := freeAddress(t)

	output, err := run(t, sourceFiles{"main.ric": echoClient(address)}, ricol.ModeFull)

	if err != nil {
		t.Fatalf("ricol.Run() unexpected error: %v", err)
	}
	if expected := fmt.Sprintf(`SessionError.ConnectionFailed("cannot connect to Echo at %v: connection refused")`, address); output != expected {
		t.Errorf("ricol.Run() output = %q; want %q", output, expected)
	}
}

func TestServerProgramThatCannotListen(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() unexpected error: %v", err)
	}
	defer occupied.Close()
	address := localAddress(occupied.Addr())
	program := echoService + fmt.Sprintf(`
server Echo at "%v" {
    func echo(text: String) -> String {
        return text;
    }

    print "body";
}
`, address)
	line := strings.Count(echoService, "\n") + 2

	output, err := run(t, sourceFiles{"main.ric": program}, ricol.ModeFull)

	if expected := fmt.Sprintf("[main.ric, line %d, column 16] Cannot listen on %v: address already in use", line, address); err == nil || err.Error() != expected {
		t.Errorf("ricol.Run() error = %v; want %q", err, expected)
	}
	if output != "body" {
		t.Errorf("ricol.Run() output = %q; want the output of the body %q", output, "body")
	}
}
