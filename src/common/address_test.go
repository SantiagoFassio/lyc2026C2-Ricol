package common_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
)

type addressTestCase struct {
	name       string
	rawAddress string
	expected   common.Address
}

type invalidAddressTestCase struct {
	name       string
	rawAddress string
}

func TestParseAddress(t *testing.T) {
	testCases := []addressTestCase{
		{"host name", "localhost:3825", common.Address{Host: "localhost", Port: 3825}},
		{"all interfaces", "0.0.0.0:3825", common.Address{Host: "0.0.0.0", Port: 3825}},
		{"lowest port", "0.0.0.0:1", common.Address{Host: "0.0.0.0", Port: 1}},
		{"highest port", "127.0.0.1:65535", common.Address{Host: "127.0.0.1", Port: 65535}},
		{"ipv6 address", "[::1]:3825", common.Address{Host: "::1", Port: 3825}},
		{"domain name", "bank.example.com:8080", common.Address{Host: "bank.example.com", Port: 8080}},
		{"port with leading zeros", "localhost:03825", common.Address{Host: "localhost", Port: 3825}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := common.ParseAddress(testCase.rawAddress)

			if !ok {
				t.Fatalf("ParseAddress(%q) is invalid; want %v", testCase.rawAddress, testCase.expected)
			}
			if result != testCase.expected {
				t.Errorf("ParseAddress(%q) = %v; want %v", testCase.rawAddress, result, testCase.expected)
			}
		})
	}
}

func TestParseInvalidAddress(t *testing.T) {
	testCases := []invalidAddressTestCase{
		{"empty", ""},
		{"tcp scheme", "tcp://localhost:3825"},
		{"other scheme", "http://localhost:3825"},
		{"missing host", ":3825"},
		{"missing port", "localhost:"},
		{"missing colon", "localhost"},
		{"only colon", ":"},
		{"port zero", "localhost:0"},
		{"port above the highest", "localhost:65536"},
		{"negative port", "localhost:-1"},
		{"port with sign", "localhost:+3825"},
		{"non numeric port", "localhost:http"},
		{"port with spaces", "localhost: 3825"},
		{"path after the port", "localhost:3825/account"},
		{"ipv6 address without brackets", "::1:3825"},
		{"unclosed brackets", "[::1:3825"},
		{"several ports", "localhost:3825:3826"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, ok := common.ParseAddress(testCase.rawAddress)

			if ok {
				t.Errorf("ParseAddress(%q) = %v; want invalid", testCase.rawAddress, result)
			}
		})
	}
}
