package types_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common/types"
)

func TestSessionRepresentation(t *testing.T) {
	value := types.NewSession("Account")

	if result := value.String(); result != "session<Account>" {
		t.Errorf("(%v).String() = %q; want %q", value, result, "session<Account>")
	}
	if result := value.Display(); result != "session<Account>" {
		t.Errorf("(%v).Display() = %q; want %q", value, result, "session<Account>")
	}
}

func TestSessionService(t *testing.T) {
	value := types.NewSession("Account")

	if service := value.Service(); service != "Account" {
		t.Errorf("(%v).Service() = %q; want %q", value, service, "Account")
	}
}

func TestSessionEquals(t *testing.T) {
	session := types.NewSession("Account")
	testCases := []struct {
		name     string
		left     *types.Session
		right    types.Value
		expected bool
	}{
		{"same session", session, session, true},
		{"different sessions with the same service", session, types.NewSession("Account"), false},
		{"sessions with different services", session, types.NewSession("Counter"), false},
		{"session and connection of the same service", session, connection("Account", "a", 1), false},
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

func TestNewSessionIdentity(t *testing.T) {
	if types.NewSession("Account") == types.NewSession("Account") {
		t.Errorf("two NewSession calls with the same service returned the same *Session")
	}
}
