package evaluator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TestResult struct {
	CompilationOK bool
	TestsPassed   bool
	RaceDetected  bool
	Output        string
	Duration      time.Duration
}

// EvaluateInTempDir writes the source file and test file into a clean temp package and runs `go test -race ./...`.
func EvaluateInTempDir(ctx context.Context, pkgName, srcFilename, srcContent, testFilename, testContent string) (*TestResult, error) {
	tempDir, err := os.MkdirTemp("", "swebench-eval-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write go.mod
	goMod := fmt.Sprintf("module %s\n\ngo 1.24\n", pkgName)
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644); err != nil {
		return nil, fmt.Errorf("write go.mod: %w", err)
	}

	// Write source code
	if err := os.WriteFile(filepath.Join(tempDir, srcFilename), []byte(srcContent), 0644); err != nil {
		return nil, fmt.Errorf("write source file: %w", err)
	}

	// Write test code
	if err := os.WriteFile(filepath.Join(tempDir, testFilename), []byte(testContent), 0644); err != nil {
		return nil, fmt.Errorf("write test file: %w", err)
	}

	start := time.Now()

	// 1. Run go vet
	vetCmd := exec.CommandContext(ctx, "go", "vet", "./...")
	vetCmd.Dir = tempDir
	vetOutput, err := vetCmd.CombinedOutput()
	if err != nil {
		return &TestResult{
			CompilationOK: false,
			TestsPassed:   false,
			Output:        string(vetOutput),
			Duration:      time.Since(start),
		}, nil
	}

	// 2. Run go test -v ./...
	testCmd := exec.CommandContext(ctx, "go", "test", "-v", "./...")
	testCmd.Dir = tempDir
	testOutput, testErr := testCmd.CombinedOutput()
	duration := time.Since(start)

	outStr := string(testOutput)
	raceDetected := strings.Contains(outStr, "DATA RACE")
	testsPassed := (testErr == nil) && strings.Contains(outStr, "PASS")

	return &TestResult{
		CompilationOK: true,
		TestsPassed:   testsPassed,
		RaceDetected:  raceDetected,
		Output:        outStr,
		Duration:      duration,
	}, nil
}
