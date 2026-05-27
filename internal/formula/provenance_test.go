package formula

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackRootForFormulaSource(t *testing.T) {
	dir := t.TempDir()
	packRoot := filepath.Join(dir, "packs", "demo")
	sourcePath := filepath.Join(packRoot, "formulas", "nested", "work.toml")

	got, ok := PackRootForFormulaSource(sourcePath)
	if !ok {
		t.Fatal("PackRootForFormulaSource ok = false, want true")
	}
	if got != packRoot {
		t.Fatalf("PackRootForFormulaSource root = %q, want %q", got, packRoot)
	}
}

func TestPackRootForFormulaSourceReportsFalseWithoutFormulasAncestor(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "not-formulas", "work.toml")

	got, ok := PackRootForFormulaSource(sourcePath)
	if ok {
		t.Fatal("PackRootForFormulaSource ok = true, want false")
	}
	if got != "" {
		t.Fatalf("PackRootForFormulaSource root = %q, want empty", got)
	}
}

func TestCompileExpandedTemplateStepCarriesTemplatePackRoot(t *testing.T) {
	dir := t.TempDir()
	packRoot := filepath.Join(dir, "pack")
	formulasDir := filepath.Join(packRoot, "formulas")
	if err := os.MkdirAll(formulasDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(formulas): %v", err)
	}

	expansionPath := filepath.Join(formulasDir, "wrap.toml")
	writeFormulaFile(t, expansionPath, `
formula = "wrap"
version = 1
type = "expansion"

[[template]]
id = "{target}.do"
title = "{{pack_root}}"
`)
	writeFormulaFile(t, filepath.Join(formulasDir, "work.toml"), `
formula = "work"
version = 1
type = "workflow"

[[steps]]
id = "build"
title = "Build"
expand = "wrap"
`)

	recipe, err := Compile(context.Background(), "work", []string{formulasDir}, nil)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	step := recipe.StepByID("work.build.do")
	if step == nil {
		t.Fatal("compiled step work.build.do not found")
	}
	if got := step.SourcePath; got != expansionPath {
		t.Fatalf("RecipeStep SourcePath = %q, want %q", got, expansionPath)
	}
	if got := step.PackRoot; got != packRoot {
		t.Fatalf("RecipeStep PackRoot = %q, want %q", got, packRoot)
	}
}

func TestCompileRejectsPackRootInUnsupportedCompileTimeFields(t *testing.T) {
	for name, body := range map[string]string{
		"condition": `
formula = "work"
version = 1
type = "workflow"

[[steps]]
id = "step"
title = "Step"
condition = "{{pack_root}}"
`,
		"loop-range": `
formula = "work"
version = 1
type = "workflow"

[[steps]]
id = "loop"
title = "Loop"

[steps.loop]
range = "{{pack_root}}"

[[steps.loop.body]]
id = "step"
title = "Step"
`,
		"step-expand": `
formula = "work"
version = 1
type = "workflow"

[[steps]]
id = "step"
title = "Step"
expand = "{{pack_root}}"
`,
		"compose-expand-target": `
formula = "work"
version = 1
type = "workflow"

[[steps]]
id = "step"
title = "Step"

[compose]
expand = [{ target = "{{pack_root}}", with = "wrap" }]
`,
		"compose-map-select": `
formula = "work"
version = 1
type = "workflow"

[[steps]]
id = "step"
title = "Step"

[compose]
map = [{ select = "{{pack_root}}", with = "wrap" }]
`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeFormulaFile(t, filepath.Join(dir, "work.toml"), body)

			_, err := Compile(context.Background(), "work", []string{dir}, nil)
			if err == nil {
				t.Fatal("Compile succeeded, want pack_root compile-time validation error")
			}
			if !strings.Contains(err.Error(), PackRootIntrinsic) {
				t.Fatalf("Compile error = %v, want pack_root mention", err)
			}
			if !strings.Contains(err.Error(), "not available in compile-time formula locations") {
				t.Fatalf("Compile error = %v, want unsupported-location message", err)
			}
		})
	}
}
