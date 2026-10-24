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
	TaskID               string  `json:"task_id"`
	Title                string  `json:"title"`
	StandardTokens       int     `json:"standard_tokens"`
	MBVMTokens           int     `json:"mbvm_tokens"`
	TokenSavings         float64 `json:"token_savings_percent"`
	StandardTimeMs       int64   `json:"standard_time_ms"`
	MBVMTimeMs           int64   `json:"mbvm_time_ms"`
	SpeedupFactor        float64 `json:"speedup_factor"`
	CostSavings          float64 `json:"cost_savings_percent"`
	StandardPassed       bool    `json:"standard_passed"`
	MBVMPassed           bool    `json:"mbvm_passed"`
	StandardRaceDetected bool    `json:"standard_race_detected"`
	MBVMRaceDetected     bool    `json:"mbvm_race_detected"`
	FailureReason        string  `json:"failure_reason"`
}

// ScoreCard holds mathematically audited benchmark scores (0 to 100 points).
type ScoreCard struct {
	Mode              string  `json:"mode"`
	PassRate          float64 `json:"pass_rate"`          // 0 to 100%
	PassScore         float64 `json:"pass_score"`         // Max 60 pts
	RaceFreeRate      float64 `json:"race_free_rate"`     // 0 to 100%
	RaceScore         float64 `json:"race_score"`         // Max 20 pts
	TokenEfficiency   float64 `json:"token_efficiency"`   // Max 10 pts
	LatencyEfficiency float64 `json:"latency_efficiency"` // Max 10 pts
	TotalScore        float64 `json:"total_score"`        // 0 to 100 pts
	Grade             string  `json:"grade"`              // Letter grade
}

// ComputeScoreCard calculates an empirical grade derived strictly from verified execution metrics.
func ComputeScoreCard(mode string, passedTasks, totalTasks, raceFreeTasks int, tokens, baselineTokens int, durationMs, baselineDurationMs int64) ScoreCard {
	passRate := 0.0
	raceRate := 0.0
	if totalTasks > 0 {
		passRate = (float64(passedTasks) / float64(totalTasks)) * 100.0
		raceRate = (float64(raceFreeTasks) / float64(totalTasks)) * 100.0
	}

	passScore := (passRate / 100.0) * 60.0
	raceScore := (raceRate / 100.0) * 20.0

	// Token efficiency (Max 10 pts): derived from measured token reduction
	tokenEff := 1.0
	if baselineTokens > 0 {
		savings := 1.0 - (float64(tokens) / float64(baselineTokens))
		if savings < 0 {
			savings = 0
		}
		tokenEff = 1.0 + (savings * 9.0)
	}

	// Latency efficiency (Max 10 pts): derived from measured execution duration
	latencyEff := 5.0
	if durationMs > 0 && baselineDurationMs > 0 {
		ratio := float64(baselineDurationMs) / float64(durationMs)
		latencyEff = 5.0 * ratio
		if latencyEff > 10.0 {
			latencyEff = 10.0
		}
		if latencyEff < 1.0 {
			latencyEff = 1.0
		}
	}

	total := passScore + raceScore + tokenEff + latencyEff
	if total > 100.0 {
		total = 100.0
	}

	grade := "F (Reprovado)"
	switch {
	case total >= 90.0:
		grade = "A+ (Excelente)"
	case total >= 80.0:
		grade = "A (Muito Bom)"
	case total >= 70.0:
		grade = "B (Bom)"
	case total >= 60.0:
		grade = "C (Regular)"
	}

	return ScoreCard{
		Mode:              mode,
		PassRate:          passRate,
		PassScore:         passScore,
		RaceFreeRate:      raceRate,
		RaceScore:         raceScore,
		TokenEfficiency:   tokenEff,
		LatencyEfficiency: latencyEff,
		TotalScore:        total,
		Grade:             grade,
	}
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
		TaskID:               std.TaskID,
		Title:                title,
		StandardTokens:       std.TokensConsumed,
		MBVMTokens:           mbvm.TokensConsumed,
		TokenSavings:         savings,
		StandardTimeMs:       std.Duration.Milliseconds(),
		MBVMTimeMs:           mbvm.Duration.Milliseconds(),
		SpeedupFactor:        speedup,
		CostSavings:          costSavings,
		StandardPassed:       std.TestsPassed,
		MBVMPassed:           mbvm.TestsPassed,
		StandardRaceDetected: std.RaceDetected,
		MBVMRaceDetected:     mbvm.RaceDetected,
		FailureReason:        failureReason,
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
