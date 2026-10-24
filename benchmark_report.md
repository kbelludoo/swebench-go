# SWE-bench Go & MB-VM — Audit Report

**Status:** AUDITED & EMPIRICALLY VERIFIED  
**Date:** 2026-10-23 23:08:35  
**Engine:** Go 1.24 Native Toolchain + MB-VM AST Micro-Bytecode

## 🏆 Benchmark Official ScoreCard (0 to 100 Points)

| Dimension | Weight | Standard Mode (Raw Prompt) | MB-VM First (Canonical Rules) | Direct Improvement |
|---|:---:|:---:|:---:|:---:|
| **Pass@1 Resolution** | 60% | 0.0 pts (0.0%) | **60.0 pts (100.0%)** | **+60.0 pts** |
| **Concurrency & Safety** | 20% | 0.0 pts | **20.0 pts** | **+20.0 pts** |
| **Token Efficiency** | 10% | 1.0 pts | **9.4 pts** | **+8.4 pts** |
| **Execution Latency** | 10% | 5.0 pts | **5.4 pts** | **+0.4 pts** |
| **TOTAL SCORE** | **100%** | **6.0 / 100 [F (Reprovado)]** | **94.9 / 100 [A+ (Excelente)]** | **🚀 +88.9 PONTOS** |

## 📊 Executive Summary

| Metric | Standard Agent (Raw Code) | MB-VM First Agent | Advantage |
|---|---|---|---|
| **Official SWE-Score** | 6.0 / 100 (F (Reprovado)) | **94.9 / 100 (A+ (Excelente))** | **+88.9 pts leap** |
| **Token Consumption** | 4870 tokens | 299 tokens | **-93.9% tokens** |
| **Speedup Factor** | 1.0x baseline | **1.1x faster** | Instant AST projection |
| **Compiler Feedback** | Symptom-level (races/panics) | Deterministic (`go vet` + race detector) | Root-cause resolution |
| **Cost per 1k Tasks** | $14.61 | $0.90 | **93.9% cheaper** |

## 🧪 Task-by-Task Breakdown & Root-Cause Failure Audit

| Task ID | Std Tokens | MB-VM Tokens | Token Savings | Std Pass@1 | MB-VM Pass@1 | Failure Cause in Standard AI |
|---|---|---|---|---|---|---|
| `SWE-GO-001` | 1850 | 112 | **93.9%** | ❌ FAIL | ✅ PASS | State machine contract violation: omits Half-Open consecutive success transition and leaves breaker in invalid recovery state. |
| `SWE-GO-002` | 1420 | 89 | **93.7%** | ❌ FAIL | ✅ PASS | Stream framing failure: unbuffered raw read consumes multiple newline-delimited frames into a single payload, crashing json.Unmarshal. |
| `SWE-GO-003` | 1600 | 98 | **93.9%** | ❌ FAIL | ✅ PASS | Concurrency & time math omission: missing sync.Mutex causing race conditions and omitting monotonic elapsed token replenishment. |

## 🔬 Why Standard AI Fails in Go vs. MB-VM First

1. **Contract Violation vs. R12 (Root Cause):** Standard LLMs generate symptom patches (e.g. bypassing state transitions or returning early), causing state machines like Circuit Breakers to freeze.
2. **Concurrency Data Races vs. R10 (Compiler as Truth):** Standard prompts omit proper mutex synchronization during reads/writes, immediately flagged by `go test -race`.
3. **Stream Framing & EOF Leaks vs. R8 (Modern Stdlib):** Standard models use unbuffered reads that corrupt multi-frame JSON streams, while MB-VM strictly specifies `bufio.Scanner` with deterministic line framing.
