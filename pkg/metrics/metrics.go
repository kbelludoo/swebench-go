package metrics

import (
	"fmt"
	"time"
)

// RunResult holds execution metrics for a benchmark mode.
type RunResult struct {
	Mode           string        `json:"mode"`
	TaskID         string        `json:"task_id"`
	TokensConsumed int           `json:"tokens_consumed"`
	Duration       time.Duration `json:"duration"`
	CompilationOK  bool          `json:"compilation_ok"`
	TestsPassed    bool          `json:"tests_passed"`
	RaceDetected   bool          `json:"race_detected"`
	CostUSD        float64       `json:"cost_usd"`
}

// Comparison summarizes results between Standard and MB-VM First modes.
type Comparison struct {
	TaskID         string  `json:"task_id"`
	Title          string  `json:"title"`
	StandardTokens int     `json:"standard_tokens"`
	MBVMTokens     int     `json:"mbvm_tokens"`
	TokenSavings   float64 `json:"token_savings_percent"`
	StandardTimeMs int64   `json:"standard_time_ms"`
	MBVMTimeMs     int64   `json:"mbvm_time_ms"`
	SpeedupFactor  float64 `json:"speedup_factor"`
	CostSavings    float64 `json:"cost_savings_percent"`
	StandardPassed bool    `json:"standard_passed"`
	MBVMPassed     bool    `json:"mbvm_passed"`
	FailureReason  string  `json:"failure_reason"`
}

// ComputeCost estimates prompt cost assuming typical frontier model pricing ($3.00/1M prompt tokens).
func ComputeCost(tokens int) float64 {
	return (float64(tokens) / 1000000.0) * 3.00
}

// Compare computes the ratio and savings.
func Compare(std, mbvm RunResult, failureReason, title string) Comparison {
	savings := 0.0
	if std.TokensConsumed > 0 {
		savings = (1.0 - (float64(mbvm.TokensConsumed) / float64(std.TokensConsumed))) * 100.0
	}

	speedup := 1.0
	if mbvm.Duration.Milliseconds() > 0 {
		speedup = float64(std.Duration.Milliseconds()) / float64(mbvm.Duration.Milliseconds())
	}

	costSavings := 0.0
	if std.CostUSD > 0 {
		costSavings = (1.0 - (mbvm.CostUSD / std.CostUSD)) * 100.0
	}

	return Comparison{
		TaskID:         std.TaskID,
		Title:          title,
		StandardTokens: std.TokensConsumed,
		MBVMTokens:     mbvm.TokensConsumed,
		TokenSavings:   savings,
		StandardTimeMs: std.Duration.Milliseconds(),
		MBVMTimeMs:     mbvm.Duration.Milliseconds(),
		SpeedupFactor:  speedup,
		CostSavings:    costSavings,
		StandardPassed: std.TestsPassed,
		MBVMPassed:     mbvm.TestsPassed,
		FailureReason:  failureReason,
	}
}

// FormatTable returns an ANSI-formatted summary string.
func (c Comparison) FormatTable() string {
	passStd := "PASS"
	if !c.StandardPassed {
		passStd = "FAIL"
	}
	passMBVM := "PASS"
	if !c.MBVMPassed {
		passMBVM = "FAIL"
	}

	return fmt.Sprintf("| %-10s | %8d tok | %8d tok | %6.1f%% | %6d ms | %6d ms | %5.1fx | %4s | %4s |",
		c.TaskID,
		c.StandardTokens,
		c.MBVMTokens,
		c.TokenSavings,
		c.StandardTimeMs,
		c.MBVMTimeMs,
		c.SpeedupFactor,
		passStd,
		passMBVM,
	)
}
