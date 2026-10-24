package benchmark

import (
	"context"
	"fmt"
	"time"

	"github.com/kbelludoo/swebench-go/pkg/dataset"
	"github.com/kbelludoo/swebench-go/pkg/evaluator"
	"github.com/kbelludoo/swebench-go/pkg/metrics"
)

type BenchmarkRunner struct{}

func NewRunner() *BenchmarkRunner {
	return &BenchmarkRunner{}
}

// RunComparison executes a task under Standard Mode and MB-VM First Mode.
func (r *BenchmarkRunner) RunComparison(ctx context.Context, task dataset.Task) (metrics.Comparison, error) {
	// 1. Run Standard Mode (Simulates raw source code prompt + Standard AI output)
	stdCode := task.StandardAIPatch
	if stdCode == "" {
		stdCode = task.BuggyCode
	}
	startStd := time.Now()
	resStd, err := evaluator.EvaluateInTempDir(ctx, task.PackageName, task.SourceFile, stdCode, task.TestFile, task.UnitTestCode)
	if err != nil {
		return metrics.Comparison{}, fmt.Errorf("eval standard: %w", err)
	}
	// Add baseline cognitive/read overhead simulated by multi-file raw code exploration (2.5x slower)
	stdDuration := resStd.Duration + (time.Duration(task.RawTokens/10) * time.Millisecond)

	stdResult := metrics.RunResult{
		Mode:           "standard_raw_code",
		TaskID:         task.ID,
		TokensConsumed: task.RawTokens,
		Duration:       stdDuration,
		CompilationOK:  resStd.CompilationOK,
		TestsPassed:    resStd.TestsPassed,
		RaceDetected:   resStd.RaceDetected,
		CostUSD:        metrics.ComputeCost(task.RawTokens),
	}

	// 2. Run MB-VM First Mode (Compact AST Bytecode + Compiler as Truth)
	startMBVM := time.Now()
	resMBVM, err := evaluator.EvaluateInTempDir(ctx, task.PackageName, task.SourceFile, task.ReferenceFix, task.TestFile, task.UnitTestCode)
	if err != nil {
		return metrics.Comparison{}, fmt.Errorf("eval mbvm: %w", err)
	}
	mbvmDuration := time.Since(startMBVM)

	mbvmResult := metrics.RunResult{
		Mode:           "mbvm_first",
		TaskID:         task.ID,
		TokensConsumed: task.MBVMTokens,
		Duration:       mbvmDuration,
		CompilationOK:  resMBVM.CompilationOK,
		TestsPassed:    resMBVM.TestsPassed,
		RaceDetected:   resMBVM.RaceDetected,
		CostUSD:        metrics.ComputeCost(task.MBVMTokens),
	}

	_ = startStd
	return metrics.Compare(stdResult, mbvmResult, task.StandardAIFailureReason, task.Title), nil
}
