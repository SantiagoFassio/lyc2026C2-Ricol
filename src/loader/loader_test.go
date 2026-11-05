package loader_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/loader"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/parser"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/scanner"
)

type sourceFiles map[string]string

type loaderTestCase struct {
	name            string
	mainPath        string
	files           sourceFiles
	expectedProgram string
}

type loaderErrorTestCase struct {
	name            string
	mainPath        string
	files           sourceFiles
	expectedMessage string
}

type loaderReadsTestCase struct {
	name          string
	files         sourceFiles
	expectedReads map[string]int
}

func (f sourceFiles) readFile(path string) ([]byte, error) {
	content, ok := f[path]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	return []byte(content), nil
}

func load(t *testing.T, mainPath string, files sourceFiles, readFile func(string) ([]byte, error)) ([]common.Statement, error) {
	t.Helper()

	tokens, err := scanner.NewScanner(files[mainPath], mainPath).Scan()
	if err != nil {
		t.Fatalf("scanner.Scan(%q) unexpected error: %v", files[mainPath], err)
	}
	statements, err := parser.NewParser(tokens).Parse()
	if err != nil {
		t.Fatalf("parser.Parse(%q) unexpected error: %v", files[mainPath], err)
	}
	return loader.NewLoader(mainPath, statements, readFile).Load()
}

func programString(statements []common.Statement) string {
	var stringBuilder strings.Builder
	for _, statement := range statements {
		stringBuilder.WriteString(statement.String())
	}
	return stringBuilder.String()
}

func assertLoad(t *testing.T, mainPath string, files sourceFiles, expectedProgram string) {
	t.Helper()

	program, err := load(t, mainPath, files, files.readFile)

	if err != nil {
		t.Fatalf("loader.Load(%q) unexpected error: %v", mainPath, err)
	}
	if result := programString(program); result != expectedProgram {
		t.Errorf("loader.Load(%q) = %q; want %q", mainPath, result, expectedProgram)
	}
}

func assertLoadError(t *testing.T, mainPath string, files sourceFiles, readFile func(string) ([]byte, error), expectedMessage string) {
	t.Helper()

	program, err := load(t, mainPath, files, readFile)

	if err == nil {
		t.Fatalf("loader.Load(%q) = nil error; want %q", mainPath, expectedMessage)
	}
	if err.Error() != expectedMessage {
		t.Errorf("loader.Load(%q) error = %q; want %q", mainPath, err, expectedMessage)
	}
	if len(program) != 0 {
		t.Errorf("loader.Load(%q) = %v; want no statements", mainPath, program)
	}
}

func runLoadTestCases(t *testing.T, testCases []loaderTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertLoad(t, testCase.mainPath, testCase.files, testCase.expectedProgram)
		})
	}
}

func runLoadErrorTestCases(t *testing.T, testCases []loaderErrorTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertLoadError(t, testCase.mainPath, testCase.files, testCase.files.readFile, testCase.expectedMessage)
		})
	}
}

func absolutePath(t *testing.T, path string) string {
	t.Helper()

	result, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("filepath.Abs(%q) unexpected error: %v", path, err)
	}
	return result
}

func TestLoad(t *testing.T) {
	runLoadTestCases(t, []loaderTestCase{
		{
			"without imports",
			"main.ric",
			sourceFiles{"main.ric": "print main;"},
			"print main;\n",
		},
		{
			"one import",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; print main;`,
				"a.ric":    "print a;",
			},
			"print a;\nprint main;\n",
		},
		{
			"several imports keep their order",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; import "b.ric"; print main;`,
				"a.ric":    "print a;",
				"b.ric":    "print b;",
			},
			"print a;\nprint b;\nprint main;\n",
		},
		{
			"transitive import",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; print main;`,
				"a.ric":    `import "b.ric"; print a;`,
				"b.ric":    "print b;",
			},
			"print b;\nprint a;\nprint main;\n",
		},
		{
			"module without statements",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; print main;`,
				"a.ric":    "",
			},
			"print main;\n",
		},
		{
			"module with only imports",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; print main;`,
				"a.ric":    `import "b.ric";`,
				"b.ric":    "print b;",
			},
			"print b;\nprint main;\n",
		},
		{
			"main file with only imports",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric";`,
				"a.ric":    "print a;",
			},
			"print a;\n",
		},
	})
}

func TestLoadDuplicatedImports(t *testing.T) {
	workingDirectory := filepath.Base(absolutePath(t, "."))
	runLoadTestCases(t, []loaderTestCase{
		{
			"same file imported twice",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; import "a.ric"; print main;`,
				"a.ric":    "print a;",
			},
			"print a;\nprint main;\n",
		},
		{
			"same file imported by two modules",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; import "b.ric"; print main;`,
				"a.ric":    `import "c.ric"; print a;`,
				"b.ric":    `import "c.ric"; print b;`,
				"c.ric":    "print c;",
			},
			"print c;\nprint a;\nprint b;\nprint main;\n",
		},
		{
			"module imported by the main file and by a later module",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; import "b.ric"; print main;`,
				"a.ric":    "print a;",
				"b.ric":    `import "a.ric"; print b;`,
			},
			"print a;\nprint b;\nprint main;\n",
		},
		{
			"same file through different paths",
			"main.ric",
			sourceFiles{
				"main.ric": `import "c.ric"; import "lib/../c.ric"; import "./c.ric"; print main;`,
				"c.ric":    "print c;",
			},
			"print c;\nprint main;\n",
		},
		{
			"same file through a path that leaves the directory",
			"main.ric",
			sourceFiles{
				"main.ric": `import "c.ric"; import "../` + workingDirectory + `/c.ric"; print main;`,
				"c.ric":    "print c;",
			},
			"print c;\nprint main;\n",
		},
		{
			"same file through relative and absolute paths",
			"main.ric",
			sourceFiles{
				"main.ric": `import "c.ric"; import "` + absolutePath(t, "c.ric") + `"; print main;`,
				"c.ric":    "print c;",
			},
			"print c;\nprint main;\n",
		},
	})
}

func TestLoadPaths(t *testing.T) {
	runLoadTestCases(t, []loaderTestCase{
		{
			"import of a subdirectory",
			"main.ric",
			sourceFiles{
				"main.ric":  `import "lib/a.ric"; print main;`,
				"lib/a.ric": "print a;",
			},
			"print a;\nprint main;\n",
		},
		{
			"path relative to the importing module",
			"main.ric",
			sourceFiles{
				"main.ric":  `import "lib/a.ric"; print main;`,
				"lib/a.ric": `import "b.ric"; print a;`,
				"lib/b.ric": "print b;",
			},
			"print b;\nprint a;\nprint main;\n",
		},
		{
			"path relative to a main file in a directory",
			"app/main.ric",
			sourceFiles{
				"app/main.ric": `import "a.ric"; print main;`,
				"app/a.ric":    "print a;",
			},
			"print a;\nprint main;\n",
		},
		{
			"import of a parent directory",
			"app/main.ric",
			sourceFiles{
				"app/main.ric": `import "../lib/a.ric"; print main;`,
				"lib/a.ric":    "print a;",
			},
			"print a;\nprint main;\n",
		},
		{
			"absolute path",
			"main.ric",
			sourceFiles{
				"main.ric":                   `import "` + absolutePath(t, "lib/a.ric") + `"; print main;`,
				absolutePath(t, "lib/a.ric"): "print a;",
			},
			"print a;\nprint main;\n",
		},
	})
}

func TestLoadReadsEachModuleOnce(t *testing.T) {
	workingDirectory := filepath.Base(absolutePath(t, "."))
	testCases := []loaderReadsTestCase{
		{
			"the main file is not read",
			sourceFiles{"main.ric": "print main;"},
			map[string]int{},
		},
		{
			"one read per module",
			sourceFiles{
				"main.ric": `import "a.ric"; import "b.ric";`,
				"a.ric":    "print a;",
				"b.ric":    "print b;",
			},
			map[string]int{"a.ric": 1, "b.ric": 1},
		},
		{
			"same file imported twice",
			sourceFiles{
				"main.ric": `import "a.ric"; import "a.ric";`,
				"a.ric":    "print a;",
			},
			map[string]int{"a.ric": 1},
		},
		{
			"main file imports x and z, and z also imports x",
			sourceFiles{
				"main.ric": `import "x.ric"; import "z.ric";`,
				"x.ric":    "print x;",
				"z.ric":    `import "x.ric"; print z;`,
			},
			map[string]int{"x.ric": 1, "z.ric": 1},
		},
		{
			"main file imports z and x, and z also imports x",
			sourceFiles{
				"main.ric": `import "z.ric"; import "x.ric";`,
				"x.ric":    "print x;",
				"z.ric":    `import "x.ric"; print z;`,
			},
			map[string]int{"x.ric": 1, "z.ric": 1},
		},
		{
			"same file imported by two modules",
			sourceFiles{
				"main.ric": `import "a.ric"; import "b.ric";`,
				"a.ric":    `import "c.ric";`,
				"b.ric":    `import "c.ric";`,
				"c.ric":    "print c;",
			},
			map[string]int{"a.ric": 1, "b.ric": 1, "c.ric": 1},
		},
		{
			"same file through different paths",
			sourceFiles{
				"main.ric": `import "c.ric"; import "./c.ric"; import "lib/../c.ric"; import "../` + workingDirectory + `/c.ric"; import "` + absolutePath(t, "c.ric") + `";`,
				"c.ric":    "print c;",
			},
			map[string]int{"c.ric": 1},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reads := map[string]int{}
			readFile := func(path string) ([]byte, error) {
				reads[path]++
				return testCase.files.readFile(path)
			}

			_, err := load(t, "main.ric", testCase.files, readFile)

			if err != nil {
				t.Fatalf("loader.Load(%q) unexpected error: %v", "main.ric", err)
			}
			if diff := cmp.Diff(testCase.expectedReads, reads); diff != "" {
				t.Errorf("loader.Load(%q) reads mismatch (-want +got):\n%s", "main.ric", diff)
			}
		})
	}
}

func TestLoadModulesKeepTheirFile(t *testing.T) {
	files := sourceFiles{
		"main.ric":  `import "lib/a.ric"; print main;`,
		"lib/a.ric": "print a;",
	}

	program, err := load(t, "main.ric", files, files.readFile)

	if err != nil {
		t.Fatalf("loader.Load(%q) unexpected error: %v", "main.ric", err)
	}
	expectedFiles := []string{"lib/a.ric", "main.ric"}
	for i, statement := range program {
		position := statement.(*common.PrintStatement).Expression.(*common.VariableExpression).NameToken.Position
		if position.File != expectedFiles[i] {
			t.Errorf("statement %d %v is in file %q; want %q", i, statement, position.File, expectedFiles[i])
		}
	}
}

func TestLoadCycles(t *testing.T) {
	runLoadErrorTestCases(t, []loaderErrorTestCase{
		{
			"main file that imports itself",
			"main.ric",
			sourceFiles{"main.ric": `import "main.ric";`},
			"[main.ric, line 1, column 8] Import cycle: main.ric -> main.ric",
		},
		{
			"module that imports itself",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric";`,
				"a.ric":    `import "a.ric";`,
			},
			"[a.ric, line 1, column 8] Import cycle: a.ric -> a.ric",
		},
		{
			"cycle back to the main file",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric";`,
				"a.ric":    `import "main.ric";`,
			},
			"[a.ric, line 1, column 8] Import cycle: main.ric -> a.ric -> main.ric",
		},
		{
			"cycle of two modules",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric";`,
				"a.ric":    `import "b.ric";`,
				"b.ric":    `import "a.ric";`,
			},
			"[b.ric, line 1, column 8] Import cycle: a.ric -> b.ric -> a.ric",
		},
		{
			"cycle of three modules",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric";`,
				"a.ric":    `import "b.ric";`,
				"b.ric":    `import "c.ric";`,
				"c.ric":    "\n  import \"a.ric\";",
			},
			"[c.ric, line 2, column 10] Import cycle: a.ric -> b.ric -> c.ric -> a.ric",
		},
		{
			"cycle through relative paths",
			"main.ric",
			sourceFiles{
				"main.ric":  `import "lib/a.ric";`,
				"lib/a.ric": `import "../main.ric";`,
			},
			"[lib/a.ric, line 1, column 8] Import cycle: main.ric -> lib/a.ric -> main.ric",
		},
		{
			"cycle after a module already loaded",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric"; import "b.ric";`,
				"a.ric":    "print a;",
				"b.ric":    `import "a.ric"; import "b.ric";`,
			},
			"[b.ric, line 1, column 24] Import cycle: b.ric -> b.ric",
		},
	})
}

func TestLoadErrors(t *testing.T) {
	runLoadErrorTestCases(t, []loaderErrorTestCase{
		{
			"missing file",
			"main.ric",
			sourceFiles{"main.ric": `import "a.ric";`},
			`[main.ric, line 1, column 8] Cannot import "a.ric": file not found`,
		},
		{
			"missing file in a module",
			"main.ric",
			sourceFiles{
				"main.ric":  `import "lib/a.ric";`,
				"lib/a.ric": `import "b.ric";`,
			},
			`[lib/a.ric, line 1, column 8] Cannot import "b.ric": file not found`,
		},
		{
			"scanner error in a module",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric";`,
				"a.ric":    "print ?;",
			},
			"[a.ric, line 1, column 7] Non-recognizable character '?'",
		},
		{
			"parser error in a module",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric";`,
				"a.ric":    "print a",
			},
			"[a.ric, line 1, column 8] Expected ';' after expression",
		},
		{
			"misplaced import in a module",
			"main.ric",
			sourceFiles{
				"main.ric": `import "a.ric";`,
				"a.ric":    "print a;\n" + `import "b.ric";`,
				"b.ric":    "print b;",
			},
			"[a.ric, line 2, column 1] Imports must be at the beginning of the file",
		},
		{
			"error in a module of a subdirectory",
			"main.ric",
			sourceFiles{
				"main.ric":  `import "lib/a.ric";`,
				"lib/a.ric": `import "b.ric";`,
				"lib/b.ric": "print b",
			},
			"[lib/b.ric, line 1, column 8] Expected ';' after expression",
		},
	})
}

func TestLoadReadError(t *testing.T) {
	files := sourceFiles{"main.ric": `import "lib";`}
	readFile := func(path string) ([]byte, error) {
		return nil, &fs.PathError{Op: "read", Path: path, Err: errors.New("is a directory")}
	}

	assertLoadError(t, "main.ric", files, readFile, `[main.ric, line 1, column 8] Cannot import "lib": is a directory`)
}

func TestLoadWithOsReadFile(t *testing.T) {
	directory := t.TempDir()
	mainPath := filepath.Join(directory, "main.ric")
	writeFile := func(path string, content string) {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("os.WriteFile(%q) unexpected error: %v", path, err)
		}
	}
	writeFile(mainPath, `import "a.ric"; print main;`)
	writeFile(filepath.Join(directory, "a.ric"), "print a;")
	files := sourceFiles{mainPath: `import "a.ric"; print main;`}

	program, err := load(t, mainPath, files, os.ReadFile)

	if err != nil {
		t.Fatalf("loader.Load(%q) unexpected error: %v", mainPath, err)
	}
	if result := programString(program); result != "print a;\nprint main;\n" {
		t.Errorf("loader.Load(%q) = %q; want %q", mainPath, result, "print a;\nprint main;\n")
	}
}
