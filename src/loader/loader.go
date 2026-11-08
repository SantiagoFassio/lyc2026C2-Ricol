package loader

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/SantiagoFassio/lyc2026C2-Ricol/common"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/parser"
	"github.com/SantiagoFassio/lyc2026C2-Ricol/scanner"
)

// Un módulo se muestra en los errores con su ruta, y se identifica con su ruta absoluta.
type module struct {
	path string
	key  string
}

type Loader struct {
	filePath   string
	statements []common.Statement
	readFile   func(path string) ([]byte, error)
	loaded     map[string]bool
	loading    []module
	modules    [][]common.Statement
}

func NewLoader(filePath string, statements []common.Statement, readFile func(path string) ([]byte, error)) *Loader {
	return &Loader{
		filePath:   filePath,
		statements: statements,
		readFile:   readFile,
		loaded:     map[string]bool{},
		loading:    []module{},
		modules:    [][]common.Statement{},
	}
}

func (l *Loader) Load() ([][]common.Statement, error) {
	if err := l.loadModule(l.filePath, l.statements); err != nil {
		return [][]common.Statement{}, err
	}
	return l.modules, nil
}

// Un módulo se agrega después de sus imports, así lo importado se declara antes de usarse.
func (l *Loader) loadModule(path string, statements []common.Statement) error {
	current := module{path: path, key: moduleKey(path)}
	l.loading = append(l.loading, current)
	moduleStatements := []common.Statement{}
	for _, statement := range statements {
		importStatement, ok := statement.(*common.ImportStatement)
		if !ok {
			moduleStatements = append(moduleStatements, statement)
			continue
		}
		if err := l.loadImport(path, importStatement); err != nil {
			return err
		}
	}
	l.loading = l.loading[:len(l.loading)-1]
	l.loaded[current.key] = true
	l.modules = append(l.modules, moduleStatements)
	return nil
}

func (l *Loader) loadImport(importerPath string, importStatement *common.ImportStatement) error {
	path := resolvePath(importerPath, importStatement.Path)
	key := moduleKey(path)
	pathToken := importStatement.PathToken
	if cycle, found := l.cycleTo(key); found {
		return common.NewRicolError(pathToken.Position,
			fmt.Sprintf("Import cycle: %s", strings.Join(append(cycle, path), " -> ")))
	}
	if l.loaded[key] {
		return nil
	}
	content, err := l.readFile(path)
	if err != nil {
		return common.NewRicolError(pathToken.Position, fmt.Sprintf("Cannot import %s: %s", pathToken.Lexeme, readErrorMessage(err)))
	}
	tokens, err := scanner.NewScanner(string(content), path).Scan()
	if err != nil {
		return err
	}
	statements, err := parser.NewParser(tokens).Parse()
	if err != nil {
		return err
	}
	return l.loadModule(path, statements)
}

// Si el módulo se está cargando, devuelve las rutas desde él hasta el módulo actual.
func (l *Loader) cycleTo(key string) ([]string, bool) {
	for i, loadingModule := range l.loading {
		if loadingModule.key != key {
			continue
		}
		paths := []string{}
		for _, cycleModule := range l.loading[i:] {
			paths = append(paths, cycleModule.path)
		}
		return paths, true
	}
	return nil, false
}

func resolvePath(importerPath string, importPath string) string {
	if filepath.IsAbs(importPath) {
		return filepath.Clean(importPath)
	}
	return filepath.Join(filepath.Dir(importerPath), importPath)
}

func moduleKey(path string) string {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return absolutePath
}

func readErrorMessage(err error) string {
	if errors.Is(err, fs.ErrNotExist) {
		return "file not found"
	}
	var pathError *fs.PathError
	if errors.As(err, &pathError) {
		return pathError.Err.Error()
	}
	return err.Error()
}
