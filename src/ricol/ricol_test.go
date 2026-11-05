package ricol_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/ricol"
)

const accountService = `struct Movement {
    amount: Int;
}

service Account {
    func deposit(amount: Int) -> Movement;
}
`

const accountServer = `import "account.ric";

server Account at "0.0.0.0:3825" {
    func deposit(amount: Int) -> Movement {
        return Movement { amount: amount };
    }
}
`

type sourceFiles map[string]string

type importedServiceTestCase struct {
	name  string
	files sourceFiles
}

type importedServiceErrorTestCase struct {
	name            string
	files           sourceFiles
	expectedMessage string
}

func typeCheck(t *testing.T, files sourceFiles) (string, error) {
	t.Helper()

	t.Chdir(t.TempDir())
	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("os.MkdirAll(%q) unexpected error: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("os.WriteFile(%q) unexpected error: %v", path, err)
		}
	}
	var output bytes.Buffer
	err := ricol.NewRicol("main.ric", ricol.ModeTypeCheck, &output).Run()
	return output.String(), err
}

func TestImportedServiceIsAccepted(t *testing.T) {
	testCases := []importedServiceTestCase{
		{
			"service in a module and server in the main file",
			sourceFiles{
				"account.ric": accountService,
				"main.ric":    accountServer,
			},
		},
		{
			"types of the service in another module",
			sourceFiles{
				"types.ric":   "struct Movement {\n    amount: Int;\n}\n",
				"account.ric": "import \"types.ric\";\n\nservice Account {\n    func deposit(amount: Int) -> Movement;\n}\n",
				"main.ric":    accountServer,
			},
		},
		{
			"types imported by the service and by the server",
			sourceFiles{
				"types.ric":   "struct Movement {\n    amount: Int;\n}\n",
				"account.ric": "import \"types.ric\";\n\nservice Account {\n    func deposit(amount: Int) -> Movement;\n}\n",
				"main.ric":    "import \"types.ric\";\n" + accountServer,
			},
		},
		{
			"service and server in modules",
			sourceFiles{
				"account.ric": accountService,
				"server.ric":  accountServer,
				"main.ric":    `import "server.ric";`,
			},
		},
		{
			"service in a module of a subdirectory",
			sourceFiles{
				"lib/account.ric": accountService,
				"main.ric": `import "lib/account.ric";

server Account at "0.0.0.0:3825" {
    func deposit(amount: Int) -> Movement {
        return Movement { amount: amount };
    }
}
`,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output, err := typeCheck(t, testCase.files)

			if err != nil {
				t.Fatalf("ricol.Run() unexpected error: %v", err)
			}
			if output != "Type checking OK\n" {
				t.Errorf("ricol.Run() output = %q; want %q", output, "Type checking OK\n")
			}
		})
	}
}

func TestImportedServiceErrors(t *testing.T) {
	testCases := []importedServiceErrorTestCase{
		{
			"method with another signature",
			sourceFiles{
				"account.ric": accountService,
				"main.ric": `import "account.ric";

server Account at "0.0.0.0:3825" {
    func deposit(amount: String) -> Movement {
        return Movement { amount: 1 };
    }
}
`,
			},
			"[main.ric, line 4, column 10] Method 'deposit' of server 'Account' must have signature (Int) -> Movement, got (String) -> Movement",
		},
		{
			"method not implemented",
			sourceFiles{
				"account.ric": accountService,
				"main.ric":    "import \"account.ric\";\n\nserver Account at \"0.0.0.0:3825\" {\n}\n",
			},
			"[main.ric, line 3, column 8] Server 'Account' does not implement method 'deposit'",
		},
		{
			"service not imported",
			sourceFiles{
				"account.ric": accountService,
				"main.ric":    "server Account at \"0.0.0.0:3825\" {\n}\n",
			},
			"[main.ric, line 1, column 8] Undefined service 'Account'",
		},
		{
			"undefined type in the service of a module",
			sourceFiles{
				"account.ric": "service Account {\n    func deposit(amount: Missing);\n}\n",
				"main.ric":    `import "account.ric";`,
			},
			"[account.ric, line 2, column 26] Undefined type 'Missing'",
		},
		{
			"error in a server of a module of a subdirectory",
			sourceFiles{
				"lib/account.ric": accountService,
				"lib/server.ric":  "import \"account.ric\";\n\nserver Account at \"no-port\" {\n    func deposit(amount: Int) -> Movement {\n        return Movement { amount: amount };\n    }\n}\n",
				"main.ric":        `import "lib/server.ric";`,
			},
			"[lib/server.ric, line 3, column 19] Invalid address 'no-port': expected host:port",
		},
		{
			"servers in two files",
			sourceFiles{
				"account.ric": accountService,
				"server.ric":  accountServer,
				"main.ric":    "import \"server.ric\";\n\nserver Account at \"0.0.0.0:3826\" {\n    func deposit(amount: Int) -> Movement {\n        return Movement { amount: amount };\n    }\n}\n",
			},
			"[main.ric, line 3, column 1] Only one server is allowed per program",
		},
		{
			"service declared in two files",
			sourceFiles{
				"account.ric": accountService,
				"main.ric":    "import \"account.ric\";\n\nservice Account {\n    func balance() -> Int;\n}\n",
			},
			"[main.ric, line 3, column 9] Service 'Account' already declared in this scope",
		},
		{
			"import inside a server",
			sourceFiles{
				"account.ric": accountService,
				"main.ric":    "server Account at \"0.0.0.0:3825\" {\n    import \"account.ric\";\n}\n",
			},
			"[main.ric, line 2, column 5] Imports must be at the beginning of the file",
		},
		{
			"errors in several files are reported together",
			sourceFiles{
				"account.ric": "service Account {\n    func deposit(amount: Missing);\n}\n",
				"main.ric":    "import \"account.ric\";\n\nprint x;\n",
			},
			"[account.ric, line 2, column 26] Undefined type 'Missing'\n" +
				"[main.ric, line 3, column 7] Undefined variable 'x'",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := typeCheck(t, testCase.files)

			if err == nil {
				t.Fatalf("ricol.Run() = nil error; want %q", testCase.expectedMessage)
			}
			if err.Error() != testCase.expectedMessage {
				t.Errorf("ricol.Run() error = %q; want %q", err, testCase.expectedMessage)
			}
		})
	}
}
