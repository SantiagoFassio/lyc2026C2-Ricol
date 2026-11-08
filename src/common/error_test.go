package common_test

import (
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
)

type errorStringTestCase struct {
	name     string
	err      error
	expected string
}

func TestRicolErrorString(t *testing.T) {
	testCases := []errorStringTestCase{
		{
			"error without file",
			common.NewRicolError(common.Position{Line: 3, Column: 5}, "Undefined type 'Foo'"),
			"[line 3, column 5] Undefined type 'Foo'",
		},
		{
			"error with file",
			common.NewRicolError(common.Position{File: "account.ric", Line: 3, Column: 5}, "Undefined type 'Foo'"),
			"[account.ric, line 3, column 5] Undefined type 'Foo'",
		},
		{
			"error with a relative path",
			common.NewRicolError(common.Position{File: "../examples/account.ric", Line: 1, Column: 1}, "Unterminated string"),
			"[../examples/account.ric, line 1, column 1] Unterminated string",
		},
		{
			"list of errors from different files",
			common.RicolErrorList{
				common.NewRicolError(common.Position{File: "account.ric", Line: 2, Column: 8}, "Undefined type 'Foo'"),
				common.NewRicolError(common.Position{File: "main.ric", Line: 4, Column: 1}, "Undefined variable 'x'"),
				common.NewRicolError(common.Position{Line: 5, Column: 3}, "'break' outside loop"),
			},
			"[account.ric, line 2, column 8] Undefined type 'Foo'\n" +
				"[main.ric, line 4, column 1] Undefined variable 'x'\n" +
				"[line 5, column 3] 'break' outside loop",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.err.Error()

			if result != testCase.expected {
				t.Errorf("Error() = %q; want %q", result, testCase.expected)
			}
		})
	}
}
