package interpreter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type serverCall struct {
	method    string
	arguments []types.Value
}

type fakeListener struct {
	transport *fakeTransport
}

type serveTestCase struct {
	name                string
	transport           fakeTransport
	expectedOutput      string
	expectedErrorOutput string
}

type serveErrorTestCase struct {
	name            string
	transport       fakeTransport
	expectedMessage string
}

type serveResult struct {
	interpreter       *Interpreter
	globalEnvironment *environment
	output            string
	errorOutput       string
	err               error
}

func (f *fakeTransport) listen(service string, address common.Address) (listener, error) {
	fmt.Fprintf(f.output, "[listen %s at %v]", service, address)
	if f.listenError != nil {
		return nil, f.listenError
	}
	return &fakeListener{transport: f}, nil
}

func (l *fakeListener) serve(handler methodHandler) error {
	for _, call := range l.transport.serverCalls {
		values := make([]string, len(call.arguments))
		for index, argument := range call.arguments {
			values[index] = argument.String()
		}
		result, err := handler(call.method, call.arguments)
		switch {
		case err != nil:
			fmt.Fprintf(l.transport.output, "[serve %s(%s) failed: %v]", call.method, strings.Join(values, ", "), err)
		case result == nil:
			fmt.Fprintf(l.transport.output, "[serve %s(%s) -> void]", call.method, strings.Join(values, ", "))
		default:
			fmt.Fprintf(l.transport.output, "[serve %s(%s) -> %v]", call.method, strings.Join(values, ", "), result)
		}
	}
	return l.transport.serveError
}

func (d distanceMap) accountServer() *common.ServerDeclarationStatement {
	declaration := serverStatement("Account", "localhost:3825",
		varDeclaration("balance", types.Int, integerLiteral(0)),
		funcStatement("deposit", []common.Parameter{parameter("amount", types.Int)}, types.Int,
			common.NewExpressionStatement(d.assignment("balance", 1, binary(d.variable("balance", 1), common.PLUS, "+", d.variable("amount", 0)))),
			returnStatement(d.variable("balance", 1)),
		),
		funcStatement("current", nil, types.Int, returnStatement(d.variable("balance", 1))),
		funcStatement("reset", nil, types.Void, common.NewExpressionStatement(d.assignment("balance", 1, integerLiteral(0)))),
		funcStatement("fail", nil, types.Int, returnStatement(divisionByZero())),
		common.NewExpressionStatement(d.call("deposit", 0, integerLiteral(100))),
	)
	declaration.AddressToken.Position = common.Position{Line: 3, Column: 18}
	return declaration
}

func callOf(method string, arguments ...int64) serverCall {
	values := make([]types.Value, len(arguments))
	for index, argument := range arguments {
		values[index] = types.NewInteger(argument)
	}
	return serverCall{method: method, arguments: values}
}

func serveProgram(t *testing.T, transport *fakeTransport, program func(d distanceMap) []common.Statement) serveResult {
	t.Helper()

	distances := distanceMap{}
	output := &bytes.Buffer{}
	errorOutput := &bytes.Buffer{}
	interpreter := NewInterpreter(program(distances), distances, output)
	interpreter.errorOutput = errorOutput
	if transport != nil {
		transport.output = output
		interpreter.transport = transport
	}
	globalEnvironment := interpreter.currentEnvironment
	if err := interpreter.Interpret(); err != nil {
		t.Fatalf("Interpret() unexpected error: %v", err)
	}
	err := interpreter.Serve()
	return serveResult{interpreter, globalEnvironment, output.String(), errorOutput.String(), err}
}

func accountProgram(d distanceMap) []common.Statement {
	return []common.Statement{d.accountServer()}
}

func TestServe(t *testing.T) {
	testCases := []serveTestCase{
		{"calls to the methods", fakeTransport{serverCalls: []serverCall{
			callOf("deposit", 30), callOf("deposit", 12), callOf("current"), callOf("reset"), callOf("current"),
		}}, "[listen Account at localhost:3825][serve deposit(30) -> 130][serve deposit(12) -> 142][serve current() -> 142]" +
			"[serve reset() -> void][serve current() -> 0]", ""},
		{"failure of a method", fakeTransport{serverCalls: []serverCall{callOf("fail"), callOf("deposit", 1)}},
			"[listen Account at localhost:3825][serve fail() failed: [line 0, column 0] Cannot divide by zero: 1 / 0][serve deposit(1) -> 101]",
			"Method 'fail' of server 'Account' failed: [line 0, column 0] Cannot divide by zero: 1 / 0\n"},
		{"without calls", fakeTransport{}, "[listen Account at localhost:3825]", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := serveProgram(t, &testCase.transport, accountProgram)

			if result.err != nil {
				t.Fatalf("Serve() unexpected error: %v", result.err)
			}
			if result.output != testCase.expectedOutput {
				t.Errorf("Serve() output = %q; want %q", result.output, testCase.expectedOutput)
			}
			if result.errorOutput != testCase.expectedErrorOutput {
				t.Errorf("Serve() error output = %q; want %q", result.errorOutput, testCase.expectedErrorOutput)
			}
			if result.interpreter.currentEnvironment != result.globalEnvironment || result.interpreter.callDepth != 0 {
				t.Errorf("Serve() left the environment %p at call depth %d; want the global environment %p at depth 0",
					result.interpreter.currentEnvironment, result.interpreter.callDepth, result.globalEnvironment)
			}
		})
	}
}

func TestServeError(t *testing.T) {
	testCases := []serveErrorTestCase{
		{"address that cannot be used", fakeTransport{listenError: errors.New("address already in use")},
			"[line 3, column 18] Cannot listen on localhost:3825: address already in use"},
		{"server that stops", fakeTransport{serverCalls: []serverCall{callOf("current")}, serveError: errors.New("accept failed")},
			"[line 3, column 18] Server 'Account' stopped: accept failed"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := serveProgram(t, &testCase.transport, accountProgram)

			if result.err == nil || result.err.Error() != testCase.expectedMessage {
				t.Errorf("Serve() error = %v; want %q", result.err, testCase.expectedMessage)
			}
		})
	}
}

func TestServeWithoutServer(t *testing.T) {
	result := serveProgram(t, &fakeTransport{}, func(d distanceMap) []common.Statement {
		return []common.Statement{printStatement(stringLiteral("client"))}
	})

	if result.err != nil {
		t.Errorf("Serve() unexpected error: %v", result.err)
	}
	if result.output != "client" {
		t.Errorf("Serve() output = %q; want %q", result.output, "client")
	}
}

func TestServeWithoutNetwork(t *testing.T) {
	result := serveProgram(t, nil, accountProgram)

	if expected := "[line 3, column 18] Cannot listen on localhost:3825: network runtime not available"; result.err == nil || result.err.Error() != expected {
		t.Errorf("Serve() error = %v; want %q", result.err, expected)
	}
}

func TestServeMethodThatIsNotInTheServerPanics(t *testing.T) {
	distances := distanceMap{}
	interpreter := NewInterpreter([]common.Statement{distances.accountServer()}, distances, io.Discard)
	interpreter.transport = &fakeTransport{output: io.Discard, serverCalls: []serverCall{callOf("transfer")}}
	if err := interpreter.Interpret(); err != nil {
		t.Fatalf("Interpret() unexpected error: %v", err)
	}

	assertPanics(t, "Serve() with a call to a method that is not in the server", func() { interpreter.Serve() })
}
