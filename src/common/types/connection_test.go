package types_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

type connectionRepresentationTestCase struct {
	name     string
	value    types.Connection
	expected string
}

type connectionEqualityTestCase struct {
	name     string
	left     types.Connection
	right    types.Value
	expected bool
}

func connection(service string, host string, port int) types.Connection {
	return types.NewConnection(service, host, port)
}

func TestConnectionRepresentation(t *testing.T) {
	testCases := []connectionRepresentationTestCase{
		{"host and port", connection("Account", "localhost", 3825), "connection<Account> at localhost:3825"},
		{"ip address", connection("Account", "127.0.0.1", 1), "connection<Account> at 127.0.0.1:1"},
		{"ipv6 address", connection("Counter", "::1", 3825), "connection<Counter> at [::1]:3825"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.value.String()
			displayed := testCase.value.Display()

			if result != testCase.expected || displayed != testCase.expected {
				t.Errorf("(%#v).String() = %q, Display() = %q; want %q", testCase.value, result, displayed, testCase.expected)
			}
		})
	}
}

func TestConnectionServiceHostAndPort(t *testing.T) {
	value := connection("Account", "localhost", 3825)

	if service := value.Service(); service != "Account" {
		t.Errorf("(%v).Service() = %q; want %q", value, service, "Account")
	}
	if host := value.Host(); host != "localhost" {
		t.Errorf("(%v).Host() = %q; want %q", value, host, "localhost")
	}
	if port := value.Port(); port != 3825 {
		t.Errorf("(%v).Port() = %d; want %d", value, port, 3825)
	}
}

func TestConnectionEquals(t *testing.T) {
	testCases := []connectionEqualityTestCase{
		{"same service and address", connection("Account", "a", 1), connection("Account", "a", 1), true},
		{"different host", connection("Account", "a", 1), connection("Account", "b", 1), false},
		{"different port", connection("Account", "a", 1), connection("Account", "a", 2), false},
		{"different service", connection("Account", "a", 1), connection("Counter", "a", 1), false},
		{"same address written differently", connection("Account", "localhost", 1), connection("Account", "127.0.0.1", 1), false},
		{"connection and session of the same service", connection("Account", "a", 1), types.NewSession("Account"), false},
		{"connection and string with its address", connection("Account", "a", 1), stringValue("a:1"), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Equals(testCase.right)

			if result != testCase.expected {
				t.Errorf("(%v).Equals(%v) = %v; want %v", testCase.left, testCase.right, result, testCase.expected)
			}
		})
	}
}

func TestStructWithConnectionFieldEquals(t *testing.T) {
	testCases := []structEqualityTestCase{
		{"equal connection fields", structValue("Config", field("account", connection("Account", "a", 1))),
			structValue("Config", field("account", connection("Account", "a", 1))), true},
		{"different connection fields", structValue("Config", field("account", connection("Account", "a", 1))),
			structValue("Config", field("account", connection("Account", "a", 2))), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.left.Equals(testCase.right)

			if result != testCase.expected {
				t.Errorf("(%v).Equals(%v) = %v; want %v", testCase.left, testCase.right, result, testCase.expected)
			}
		})
	}
}
