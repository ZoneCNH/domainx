package domain

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageImportBoundary(t *testing.T) {
	allowedPrefixes := []string{"crypto/", "encoding/", "errors", "fmt", "strings", "time", "github.com/ZoneCNH/decimalx/pkg/decimalx"}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			imp := strings.Trim(spec.Path.Value, `"`)
			ok := false
			for _, prefix := range allowedPrefixes {
				if imp == prefix || strings.HasPrefix(imp, prefix) {
					ok = true
					break
				}
			}
			if !ok {
				t.Fatalf("forbidden import %q in %s", imp, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
