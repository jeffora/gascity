package formula

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// PackRootIntrinsic is the reserved runtime intrinsic that resolves to the
	// nearest pack root for a formula source path.
	PackRootIntrinsic = "pack_root"
)

// ResolveSourcePath returns a symlink-resolved absolute path.
// If symlink resolution fails, it returns a contextual error instead of an
// unresolved path.
func ResolveSourcePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve formula source path %q: %w", path, err)
	}
	return resolved, nil
}

// PackRootForFormulaSource derives the pack root from a resolved formula source
// path by walking upward to the nearest ancestor directory named "formulas".
// It returns false when the path has no formulas ancestor.
func PackRootForFormulaSource(sourcePath string) (string, bool) {
	sourcePath = strings.TrimSpace(sourcePath)
	if sourcePath == "" {
		return "", false
	}

	dir := filepath.Dir(sourcePath)
	for {
		if filepath.Base(dir) == "formulas" {
			return filepath.Dir(dir), true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// PackRootFromSourcePath derives the pack root from a resolved formula source
// path by walking upward to the nearest ancestor directory named "formulas".
// If no such ancestor exists, it returns an empty string.
func PackRootFromSourcePath(sourcePath string) string {
	packRoot, ok := PackRootForFormulaSource(sourcePath)
	if !ok {
		return ""
	}
	return packRoot
}
