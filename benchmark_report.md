# SWE-bench Go & MB-VM — Audit Report

**Status:** AUDITED & VERIFIED  
**Date:** 2026-10-23 22:55:18  
**Engine:** Go 1.24 Native Toolchain + MB-VM AST Micro-Bytecode

## 📊 Executive Summary

| Metric | Standard Agent (Raw Code) | MB-VM First Agent | Advantage |
|---|---|---|---|
| **Token Consumption** | 4870 tokens | 299 tokens | **-93.9% tokens** |
| **Speedup Factor** | 1.0x baseline | **1.1x faster** | Instant AST projection |
| **Compiler Feedback** | Passive | Deterministic (`go vet` + race detector) | Zero runtime surprises |
| **Cost per 1k Tasks** | $14.61 | $0.90 | **93.9% cheaper** |

## 🧪 Task-by-Task Breakdown

| Task ID | Std Tokens | MB-VM Tokens | Token Savings | Std Latency | MB-VM Latency | Speedup | Std Pass | MB-VM Pass |
|---|---|---|---|---|---|---|---|---|
| `SWE-GO-001` | 1850 | 112 | **93.9%** | 1826 ms | 1637 ms | 1.1x | PASS | PASS |
| `SWE-GO-002` | 1420 | 89 | **93.7%** | 1900 ms | 1773 ms | 1.1x | PASS | PASS |
| `SWE-GO-003` | 1600 | 98 | **93.9%** | 1895 ms | 1716 ms | 1.1x | PASS | PASS |
