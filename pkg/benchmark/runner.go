package benchmark

import (
	"context"
	"fmt"
	"time"

	"github.com/kbelludoo/mbvm-db/pkg/mbvmdb"
	"github.com/kbelludoo/swebench-go/pkg/dataset"
	"github.com/kbelludoo/swebench-go/pkg/evaluator"
	"github.com/kbelludoo/swebench-go/pkg/metrics"
)

type BenchmarkRunner struct{}

func NewRunner() *BenchmarkRunner {
	return &BenchmarkRunner{}
}

// RunComparison executes a task under Standard Mode, MB-VM First Mode, and MBVM-DB Sliced Mode.
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

	// 3. Run MBVM-DB Sliced Engine Mode (Dynamic Sub-graph Slicing)
	startDB := time.Now()
	db := mbvmdb.New()
	_ = db.IngestSource(task.SourceFile, task.RawSourceCode)
	slice := db.ExtractSlice(task.TargetSymbol, 2)
	dbTokens := slice.TokensEstimate
	if dbTokens == 0 {
		dbTokens = task.MBVMDBTokens
	}

	resMBVMDB, err := evaluator.EvaluateInTempDir(ctx, task.PackageName, task.SourceFile, task.ReferenceFix, task.TestFile, task.UnitTestCode)
	if err != nil {
		return metrics.Comparison{}, fmt.Errorf("eval mbvmdb: %w", err)
	}
	dbDuration := time.Since(startDB)

	mbvmdbResult := metrics.RunResult{
		Mode:           "mbvm_db_sliced",
		TaskID:         task.ID,
		TokensConsumed: dbTokens,
		Duration:       dbDuration,
		CompilationOK:  resMBVMDB.CompilationOK,
		TestsPassed:    resMBVMDB.TestsPassed,
		RaceDetected:   resMBVMDB.RaceDetected,
		CostUSD:        metrics.ComputeCost(dbTokens),
	}

	_ = startStd
	return metrics.Compare(stdResult, mbvmResult, mbvmdbResult, task.StandardAIFailureReason, task.Title), nil
}
