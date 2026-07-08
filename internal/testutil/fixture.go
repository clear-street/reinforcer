package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

// minGoDirective mirrors reinforcer's own go.mod floor (go 1.25) so the
// fixture module never claims a lower or higher minimum than the real one.
const minGoDirective = "go 1.25\n"

func WriteModule(t *testing.T, modulePath string, files map[string]string) *packages.Config {
	t.Helper()
	dir := t.TempDir()

	goMod := fmt.Sprintf("module %s\n\n%s", modulePath, minGoDirective)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}

	return &packages.Config{Dir: dir}
}
