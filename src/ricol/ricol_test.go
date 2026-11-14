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

type typeCheckTestCase struct {
	name  string
	files sourceFiles
}

type typeCheckErrorTestCase struct {
	name            string
	files           sourceFiles
	expectedMessage string
}

func typeCheck(t *testing.T, files sourceFiles) (string, error) {
	t.Helper()

	return run(t, files, ricol.ModeTypeCheck)
}

func run(t *testing.T, files sourceFiles, mode ricol.Mode) (string, error) {
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
	err := ricol.NewRicol("main.ric", mode, &output).Run()
	return output.String(), err
}

func runTypeCheckTestCases(t *testing.T, testCases []typeCheckTestCase) {
	t.Helper()

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

func runTypeCheckErrorTestCases(t *testing.T, testCases []typeCheckErrorTestCase) {
	t.Helper()

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

func TestImportedServiceIsAccepted(t *testing.T) {
	runTypeCheckTestCases(t, []typeCheckTestCase{
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
	})
}

func TestImportedServiceErrors(t *testing.T) {
	runTypeCheckErrorTestCases(t, []typeCheckErrorTestCase{
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
	})
}

func TestImportedFunctionsAreAccepted(t *testing.T) {
	runTypeCheckTestCases(t, []typeCheckTestCase{
		{
			"function calls a function of an imported module",
			sourceFiles{
				"lib.ric":  "func double(n: Int) -> Int {\n    return 2 * n;\n}\n",
				"main.ric": "import \"lib.ric\";\n\nfunc quadruple(n: Int) -> Int {\n    return double(double(n));\n}\n\nprint quadruple(1);\n",
			},
		},
		{
			"mutually recursive functions in a module",
			sourceFiles{
				"parity.ric": `func isEven(n: Int) -> Bool {
    if (n == 0) {
        return True;
    }
    return isOdd(n - 1);
}

func isOdd(n: Int) -> Bool {
    if (n == 0) {
        return False;
    }
    return isEven(n - 1);
}
`,
				"main.ric": "import \"parity.ric\";\n\nfunc check() -> Bool {\n    return isEven(4);\n}\n",
			},
		},
	})
}

func TestImportedFunctionErrors(t *testing.T) {
	runTypeCheckErrorTestCases(t, []typeCheckErrorTestCase{
		{
			"function of a module calls a function of the importer",
			sourceFiles{
				"helper.ric": "func helper() -> Int {\n    return fromMain();\n}\n",
				"main.ric":   "import \"helper.ric\";\n\nfunc fromMain() -> Int {\n    return 42;\n}\n\nprint helper();\n",
			},
			"[helper.ric, line 2, column 12] Undefined function 'fromMain'",
		},
		{
			"function of a module calls a function of a later module",
			sourceFiles{
				"first.ric":  "func first() -> Int {\n    return second();\n}\n",
				"second.ric": "func second() -> Int {\n    return 2;\n}\n",
				"main.ric":   "import \"first.ric\";\nimport \"second.ric\";\n\nprint first();\n",
			},
			"[first.ric, line 2, column 12] Undefined function 'second'",
		},
	})
}
