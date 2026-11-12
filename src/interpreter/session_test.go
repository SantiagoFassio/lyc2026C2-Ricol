package interpreter

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type fakeTransport struct {
	output       io.Writer
	openFailures map[string]*failure
	results      map[string]types.Value
	failures     map[string]*failure
}

func (f *fakeTransport) open(service string, address common.Address, timeout time.Duration) (channel, *failure) {
	fmt.Fprintf(f.output, "[open %s at %v, %v]", service, address, timeout)
	if cause, ok := f.openFailures[service]; ok {
		return nil, cause
	}
	return &fakeChannel{transport: f, service: service}, nil
}

type fakeChannel struct {
	transport *fakeTransport
	service   string
}

func (c *fakeChannel) call(method string, arguments []types.Value) (types.Value, *failure) {
	values := make([]string, len(arguments))
	for index, argument := range arguments {
		values[index] = argument.String()
	}
	fmt.Fprintf(c.transport.output, "[call %s.%s(%s)]", c.service, method, strings.Join(values, ", "))
	if cause, ok := c.transport.failures[c.service+"."+method]; ok {
		return nil, cause
	}
	return c.transport.results[c.service+"."+method], nil
}

func (c *fakeChannel) close() {
	fmt.Fprintf(c.transport.output, "[close %s]", c.service)
}

type sessionTestCase struct {
	name           string
	transport      fakeTransport
	statements     func(d distanceMap) []common.Statement
	expectedOutput string
}

type sessionErrorTestCase struct {
	name            string
	transport       fakeTransport
	statements      func(d distanceMap) []common.Statement
	expectedOutput  string
	expectedMessage string
}

func interpretWithTransport(transport fakeTransport, statements func(d distanceMap) []common.Statement) (string, error) {
	distances := distanceMap{}
	program := statements(distances)
	output := &bytes.Buffer{}
	transport.output = output
	interpreter := NewInterpreter(program, distances, output)
	interpreter.transport = &transport
	err := interpreter.Interpret()
	return output.String(), err
}

func runSessionTestCases(t *testing.T, testCases []sessionTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output, err := interpretWithTransport(testCase.transport, testCase.statements)

			if err != nil {
				t.Fatalf("Interpret() unexpected error: %v", err)
			}
			if output != testCase.expectedOutput {
				t.Errorf("Interpret() output = %q; want %q", output, testCase.expectedOutput)
			}
		})
	}
}

func runSessionErrorTestCases(t *testing.T, testCases []sessionErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output, err := interpretWithTransport(testCase.transport, testCase.statements)

			if err == nil {
				t.Fatalf("Interpret() = nil error; want %q", testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("Interpret() error = %q; want %q", err, testCase.expectedMessage)
			}
			if output != testCase.expectedOutput {
				t.Errorf("Interpret() output = %q; want %q", output, testCase.expectedOutput)
			}
		})
	}
}

func connectExpression(service string, address string) *common.ConnectExpression {
	return common.NewConnectExpression(
		token(common.CONNECT, "connect"),
		token(common.IDENTIFIER, service),
		token(common.STRING, strconv.Quote(address)),
		address,
	)
}

func sessionStatement(name string, connection common.Expression, body []common.Statement, errorName string, rescueBody ...common.Statement) *common.SessionStatement {
	return common.NewSessionStatement(
		token(common.SESSION, "session"),
		token(common.IDENTIFIER, name),
		connection,
		common.Token{},
		nil,
		append([]common.Statement{}, body...),
		token(common.IDENTIFIER, errorName),
		append([]common.Statement{}, rescueBody...),
	)
}

func withTimeoutAt(line int, column int, timeout common.Expression, statement *common.SessionStatement) *common.SessionStatement {
	timed := *statement
	timed.TimeoutToken = common.NewToken(common.TIMEOUT, "timeout", common.Position{Line: line, Column: column})
	timed.Timeout = timeout
	return &timed
}

func (d distanceMap) remoteCall(session string, distance int, method string, arguments ...common.Expression) *common.MemberCallExpression {
	return common.NewMemberCallExpression(d.variable(session, distance), token(common.IDENTIFIER, method), append([]common.Expression{}, arguments...))
}

func accountConnection() *common.VarDeclarationStatement {
	return varDeclaration("conn", types.NamedConnectionType{Service: types.NamedType{Name: "Account"}}, connectExpression("Account", "localhost:3825"))
}

func (d distanceMap) accountSession(connectionDistance int, body []common.Statement, rescueBody ...common.Statement) *common.SessionStatement {
	return sessionStatement("c", d.variable("conn", connectionDistance), body, "e", rescueBody...)
}

func (d distanceMap) depositTwice() *common.FuncDeclarationStatement {
	return funcStatement("depositTwice", []common.Parameter{parameter("account", types.NamedSessionType{Service: types.NamedType{Name: "Account"}})}, types.Int,
		common.NewExpressionStatement(d.remoteCall("account", 0, "deposit", integerLiteral(1))),
		returnStatement(d.remoteCall("account", 0, "deposit", integerLiteral(1))),
	)
}

func accountResults() map[string]types.Value {
	return map[string]types.Value{"Account.balance": types.NewInteger(100), "Account.deposit": types.NewInteger(150)}
}

func TestConnectExpressionEvaluate(t *testing.T) {
	runEvaluateTestCases(t, []evaluateTestCase{
		{"connect", connectExpression("Account", "localhost:3825"), types.NewConnection("Account", "localhost", 3825)},
		{"connect to an ipv6 address", connectExpression("Account", "[::1]:3825"), types.NewConnection("Account", "::1", 3825)},
		{"connect with a port with leading zeros", connectExpression("Account", "localhost:03825"), types.NewConnection("Account", "localhost", 3825)},
		{"connect to an ipv6 address in its long form", connectExpression("Account", "[0:0:0:0:0:0:0:1]:3825"), types.NewConnection("Account", "::1", 3825)},
		{
			"equal connections",
			binary(connectExpression("Account", "localhost:3825"), common.DOUBLE_EQUAL, "==", connectExpression("Account", "localhost:3825")),
			types.NewBoolean(true),
		},
		{
			"connections to different addresses",
			binary(connectExpression("Account", "localhost:3825"), common.NOT_EQUAL, "!=", connectExpression("Account", "localhost:3826")),
			types.NewBoolean(true),
		},
		{
			"connections to the same port written differently",
			binary(connectExpression("Account", "localhost:3825"), common.DOUBLE_EQUAL, "==", connectExpression("Account", "localhost:03825")),
			types.NewBoolean(true),
		},
		{
			"connections to the same ipv6 address written differently",
			binary(connectExpression("Account", "[::1]:3825"), common.DOUBLE_EQUAL, "==", connectExpression("Account", "[0::1]:3825")),
			types.NewBoolean(true),
		},
		{
			"connections to different services",
			binary(connectExpression("Account", "localhost:3825"), common.DOUBLE_EQUAL, "==", connectExpression("Other", "localhost:3825")),
			types.NewBoolean(false),
		},
	})
}

func TestFailureValue(t *testing.T) {
	message := func(text string) []types.Value { return []types.Value{types.NewString(text)} }
	testCases := []struct {
		cause    failure
		expected types.Value
	}{
		{failure{variant: "ConnectionFailed", message: "refused"}, types.NewEnum("SessionError", "ConnectionFailed", message("refused"))},
		{failure{variant: "ConnectionLost", message: "reset"}, types.NewEnum("SessionError", "ConnectionLost", message("reset"))},
		{failure{variant: "Timeout"}, types.NewEnum("SessionError", "Timeout", nil)},
		{failure{variant: "ProtocolError", message: "unknown method"}, types.NewEnum("SessionError", "ProtocolError", message("unknown method"))},
		{failure{variant: "ServerError", message: "division by zero"}, types.NewEnum("SessionError", "ServerError", message("division by zero"))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.cause.variant, func(t *testing.T) {
			if value := testCase.cause.value(); !value.Equals(testCase.expected) {
				t.Errorf("value of %+v = %v; want %v", testCase.cause, value, testCase.expected)
			}
		})
	}

	t.Run("unknown variant", func(t *testing.T) {
		assertPanics(t, "value of an unknown failure", func() { (&failure{variant: "Refused"}).value() })
	})
}

func TestSessionStatementExecute(t *testing.T) {
	transport := fakeTransport{results: accountResults()}

	runSessionTestCases(t, []sessionTestCase{
		{"empty body", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, nil, printStatement(stringLiteral("rescued")))}
		}, "[open Account at localhost:3825, 5s][close Account]"},
		{"remote calls", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{
				printStatement(d.remoteCall("c", 0, "balance")),
				printStatement(d.remoteCall("c", 0, "deposit", integerLiteral(50))),
				common.NewExpressionStatement(d.remoteCall("c", 0, "reset")),
			})}
		}, "[open Account at localhost:3825, 5s][call Account.balance()]100[call Account.deposit(50)]150[call Account.reset()][close Account]"},
		{"arguments of a remote call", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{
				common.NewExpressionStatement(d.remoteCall("c", 0, "transfer",
					stringLiteral("savings"), binary(integerLiteral(20), common.PLUS, "+", integerLiteral(30)), booleanLiteral(true))),
			})}
		}, `[open Account at localhost:3825, 5s][call Account.transfer("savings", 50, True)][close Account]`},
		{"statements after the session", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, nil), printStatement(stringLiteral("after"))}
		}, "[open Account at localhost:3825, 5s][close Account]after"},
		{"printed session", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{printStatement(d.variable("c", 0))})}
		}, "[open Account at localhost:3825, 5s]session<Account>[close Account]"},
		{"variables of the body", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{
				varDeclaration("total", types.Int, d.remoteCall("c", 0, "balance")),
				printStatement(binary(d.variable("total", 0), common.PLUS, "+", integerLiteral(1))),
			})}
		}, "[open Account at localhost:3825, 5s][call Account.balance()]101[close Account]"},
		{"timeout", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), withTimeoutAt(2, 18, integerLiteral(1500), d.accountSession(0, nil))}
		}, "[open Account at localhost:3825, 1.5s][close Account]"},
		{"highest timeout", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), withTimeoutAt(2, 18, integerLiteral(9223372036854), d.accountSession(0, nil))}
		}, "[open Account at localhost:3825, 2562047h47m16.854s][close Account]"},
		{"address normalized before opening the session", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{sessionStatement("c", connectExpression("Account", "[0:0::1]:03825"), nil, "e")}
		}, "[open Account at [::1]:3825, 5s][close Account]"},
		{"timeout expression", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				varDeclaration("seconds", types.Int, integerLiteral(2)),
				withTimeoutAt(2, 18, binary(d.variable("seconds", 0), common.STAR, "*", integerLiteral(1000)), d.accountSession(0, nil)),
			}
		}, "[open Account at localhost:3825, 2s][close Account]"},
		{"connect in the header", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{sessionStatement("o", connectExpression("Other", "[::1]:3826"), nil, "e")}
		}, "[open Other at [::1]:3826, 5s][close Other]"},
		{"session passed to a function", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				d.depositTwice(),
				d.accountSession(0, []common.Statement{printStatement(d.call("depositTwice", 1, d.variable("c", 0)))}),
			}
		}, "[open Account at localhost:3825, 5s][call Account.deposit(1)][call Account.deposit(1)]150[close Account]"},
		{"session in a loop", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				varDeclaration("i", types.Int, integerLiteral(0)),
				whileStatement(binary(d.variable("i", 0), common.LESS, "<", integerLiteral(2)), block(
					common.NewExpressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
					d.accountSession(1, []common.Statement{printStatement(d.remoteCall("c", 0, "balance"))}),
				)),
			}
		}, "[open Account at localhost:3825, 5s][call Account.balance()]100[close Account]" +
			"[open Account at localhost:3825, 5s][call Account.balance()]100[close Account]"},
	})
}

func TestSessionRescueExecute(t *testing.T) {
	refused := map[string]*failure{"Account": {variant: "ConnectionFailed", message: "connection refused"}}
	depositLost := map[string]*failure{"Account.deposit": {variant: "ConnectionLost", message: "reset by peer"}}

	runSessionTestCases(t, []sessionTestCase{
		{"failure opening the session", fakeTransport{openFailures: refused}, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{printStatement(stringLiteral("body"))},
				printStatement(d.variable("e", 0)))}
		}, `[open Account at localhost:3825, 5s]SessionError.ConnectionFailed("connection refused")`},
		{"failure in a remote call", fakeTransport{results: accountResults(), failures: depositLost}, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{
				printStatement(stringLiteral("before")),
				common.NewExpressionStatement(d.remoteCall("c", 0, "deposit", integerLiteral(1))),
				printStatement(stringLiteral("after")),
			}, matchStatement(d.variable("e", 0),
				matchArm(pattern("ConnectionLost", "message"), printStatement(d.variable("message", 0))),
				matchArm(wildcard(), printStatement(stringLiteral("other"))),
			))}
		}, "[open Account at localhost:3825, 5s]before[call Account.deposit(1)][close Account]reset by peer"},
		{"failure in a function called from the body", fakeTransport{failures: depositLost}, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				d.depositTwice(),
				d.accountSession(0, []common.Statement{
					printStatement(d.call("depositTwice", 1, d.variable("c", 0))),
				}, printStatement(stringLiteral("rescued"))),
			}
		}, "[open Account at localhost:3825, 5s][call Account.deposit(1)][close Account]rescued"},
		{"failure in a loop of the body", fakeTransport{failures: depositLost}, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{
				whileStatement(booleanLiteral(true), block(common.NewExpressionStatement(d.remoteCall("c", 1, "deposit", integerLiteral(1))))),
			}, printStatement(stringLiteral("rescued")))}
		}, "[open Account at localhost:3825, 5s][call Account.deposit(1)][close Account]rescued"},
		{"statements after a rescued session", fakeTransport{openFailures: refused}, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				d.accountSession(0, nil, printStatement(stringLiteral("rescued"))),
				printStatement(stringLiteral("after")),
			}
		}, "[open Account at localhost:3825, 5s]rescuedafter"},
		{"rescue sees the names around the session", fakeTransport{openFailures: refused}, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				varDeclaration("c", types.Int, integerLiteral(7)),
				d.accountSession(0, nil, printStatement(d.variable("c", 1))),
			}
		}, "[open Account at localhost:3825, 5s]7"},
		{"return from the rescue", fakeTransport{openFailures: refused}, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				funcStatement("f", nil, types.Int, d.accountSession(1,
					[]common.Statement{returnStatement(d.remoteCall("c", 0, "balance"))},
					returnStatement(integerLiteral(0)),
				)),
				printStatement(d.call("f", 0)),
			}
		}, "[open Account at localhost:3825, 5s]0"},
	})
}

func TestNestedSessionsExecute(t *testing.T) {
	timeout := &failure{variant: "Timeout"}
	nested := func(d distanceMap) []common.Statement {
		return []common.Statement{
			accountConnection(),
			sessionStatement("a", d.variable("conn", 0), []common.Statement{
				sessionStatement("o", connectExpression("Other", "localhost:3826"), []common.Statement{
					common.NewExpressionStatement(d.remoteCall("o", 0, "ping")),
					common.NewExpressionStatement(d.remoteCall("a", 1, "deposit", integerLiteral(1))),
					printStatement(stringLiteral("inner body")),
				}, "inner",
					printStatement(stringLiteral("inner rescue")),
					printStatement(d.remoteCall("a", 1, "balance")),
				),
				printStatement(stringLiteral("outer body")),
			}, "outer", printStatement(stringLiteral("outer rescue"))),
		}
	}

	runSessionTestCases(t, []sessionTestCase{
		{"without failures", fakeTransport{results: accountResults()}, nested,
			"[open Account at localhost:3825, 5s][open Other at localhost:3826, 5s][call Other.ping()][call Account.deposit(1)]inner body" +
				"[close Other]outer body[close Account]"},
		{"failure of the inner session goes to the inner rescue", fakeTransport{
			results:  accountResults(),
			failures: map[string]*failure{"Other.ping": timeout},
		}, nested,
			"[open Account at localhost:3825, 5s][open Other at localhost:3826, 5s][call Other.ping()][close Other]inner rescue" +
				"[call Account.balance()]100outer body[close Account]"},
		{"failure of the outer session inside the inner one goes to the outer rescue", fakeTransport{
			failures: map[string]*failure{"Account.deposit": timeout},
		}, nested,
			"[open Account at localhost:3825, 5s][open Other at localhost:3826, 5s][call Other.ping()][call Account.deposit(1)]" +
				"[close Other][close Account]outer rescue"},
		{"failure opening the inner session", fakeTransport{
			results:      accountResults(),
			openFailures: map[string]*failure{"Other": timeout},
		}, nested,
			"[open Account at localhost:3825, 5s][open Other at localhost:3826, 5s]inner rescue[call Account.balance()]100outer body[close Account]"},
		{"failure of the outer session in the inner rescue", fakeTransport{
			openFailures: map[string]*failure{"Other": timeout},
			failures:     map[string]*failure{"Account.balance": timeout},
		}, nested,
			"[open Account at localhost:3825, 5s][open Other at localhost:3826, 5s]inner rescue[call Account.balance()][close Account]outer rescue"},
	})
}

func TestSessionClosesTheConnection(t *testing.T) {
	transport := fakeTransport{results: accountResults()}

	runSessionTestCases(t, []sessionTestCase{
		{"return from the body", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				funcStatement("f", nil, types.Int, d.accountSession(1,
					[]common.Statement{returnStatement(d.remoteCall("c", 0, "balance"))},
					returnStatement(integerLiteral(0)),
				)),
				printStatement(d.call("f", 0)),
			}
		}, "[open Account at localhost:3825, 5s][call Account.balance()][close Account]100"},
		{"break from the body", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				whileStatement(booleanLiteral(true), block(d.accountSession(1, []common.Statement{breakStatement()}))),
				printStatement(stringLiteral("after")),
			}
		}, "[open Account at localhost:3825, 5s][close Account]after"},
		{"continue from the body", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{
				accountConnection(),
				varDeclaration("i", types.Int, integerLiteral(0)),
				whileStatement(binary(d.variable("i", 0), common.LESS, "<", integerLiteral(2)), block(
					common.NewExpressionStatement(d.assignment("i", 1, binary(d.variable("i", 1), common.PLUS, "+", integerLiteral(1)))),
					d.accountSession(1, []common.Statement{continueStatement()}),
					printStatement(stringLiteral("unreachable")),
				)),
			}
		}, "[open Account at localhost:3825, 5s][close Account][open Account at localhost:3825, 5s][close Account]"},
	})

	runSessionErrorTestCases(t, []sessionErrorTestCase{
		{"local error in the body", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{
				printStatement(stringLiteral("before")),
				printStatement(divisionByZero()),
			}, printStatement(stringLiteral("rescued")))}
		}, "[open Account at localhost:3825, 5s]before[close Account]", "[line 0, column 0] Cannot divide by zero: 1 / 0"},
		{"local error in an argument of a remote call", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, []common.Statement{
				common.NewExpressionStatement(d.remoteCall("c", 0, "deposit", divisionByZero())),
			}, printStatement(stringLiteral("rescued")))}
		}, "[open Account at localhost:3825, 5s][close Account]", "[line 0, column 0] Cannot divide by zero: 1 / 0"},
	})
}

func TestSessionStatementExecuteError(t *testing.T) {
	transport := fakeTransport{results: accountResults()}
	refused := fakeTransport{openFailures: map[string]*failure{"Account": {variant: "ConnectionFailed", message: "connection refused"}}}

	runSessionErrorTestCases(t, []sessionErrorTestCase{
		{"zero timeout", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), withTimeoutAt(2, 18, integerLiteral(0),
				d.accountSession(0, []common.Statement{printStatement(stringLiteral("body"))}, printStatement(stringLiteral("rescued"))))}
		}, "", "[line 2, column 18] Session timeout must be positive, got 0"},
		{"negative timeout", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), withTimeoutAt(2, 18, negation(integerLiteral(5)), d.accountSession(0, nil))}
		}, "", "[line 2, column 18] Session timeout must be positive, got -5"},
		{"timeout above the highest", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), withTimeoutAt(2, 18, integerLiteral(9223372036855), d.accountSession(0, nil))}
		}, "", "[line 2, column 18] Session timeout must be at most 9223372036854, got 9223372036855"},
		{"error in the connection", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{sessionStatement("c", grouping(divisionByZero()), nil, "e", printStatement(stringLiteral("rescued")))}
		}, "", "[line 0, column 0] Cannot divide by zero: 1 / 0"},
		{"error in the timeout", transport, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), withTimeoutAt(2, 18, moduloByZero(), d.accountSession(0, nil, printStatement(stringLiteral("rescued"))))}
		}, "", "[line 0, column 0] Cannot divide by zero: 2 % 0"},
		{"error in the rescue", refused, func(d distanceMap) []common.Statement {
			return []common.Statement{accountConnection(), d.accountSession(0, nil,
				printStatement(stringLiteral("rescue")),
				printStatement(divisionByZero()),
				printStatement(stringLiteral("after the error")),
			)}
		}, "[open Account at localhost:3825, 5s]rescue", "[line 0, column 0] Cannot divide by zero: 1 / 0"},
	})
}

func TestSessionsWithoutNetworkGoToTheRescue(t *testing.T) {
	distances := distanceMap{}
	program := []common.Statement{
		accountConnection(),
		distances.accountSession(0, []common.Statement{printStatement(stringLiteral("connected"))},
			matchStatement(distances.variable("e", 0),
				matchArm(pattern("ConnectionFailed", "message"), printStatement(distances.variable("message", 0))),
				matchArm(wildcard(), printStatement(stringLiteral("other"))),
			)),
	}
	output := &bytes.Buffer{}

	if err := NewInterpreter(program, distances, output).Interpret(); err != nil {
		t.Fatalf("Interpret() unexpected error: %v", err)
	}

	if expected := "cannot connect to Account at localhost:3825: network runtime not available"; output.String() != expected {
		t.Errorf("Interpret() output = %q; want %q", output, expected)
	}
}

func TestSessionStatementRestoresTheEnvironment(t *testing.T) {
	testCases := []struct {
		name      string
		transport fakeTransport
		body      func(d distanceMap) []common.Statement
	}{
		{"after the body", fakeTransport{}, func(d distanceMap) []common.Statement {
			return []common.Statement{varDeclaration("x", types.Int, integerLiteral(1))}
		}},
		{"after the rescue", fakeTransport{openFailures: map[string]*failure{"Account": {variant: "Timeout"}}}, func(d distanceMap) []common.Statement {
			return nil
		}},
		{"after a failure in the body", fakeTransport{failures: map[string]*failure{"Account.reset": {variant: "Timeout"}}}, func(d distanceMap) []common.Statement {
			return []common.Statement{varDeclaration("x", types.Int, integerLiteral(1)), common.NewExpressionStatement(d.remoteCall("c", 0, "reset"))}
		}},
		{"after a local error in the body", fakeTransport{}, func(d distanceMap) []common.Statement {
			return []common.Statement{varDeclaration("x", types.Int, integerLiteral(1)), common.NewExpressionStatement(divisionByZero())}
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			distances := distanceMap{}
			statement := distances.accountSession(0, testCase.body(distances), varDeclaration("y", types.Int, integerLiteral(2)))
			interpreter := NewInterpreter(nil, distances, io.Discard)
			transport := testCase.transport
			transport.output = io.Discard
			interpreter.transport = &transport
			globalEnvironment := interpreter.currentEnvironment
			if err := interpreter.execute(accountConnection()); err != nil {
				t.Fatalf("execute(%s) unexpected error: %v", accountConnection(), err)
			}

			interpreter.execute(statement)

			if interpreter.currentEnvironment != globalEnvironment {
				t.Errorf("execute(%s) did not restore the global environment", statement)
			}
			for _, name := range []string{"c", "e", "x", "y"} {
				if _, ok := globalEnvironment.values[name]; ok {
					t.Errorf("execute(%s) left '%s' in the global environment", statement, name)
				}
			}
			if len(interpreter.channels) != 0 {
				t.Errorf("execute(%s) left %d open channels", statement, len(interpreter.channels))
			}
		})
	}
}

func TestSessionErrorIsPredefined(t *testing.T) {
	distances := distanceMap{}
	timeout := distances.variant("SessionError", 0, "Timeout")
	interpreter := NewInterpreter(nil, distances, io.Discard)

	if !interpreter.currentEnvironment.isEnum("SessionError", 0) {
		t.Fatalf("the global environment does not define the enum SessionError")
	}
	result, err := interpreter.evaluate(timeout)
	if err != nil {
		t.Fatalf("evaluate(%s) unexpected error: %v", timeout, err)
	}
	if expected := types.NewEnum("SessionError", "Timeout", nil); !result.Equals(expected) {
		t.Errorf("evaluate(%s) = %v; want %v", timeout, result, expected)
	}
}

func TestSessionInvariantPanics(t *testing.T) {
	testCases := []struct {
		name       string
		statements func(interpreter *Interpreter, d distanceMap) []common.Statement
	}{
		{"session on a value that is not a connection", func(interpreter *Interpreter, d distanceMap) []common.Statement {
			return []common.Statement{sessionStatement("c", integerLiteral(1), nil, "e")}
		}},
		{"timeout that is not an integer", func(interpreter *Interpreter, d distanceMap) []common.Statement {
			return []common.Statement{withTimeoutAt(0, 0, floatLiteral(1.5), sessionStatement("c", connectExpression("Account", "localhost:3825"), nil, "e"))}
		}},
		{"connect with an invalid address", func(interpreter *Interpreter, d distanceMap) []common.Statement {
			return []common.Statement{sessionStatement("c", connectExpression("Account", "localhost"), nil, "e")}
		}},
		{"remote call on a closed session", func(interpreter *Interpreter, d distanceMap) []common.Statement {
			interpreter.currentEnvironment.define("c", types.NewSession("Account"))
			return []common.Statement{common.NewExpressionStatement(d.remoteCall("c", 0, "balance"))}
		}},
		{"session failure that escapes its session", func(interpreter *Interpreter, d distanceMap) []common.Statement {
			session := types.NewSession("Account")
			interpreter.currentEnvironment.define("c", session)
			failing := &fakeTransport{output: io.Discard, failures: map[string]*failure{"Account.balance": {variant: "Timeout"}}}
			interpreter.channels[session] = &fakeChannel{transport: failing, service: "Account"}
			return []common.Statement{common.NewExpressionStatement(d.remoteCall("c", 0, "balance"))}
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			distances := distanceMap{}
			interpreter := NewInterpreter(nil, distances, io.Discard)
			interpreter.transport = &fakeTransport{output: io.Discard}
			interpreter.statements = testCase.statements(interpreter, distances)

			defer func() {
				if recover() == nil {
					t.Errorf("Interpret(%v) did not panic", interpreter.statements)
				}
			}()

			interpreter.Interpret()
		})
	}
}
