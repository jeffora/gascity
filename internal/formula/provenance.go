package formula

import (
	"path/filepath"
	"strings"
)

const (
	// PackRootIntrinsic is the reserved runtime intrinsic that resolves to the
	// nearest pack root for a formula source path.
	PackRootIntrinsic = "pack_root"
)

// ResolveSourcePath returns a symlink-resolved absolute path when possible.
// If symlink resolution fails, the original path is returned.
func ResolveSourcePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
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
